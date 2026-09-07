package ws

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/cifo-monitoring/backend/internal/integration"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// Streamer orchestrates live streams
type Streamer struct {
	hub          *Hub
	dockerClient integration.DockerClient
	k8sClient    integration.KubernetesClient
	logger       *slog.Logger

	activeStreams   map[string]context.CancelFunc
	activeStreamsMu sync.Mutex
}

// NewStreamer creates streamer instance
func NewStreamer(hub *Hub, dockerClient integration.DockerClient, k8sClient integration.KubernetesClient, logger *slog.Logger) *Streamer {
	s := &Streamer{
		hub:           hub,
		dockerClient:  dockerClient,
		k8sClient:     k8sClient,
		logger:        logger,
		activeStreams: make(map[string]context.CancelFunc),
	}

	// register topic hooks
	hub.SetCallbacks(s.handleTopicSubscribed, s.handleTopicEmpty)
	return s
}

// Start initiates background listeners
func (s *Streamer) Start(ctx context.Context) {
	if s.dockerClient != nil {
		go s.listenDockerEvents(ctx)
	}
	if s.k8sClient != nil {
		go s.watchK8sEvents(ctx)
	}
}

// SendNotification dispatches push alert
func (s *Streamer) SendNotification(payload NotificationPayload) {
	msg := NewWSMessage(TypeNotification, "notifications", payload)
	s.hub.Broadcast("notifications", msg)
}

// handleTopicSubscribed hooks new topics
func (s *Streamer) handleTopicSubscribed(topic string) {
	if strings.HasPrefix(topic, "docker_logs:") {
		containerID := strings.TrimPrefix(topic, "docker_logs:")
		s.startDockerLogStream(containerID)
	} else if strings.HasPrefix(topic, "k8s_logs:") {
		s.handleK8sLogSubscribed(topic)
	}
}

// handleTopicEmpty hooks empty topics
func (s *Streamer) handleTopicEmpty(topic string) {
	if strings.HasPrefix(topic, "docker_logs:") {
		containerID := strings.TrimPrefix(topic, "docker_logs:")
		s.stopDockerLogStream(containerID)
	} else if strings.HasPrefix(topic, "k8s_logs:") {
		s.stopK8sLogStream(topic)
	}
}

// startDockerLogStream starts log reader
func (s *Streamer) startDockerLogStream(containerID string) {
	if s.dockerClient == nil {
		return
	}

	s.activeStreamsMu.Lock()
	if _, exists := s.activeStreams[containerID]; exists {
		s.activeStreamsMu.Unlock()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.activeStreams[containerID] = cancel
	s.activeStreamsMu.Unlock()

	go func() {
		defer s.stopDockerLogStream(containerID)

		reader, err := s.dockerClient.StreamContainerLogs(ctx, containerID, 100, true)
		if err != nil {
			s.logger.Warn("docker log stream failed", slog.String("id", containerID), slog.String("error", err.Error()))
			return
		}
		defer reader.Close()

		topic := "docker_logs:" + containerID
		bufReader := bufio.NewReader(reader)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				line, err := bufReader.ReadString('\n')
				if err != nil {
					if err != io.EOF && ctx.Err() == nil {
						s.logger.Debug("log reader read error", slog.String("id", containerID), slog.String("error", err.Error()))
					}
					return
				}

				cleanLine := cleanDockerLogLine(line)
				if cleanLine == "" {
					continue
				}

				streamType := "stdout"
				if strings.Contains(strings.ToLower(cleanLine), "error") || strings.Contains(strings.ToLower(cleanLine), "fail") {
					streamType = "stderr"
				}

				payload := LogPayload{
					Source:    "docker",
					ID:        containerID,
					Stream:    streamType,
					Log:       cleanLine,
					Timestamp: time.Now().Format("15:04:05"),
				}

				s.hub.Broadcast(topic, NewWSMessage(TypeLogEntry, topic, payload))
			}
		}
	}()
}

// stopDockerLogStream terminates log reader
func (s *Streamer) stopDockerLogStream(containerID string) {
	s.activeStreamsMu.Lock()
	defer s.activeStreamsMu.Unlock()

	if cancel, exists := s.activeStreams[containerID]; exists {
		cancel()
		delete(s.activeStreams, containerID)
		s.logger.Debug("stopped docker log stream", slog.String("id", containerID))
	}
}

func (s *Streamer) handleK8sLogSubscribed(topic string) {
	raw := strings.TrimPrefix(topic, "k8s_logs:")
	parts := strings.Split(raw, ":")
	if len(parts) >= 2 {
		namespace := parts[0]
		pod := parts[1]
		container := ""
		if len(parts) >= 3 {
			container = parts[2]
		}
		s.startK8sLogStream(topic, namespace, pod, container)
	}
}

// startK8sLogStream starts pod log reader
func (s *Streamer) startK8sLogStream(topic string, namespace string, pod string, container string) {
	if s.k8sClient == nil {
		return
	}

	s.activeStreamsMu.Lock()
	if _, exists := s.activeStreams[topic]; exists {
		s.activeStreamsMu.Unlock()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.activeStreams[topic] = cancel
	s.activeStreamsMu.Unlock()

	go func() {
		defer s.stopK8sLogStream(topic)

		reader, err := s.k8sClient.GetPodLogs(ctx, namespace, pod, container, 100)
		if err != nil {
			s.logger.Warn("k8s log stream failed", slog.String("pod", pod), slog.String("error", err.Error()))
			return
		}
		defer reader.Close()

		bufReader := bufio.NewReader(reader)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				line, err := bufReader.ReadString('\n')
				if err != nil {
					return
				}
				cleanLine := strings.TrimSpace(line)
				if cleanLine == "" {
					continue
				}

				streamType := "stdout"
				lower := strings.ToLower(cleanLine)
				if strings.Contains(lower, "error") || strings.Contains(lower, "fail") {
					streamType = "stderr"
				}

				payload := LogPayload{
					Source:    "kubernetes",
					ID:        fmt.Sprintf("%s/%s", namespace, pod),
					Stream:    streamType,
					Log:       cleanLine,
					Timestamp: time.Now().Format("15:04:05"),
				}

				s.hub.Broadcast(topic, NewWSMessage(TypeLogEntry, topic, payload))
			}
		}
	}()
}

// stopK8sLogStream terminates pod log reader
func (s *Streamer) stopK8sLogStream(topic string) {
	s.activeStreamsMu.Lock()
	defer s.activeStreamsMu.Unlock()

	if cancel, exists := s.activeStreams[topic]; exists {
		cancel()
		delete(s.activeStreams, topic)
		s.logger.Debug("stopped k8s log stream", slog.String("topic", topic))
	}
}

// cleanDockerLogLine strips docker frame header
func cleanDockerLogLine(raw string) string {
	bytes := []byte(raw)
	// strip 8-byte multiplex header if present
	if len(bytes) >= 8 && (bytes[0] == 1 || bytes[0] == 2) && bytes[1] == 0 && bytes[2] == 0 && bytes[3] == 0 {
		bytes = bytes[8:]
	}
	return strings.TrimSpace(string(bytes))
}

// listenDockerEvents streams docker events
func (s *Streamer) listenDockerEvents(ctx context.Context) {
	msgChan, errChan := s.dockerClient.ListenEvents(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case err, ok := <-errChan:
			if !ok || err == nil {
				return
			}
			s.logger.Warn("docker events error", slog.String("error", err.Error()))
			time.Sleep(2 * time.Second)
			msgChan, errChan = s.dockerClient.ListenEvents(ctx)
		case event, ok := <-msgChan:
			if !ok {
				return
			}

			// filter container events
			if event.Type != "container" {
				continue
			}

			action := string(event.Action)
			name := event.Actor.Attributes["name"]
			if name == "" {
				name = event.Actor.ID[:12]
			}

			eventPayload := EventPayload{
				Type:      "docker",
				Resource:  name,
				Action:    action,
				Message:   fmt.Sprintf("Container %s %s", name, action),
				Timestamp: time.Now().Format("15:04:05"),
			}

			// broadcast to container_event and system_event
			s.hub.Broadcast("container_events", NewWSMessage(TypeContainerEvent, "container_events", eventPayload))
			s.hub.Broadcast("system_events", NewWSMessage(TypeSystemEvent, "system_events", eventPayload))

			// trigger notification for abnormal terminations
			if action == "die" || action == "oom" {
				s.SendNotification(NotificationPayload{
					ID:        fmt.Sprintf("docker-%s-%d", action, time.Now().UnixNano()),
					Title:     fmt.Sprintf("Container %s: %s", strings.ToUpper(action), name),
					Message:   fmt.Sprintf("Container %s exited with state %s", name, action),
					Severity:  "critical",
					Timestamp: time.Now().Format("15:04:05"),
					Source:    "docker",
				})
			}
		}
	}
}

// watchK8sEvents streams k8s events
func (s *Streamer) watchK8sEvents(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		watcher, err := s.k8sClient.WatchEvents(ctx, "")
		if err != nil {
			s.logger.Warn("k8s watch events failed", slog.String("error", err.Error()))
			time.Sleep(5 * time.Second)
			continue
		}

		for event := range watcher.ResultChan() {
			if event.Type == watch.Error {
				break
			}

			k8sEvt, ok := event.Object.(*corev1.Event)
			if !ok {
				continue
			}

			resource := fmt.Sprintf("%s/%s", k8sEvt.InvolvedObject.Kind, k8sEvt.InvolvedObject.Name)
			eventPayload := EventPayload{
				Type:      "kubernetes",
				Resource:  resource,
				Action:    k8sEvt.Reason,
				Reason:    k8sEvt.Reason,
				Message:   k8sEvt.Message,
				Timestamp: time.Now().Format("15:04:05"),
			}

			s.hub.Broadcast("k8s_events", NewWSMessage(TypeK8sEvent, "k8s_events", eventPayload))
			s.hub.Broadcast("system_events", NewWSMessage(TypeSystemEvent, "system_events", eventPayload))

			// send notification if warning/critical k8s event
			if k8sEvt.Type == corev1.EventTypeWarning {
				s.SendNotification(NotificationPayload{
					ID:        fmt.Sprintf("k8s-%d", time.Now().UnixNano()),
					Title:     fmt.Sprintf("K8s Warning: %s", k8sEvt.Reason),
					Message:   fmt.Sprintf("%s on %s: %s", k8sEvt.Reason, resource, k8sEvt.Message),
					Severity:  "warning",
					Timestamp: time.Now().Format("15:04:05"),
					Source:    "kubernetes",
				})
			}
		}

		watcher.Stop()
		time.Sleep(2 * time.Second)
	}
}
