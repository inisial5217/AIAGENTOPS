package integration

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/cifo-monitoring/backend/pkg/apperror"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesClient interface for k8s api operations
type KubernetesClient interface {
	Ping(ctx context.Context) error
	ListPods(ctx context.Context, namespace string) (*corev1.PodList, error)
	GetPod(ctx context.Context, namespace string, name string) (*corev1.Pod, error)
	GetPodLogs(ctx context.Context, namespace string, name string, container string, tailLines int64) (io.ReadCloser, error)
	ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error)
	GetDeployment(ctx context.Context, namespace string, name string) (*appsv1.Deployment, error)
	RestartDeployment(ctx context.Context, namespace string, name string) error
	ScaleDeployment(ctx context.Context, namespace string, name string, replicas int32) error
	ListNodes(ctx context.Context) (*corev1.NodeList, error)
	GetNode(ctx context.Context, name string) (*corev1.Node, error)
	ListServices(ctx context.Context, namespace string) (*corev1.ServiceList, error)
	ListEvents(ctx context.Context, namespace string) (*corev1.EventList, error)
	WatchEvents(ctx context.Context, namespace string) (watch.Interface, error)
	GetRESTConfig() *rest.Config
}

type kubernetesClientImpl struct {
	clientset  *kubernetes.Clientset
	restConfig *rest.Config
}

// NewKubernetesClient initializes k8s client
func NewKubernetesClient(kubeconfigPath string) (KubernetesClient, error) {
	var config *rest.Config
	var err error

	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("KUBECONFIG")
	}

	if kubeconfigPath == "" {
		home, _ := os.UserHomeDir()
		if home != "" {
			defaultPath := filepath.Join(home, ".kube", "config")
			if _, statErr := os.Stat(defaultPath); statErr == nil {
				kubeconfigPath = defaultPath
			}
		}
	}

	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("failed to load kubeconfig: %v", err))
	}

	config.Timeout = 15 * time.Second
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("failed to init kubernetes clientset: %v", err))
	}

	return &kubernetesClientImpl{
		clientset:  clientset,
		restConfig: config,
	}, nil
}

// Ping checks cluster connectivity
func (c *kubernetesClientImpl) Ping(ctx context.Context) error {
	_, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return apperror.NewInternal(fmt.Sprintf("cluster ping failed: %v", err))
	}
	return nil
}

// ListPods returns pods in namespace
func (c *kubernetesClientImpl) ListPods(ctx context.Context, namespace string) (*corev1.PodList, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("list pods: %v", err))
	}
	return pods, nil
}

// GetPod inspects single pod
func (c *kubernetesClientImpl) GetPod(ctx context.Context, namespace string, name string) (*corev1.Pod, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, apperror.NewNotFound(fmt.Sprintf("pod %s not found in namespace %s: %v", name, namespace, err))
	}
	return pod, nil
}

// GetPodLogs returns log reader
func (c *kubernetesClientImpl) GetPodLogs(ctx context.Context, namespace string, name string, container string, tailLines int64) (io.ReadCloser, error) {
	opts := &corev1.PodLogOptions{}
	if tailLines > 0 {
		opts.TailLines = &tailLines
	}
	if container != "" {
		opts.Container = container
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(name, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("get pod logs: %v", err))
	}
	return stream, nil
}

// ListDeployments returns deployments in namespace
func (c *kubernetesClientImpl) ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error) {
	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("list deployments: %v", err))
	}
	return deployments, nil
}

// GetDeployment inspects single deployment
func (c *kubernetesClientImpl) GetDeployment(ctx context.Context, namespace string, name string) (*appsv1.Deployment, error) {
	dep, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, apperror.NewNotFound(fmt.Sprintf("deployment %s not found in namespace %s: %v", name, namespace, err))
	}
	return dep, nil
}

// RestartDeployment triggers rollout restart
func (c *kubernetesClientImpl) RestartDeployment(ctx context.Context, namespace string, name string) error {
	patch := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":%q}}}}}`, time.Now().Format(time.RFC3339))
	_, err := c.clientset.AppsV1().Deployments(namespace).Patch(ctx, name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return apperror.NewInternal(fmt.Sprintf("restart deployment: %v", err))
	}
	return nil
}

// ScaleDeployment updates replicas count
func (c *kubernetesClientImpl) ScaleDeployment(ctx context.Context, namespace string, name string, replicas int32) error {
	scale, err := c.clientset.AppsV1().Deployments(namespace).GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return apperror.NewInternal(fmt.Sprintf("get scale: %v", err))
	}
	scale.Spec.Replicas = replicas
	_, err = c.clientset.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
	if err != nil {
		return apperror.NewInternal(fmt.Sprintf("update scale: %v", err))
	}
	return nil
}

// ListNodes returns cluster nodes
func (c *kubernetesClientImpl) ListNodes(ctx context.Context) (*corev1.NodeList, error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("list nodes: %v", err))
	}
	return nodes, nil
}

// GetNode inspects single node
func (c *kubernetesClientImpl) GetNode(ctx context.Context, name string) (*corev1.Node, error) {
	node, err := c.clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, apperror.NewNotFound(fmt.Sprintf("node %s not found: %v", name, err))
	}
	return node, nil
}

// ListServices returns services in namespace
func (c *kubernetesClientImpl) ListServices(ctx context.Context, namespace string) (*corev1.ServiceList, error) {
	services, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("list services: %v", err))
	}
	return services, nil
}

// ListEvents returns cluster events
func (c *kubernetesClientImpl) ListEvents(ctx context.Context, namespace string) (*corev1.EventList, error) {
	events, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("list events: %v", err))
	}
	return events, nil
}

// WatchEvents streams cluster events
func (c *kubernetesClientImpl) WatchEvents(ctx context.Context, namespace string) (watch.Interface, error) {
	watcher, err := c.clientset.CoreV1().Events(namespace).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("watch events: %v", err))
	}
	return watcher, nil
}

// GetRESTConfig returns raw rest config
func (c *kubernetesClientImpl) GetRESTConfig() *rest.Config {
	return c.restConfig
}
