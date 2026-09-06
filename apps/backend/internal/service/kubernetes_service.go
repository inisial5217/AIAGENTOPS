package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/cifo-monitoring/backend/internal/integration"
	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/repository"
	"github.com/cifo-monitoring/backend/pkg/apperror"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// KubernetesService business logic for kubernetes
type KubernetesService interface {
	ListPods(ctx context.Context, namespace string) ([]model.PodSummary, error)
	GetPod(ctx context.Context, namespace string, name string) (*model.PodDetail, error)
	GetPodLogs(ctx context.Context, namespace string, name string, container string, tailLines int64) (io.ReadCloser, error)
	ListDeployments(ctx context.Context, namespace string) ([]model.DeploymentSummary, error)
	GetDeployment(ctx context.Context, namespace string, name string) (*model.DeploymentDetail, error)
	RestartDeployment(ctx context.Context, namespace string, name string, actor string, ip string) error
	ScaleDeployment(ctx context.Context, namespace string, name string, replicas int32, actor string, ip string) error
	ListNodes(ctx context.Context) ([]model.NodeSummary, error)
	ListServices(ctx context.Context, namespace string) ([]model.ServiceSummary, error)
	GetClusterOverview(ctx context.Context) (*model.K8sClusterOverview, error)
}

type kubernetesServiceImpl struct {
	client    integration.KubernetesClient
	auditRepo repository.AuditRepository
	logger    *slog.Logger
}

// NewKubernetesService creates kubernetes service
func NewKubernetesService(
	client integration.KubernetesClient,
	auditRepo repository.AuditRepository,
	logger *slog.Logger,
) KubernetesService {
	return &kubernetesServiceImpl{
		client:    client,
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// ListPods returns pod summaries
func (s *kubernetesServiceImpl) ListPods(ctx context.Context, namespace string) ([]model.PodSummary, error) {
	podList, err := s.client.ListPods(ctx, namespace)
	if err != nil {
		return nil, err
	}

	result := make([]model.PodSummary, 0, len(podList.Items))
	for _, pod := range podList.Items {
		result = append(result, mapPodToSummary(&pod))
	}
	return result, nil
}

// GetPod returns pod detail
func (s *kubernetesServiceImpl) GetPod(ctx context.Context, namespace string, name string) (*model.PodDetail, error) {
	pod, err := s.client.GetPod(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	summary := mapPodToSummary(pod)
	containers := make([]model.ContainerInfo, 0, len(pod.Spec.Containers))

	statusMap := make(map[string]corev1.ContainerStatus)
	for _, cs := range pod.Status.ContainerStatuses {
		statusMap[cs.Name] = cs
	}

	for _, c := range pod.Spec.Containers {
		cInfo := model.ContainerInfo{
			Name:  c.Name,
			Image: c.Image,
			Ports: make([]string, 0),
		}
		for _, port := range c.Ports {
			proto := string(port.Protocol)
			if proto == "" {
				proto = "TCP"
			}
			cInfo.Ports = append(cInfo.Ports, fmt.Sprintf("%d/%s", port.ContainerPort, proto))
		}

		if cs, ok := statusMap[c.Name]; ok {
			cInfo.Ready = cs.Ready
			cInfo.RestartCount = cs.RestartCount
			cInfo.State = getContainerStateString(cs.State)
		} else {
			cInfo.State = "Waiting"
		}
		containers = append(containers, cInfo)
	}

	conditions := make([]string, 0, len(pod.Status.Conditions))
	for _, cond := range pod.Status.Conditions {
		if cond.Status == corev1.ConditionTrue {
			conditions = append(conditions, string(cond.Type))
		}
	}

	startTime := ""
	if pod.Status.StartTime != nil {
		startTime = pod.Status.StartTime.Time.Format(time.RFC3339)
	}

	detail := &model.PodDetail{
		PodSummary: summary,
		QoSClass:   string(pod.Status.QOSClass),
		StartTime:  startTime,
		Containers: containers,
		Conditions: conditions,
		NodeIP:     pod.Status.HostIP,
	}
	return detail, nil
}

// GetPodLogs streams pod logs
func (s *kubernetesServiceImpl) GetPodLogs(ctx context.Context, namespace string, name string, container string, tailLines int64) (io.ReadCloser, error) {
	return s.client.GetPodLogs(ctx, namespace, name, container, tailLines)
}

// ListDeployments returns deployment summaries
func (s *kubernetesServiceImpl) ListDeployments(ctx context.Context, namespace string) ([]model.DeploymentSummary, error) {
	deployList, err := s.client.ListDeployments(ctx, namespace)
	if err != nil {
		return nil, err
	}

	result := make([]model.DeploymentSummary, 0, len(deployList.Items))
	for _, d := range deployList.Items {
		result = append(result, mapDeploymentToSummary(&d))
	}
	return result, nil
}

// GetDeployment returns deployment detail
func (s *kubernetesServiceImpl) GetDeployment(ctx context.Context, namespace string, name string) (*model.DeploymentDetail, error) {
	d, err := s.client.GetDeployment(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	summary := mapDeploymentToSummary(d)
	conditions := make([]string, 0, len(d.Status.Conditions))
	for _, c := range d.Status.Conditions {
		if c.Status == corev1.ConditionTrue {
			conditions = append(conditions, string(c.Type))
		}
	}

	selector := make(map[string]string)
	if d.Spec.Selector != nil && d.Spec.Selector.MatchLabels != nil {
		selector = d.Spec.Selector.MatchLabels
	}

	return &model.DeploymentDetail{
		DeploymentSummary: summary,
		Strategy:          string(d.Spec.Strategy.Type),
		Selector:          selector,
		Conditions:        conditions,
	}, nil
}

// RestartDeployment triggers rollout restart
func (s *kubernetesServiceImpl) RestartDeployment(ctx context.Context, namespace string, name string, actor string, ip string) error {
	if err := s.client.RestartDeployment(ctx, namespace, name); err != nil {
		return err
	}

	resourceID := fmt.Sprintf("%s/%s", namespace, name)
	details, _ := json.Marshal(map[string]string{
		"namespace":  namespace,
		"deployment": name,
	})
	_ = s.auditRepo.Create(ctx, &model.AuditLog{
		Timestamp:    time.Now(),
		ActorType:    "user",
		ActorID:      actor,
		Action:       "restart_k8s_deployment",
		ResourceType: "kubernetes",
		ResourceID:   &resourceID,
		Details:      details,
		IPAddress:    &ip,
		Result:       "success",
	})
	return nil
}

// ScaleDeployment scales replicas count
func (s *kubernetesServiceImpl) ScaleDeployment(ctx context.Context, namespace string, name string, replicas int32, actor string, ip string) error {
	if replicas < 0 {
		return apperror.NewBadRequest("replicas cannot be negative")
	}
	if err := s.client.ScaleDeployment(ctx, namespace, name, replicas); err != nil {
		return err
	}

	resourceID := fmt.Sprintf("%s/%s", namespace, name)
	details, _ := json.Marshal(map[string]interface{}{
		"namespace":  namespace,
		"deployment": name,
		"replicas":   replicas,
	})
	_ = s.auditRepo.Create(ctx, &model.AuditLog{
		Timestamp:    time.Now(),
		ActorType:    "user",
		ActorID:      actor,
		Action:       "scale_k8s_deployment",
		ResourceType: "kubernetes",
		ResourceID:   &resourceID,
		Details:      details,
		IPAddress:    &ip,
		Result:       "success",
	})
	return nil
}

// ListNodes returns node summaries
func (s *kubernetesServiceImpl) ListNodes(ctx context.Context) ([]model.NodeSummary, error) {
	nodeList, err := s.client.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	podsByNode := make(map[string]int)
	if podList, pErr := s.client.ListPods(ctx, ""); pErr == nil {
		for _, pod := range podList.Items {
			if pod.Spec.NodeName != "" {
				podsByNode[pod.Spec.NodeName]++
			}
		}
	}

	result := make([]model.NodeSummary, 0, len(nodeList.Items))
	for _, n := range nodeList.Items {
		status := "NotReady"
		for _, cond := range n.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				status = "Ready"
				break
			}
		}

		roles := make([]string, 0)
		for label := range n.Labels {
			if strings.HasPrefix(label, "node-role.kubernetes.io/") {
				role := strings.TrimPrefix(label, "node-role.kubernetes.io/")
				if role != "" {
					roles = append(roles, role)
				}
			}
		}
		if len(roles) == 0 {
			roles = append(roles, "worker")
		}

		internalIP := ""
		for _, addr := range n.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				internalIP = addr.Address
				break
			}
		}

		memBytes := int64(0)
		if memQuantity, ok := n.Status.Capacity[corev1.ResourceMemory]; ok {
			memBytes = memQuantity.Value()
		}

		cpuCapacity := ""
		if cpuQuantity, ok := n.Status.Capacity[corev1.ResourceCPU]; ok {
			cpuCapacity = cpuQuantity.String()
		}

		summary := model.NodeSummary{
			Name:                n.Name,
			Status:              status,
			Roles:               strings.Join(roles, ", "),
			Version:             n.Status.NodeInfo.KubeletVersion,
			InternalIP:          internalIP,
			CPUCapacity:         cpuCapacity,
			MemoryCapacityBytes: memBytes,
			PodCount:            podsByNode[n.Name],
			OSImage:             n.Status.NodeInfo.OSImage,
			KernelVersion:       n.Status.NodeInfo.KernelVersion,
			ContainerRuntime:    n.Status.NodeInfo.ContainerRuntimeVersion,
		}
		result = append(result, summary)
	}
	return result, nil
}

// ListServices returns service summaries
func (s *kubernetesServiceImpl) ListServices(ctx context.Context, namespace string) ([]model.ServiceSummary, error) {
	svcList, err := s.client.ListServices(ctx, namespace)
	if err != nil {
		return nil, err
	}

	result := make([]model.ServiceSummary, 0, len(svcList.Items))
	for _, svc := range svcList.Items {
		ports := make([]model.ServicePortMapping, 0, len(svc.Spec.Ports))
		for _, p := range svc.Spec.Ports {
			ports = append(ports, model.ServicePortMapping{
				Name:       p.Name,
				Port:       p.Port,
				TargetPort: p.TargetPort.String(),
				Protocol:   string(p.Protocol),
				NodePort:   p.NodePort,
			})
		}

		externalIP := ""
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			if svc.Status.LoadBalancer.Ingress[0].IP != "" {
				externalIP = svc.Status.LoadBalancer.Ingress[0].IP
			} else {
				externalIP = svc.Status.LoadBalancer.Ingress[0].Hostname
			}
		} else if len(svc.Spec.ExternalIPs) > 0 {
			externalIP = svc.Spec.ExternalIPs[0]
		}

		summary := model.ServiceSummary{
			Name:       svc.Name,
			Namespace:  svc.Namespace,
			Type:       string(svc.Spec.Type),
			ClusterIP:  svc.Spec.ClusterIP,
			ExternalIP: externalIP,
			Ports:      ports,
			Selector:   svc.Spec.Selector,
			Age:        formatAge(svc.CreationTimestamp.Time),
			CreatedAt:  svc.CreationTimestamp.Time.Unix(),
		}
		result = append(result, summary)
	}
	return result, nil
}

// GetClusterOverview returns cluster overview
func (s *kubernetesServiceImpl) GetClusterOverview(ctx context.Context) (*model.K8sClusterOverview, error) {
	nodes, err := s.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	pods, err := s.ListPods(ctx, "")
	if err != nil {
		return nil, err
	}

	deployments, err := s.ListDeployments(ctx, "")
	if err != nil {
		return nil, err
	}

	services, err := s.ListServices(ctx, "")
	if err != nil {
		return nil, err
	}

	readyNodes := 0
	for _, n := range nodes {
		if n.Status == "Ready" {
			readyNodes++
		}
	}

	runningPods := 0
	for _, p := range pods {
		if p.Status == "Running" {
			runningPods++
		}
	}

	readyDeployments := 0
	for _, d := range deployments {
		if d.ReadyReplicas == d.Replicas && d.Replicas > 0 {
			readyDeployments++
		}
	}

	return &model.K8sClusterOverview{
		TotalNodes:       len(nodes),
		ReadyNodes:       readyNodes,
		TotalPods:        len(pods),
		RunningPods:      runningPods,
		TotalDeployments: len(deployments),
		ReadyDeployments: readyDeployments,
		TotalServices:    len(services),
	}, nil
}

func mapPodToSummary(pod *corev1.Pod) model.PodSummary {
	restarts := int32(0)
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}

	status := string(pod.Status.Phase)
	if pod.DeletionTimestamp != nil {
		status = "Terminating"
	} else {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
				status = cs.State.Waiting.Reason
				break
			}
			if cs.State.Terminated != nil && cs.State.Terminated.Reason != "" {
				status = cs.State.Terminated.Reason
				break
			}
		}
	}

	var reqCPU, reqMem, limCPU, limMem int64
	for _, c := range pod.Spec.Containers {
		if q, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
			reqCPU += q.MilliValue()
		}
		if q, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
			reqMem += q.Value()
		}
		if q, ok := c.Resources.Limits[corev1.ResourceCPU]; ok {
			limCPU += q.MilliValue()
		}
		if q, ok := c.Resources.Limits[corev1.ResourceMemory]; ok {
			limMem += q.Value()
		}
	}

	cpuReqStr := ""
	if reqCPU > 0 {
		cpuReqStr = fmt.Sprintf("%dm", reqCPU)
	}
	memReqStr := ""
	if reqMem > 0 {
		memReqStr = fmt.Sprintf("%dMi", reqMem/(1024*1024))
	}
	cpuLimStr := ""
	if limCPU > 0 {
		cpuLimStr = fmt.Sprintf("%dm", limCPU)
	}
	memLimStr := ""
	if limMem > 0 {
		memLimStr = fmt.Sprintf("%dMi", limMem/(1024*1024))
	}

	return model.PodSummary{
		Name:          pod.Name,
		Namespace:     pod.Namespace,
		Status:        status,
		Phase:         string(pod.Status.Phase),
		Restarts:      restarts,
		CPURequest:    cpuReqStr,
		MemoryRequest: memReqStr,
		CPULimit:      cpuLimStr,
		MemoryLimit:   memLimStr,
		Node:          pod.Spec.NodeName,
		IP:            pod.Status.PodIP,
		Age:           formatAge(pod.CreationTimestamp.Time),
		CreatedAt:     pod.CreationTimestamp.Time.Unix(),
		Labels:        pod.Labels,
	}
}

func mapDeploymentToSummary(d *appsv1.Deployment) model.DeploymentSummary {
	images := make([]string, 0, len(d.Spec.Template.Spec.Containers))
	for _, c := range d.Spec.Template.Spec.Containers {
		images = append(images, c.Image)
	}

	replicas := int32(1)
	if d.Spec.Replicas != nil {
		replicas = *d.Spec.Replicas
	}

	return model.DeploymentSummary{
		Name:              d.Name,
		Namespace:         d.Namespace,
		Replicas:          replicas,
		ReadyReplicas:     d.Status.ReadyReplicas,
		AvailableReplicas: d.Status.AvailableReplicas,
		UpdatedReplicas:   d.Status.UpdatedReplicas,
		Images:            images,
		Age:               formatAge(d.CreationTimestamp.Time),
		CreatedAt:         d.CreationTimestamp.Time.Unix(),
		Labels:            d.Labels,
	}
}

func getContainerStateString(state corev1.ContainerState) string {
	if state.Running != nil {
		return "Running"
	}
	if state.Waiting != nil {
		if state.Waiting.Reason != "" {
			return state.Waiting.Reason
		}
		return "Waiting"
	}
	if state.Terminated != nil {
		if state.Terminated.Reason != "" {
			return state.Terminated.Reason
		}
		return "Terminated"
	}
	return "Unknown"
}

func formatAge(t time.Time) string {
	if t.IsZero() {
		return "0s"
	}
	d := time.Since(t)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%dd", days)
}
