package service

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
)

// MockKubernetesClient mock k8s client
type MockKubernetesClient struct {
	mock.Mock
}

func (m *MockKubernetesClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockKubernetesClient) ListPods(ctx context.Context, namespace string) (*corev1.PodList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*corev1.PodList), args.Error(1)
}

func (m *MockKubernetesClient) GetPod(ctx context.Context, namespace string, name string) (*corev1.Pod, error) {
	args := m.Called(ctx, namespace, name)
	return args.Get(0).(*corev1.Pod), args.Error(1)
}

func (m *MockKubernetesClient) GetPodLogs(ctx context.Context, namespace string, name string, container string, tailLines int64) (io.ReadCloser, error) {
	args := m.Called(ctx, namespace, name, container, tailLines)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockKubernetesClient) ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*appsv1.DeploymentList), args.Error(1)
}

func (m *MockKubernetesClient) GetDeployment(ctx context.Context, namespace string, name string) (*appsv1.Deployment, error) {
	args := m.Called(ctx, namespace, name)
	return args.Get(0).(*appsv1.Deployment), args.Error(1)
}

func (m *MockKubernetesClient) RestartDeployment(ctx context.Context, namespace string, name string) error {
	args := m.Called(ctx, namespace, name)
	return args.Error(0)
}

func (m *MockKubernetesClient) ScaleDeployment(ctx context.Context, namespace string, name string, replicas int32) error {
	args := m.Called(ctx, namespace, name, replicas)
	return args.Error(0)
}

func (m *MockKubernetesClient) ListNodes(ctx context.Context) (*corev1.NodeList, error) {
	args := m.Called(ctx)
	return args.Get(0).(*corev1.NodeList), args.Error(1)
}

func (m *MockKubernetesClient) GetNode(ctx context.Context, name string) (*corev1.Node, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*corev1.Node), args.Error(1)
}

func (m *MockKubernetesClient) ListServices(ctx context.Context, namespace string) (*corev1.ServiceList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*corev1.ServiceList), args.Error(1)
}

func (m *MockKubernetesClient) ListEvents(ctx context.Context, namespace string) (*corev1.EventList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*corev1.EventList), args.Error(1)
}

func (m *MockKubernetesClient) WatchEvents(ctx context.Context, namespace string) (watch.Interface, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(watch.Interface), args.Error(1)
}

func (m *MockKubernetesClient) GetRESTConfig() *rest.Config {
	return nil
}

// MockAuditRepo mock audit repository
type MockAuditRepo struct {
	mock.Mock
}

func (m *MockAuditRepo) Create(ctx context.Context, log *model.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditRepo) List(ctx context.Context, limit, offset int) ([]*model.AuditLog, int, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*model.AuditLog), args.Int(1), args.Error(2)
}

func TestListPods(t *testing.T) {
	mockClient := new(MockKubernetesClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewKubernetesService(mockClient, nil, logger)

	podList := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "nginx-abc-123",
					Namespace:         "default",
					CreationTimestamp: metav1.Time{Time: time.Now().Add(-10 * time.Minute)},
					Labels:            map[string]string{"app": "nginx"},
				},
				Spec: corev1.PodSpec{
					NodeName: "node-1",
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					PodIP: "10.42.0.15",
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "nginx",
							Ready:        true,
							RestartCount: 2,
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
		},
	}

	mockClient.On("ListPods", mock.Anything, "default").Return(podList, nil)

	ctx := context.Background()
	pods, err := svc.ListPods(ctx, "default")
	assert.NoError(t, err)
	assert.Len(t, pods, 1)
	assert.Equal(t, "nginx-abc-123", pods[0].Name)
	assert.Equal(t, "Running", pods[0].Status)
	assert.Equal(t, int32(2), pods[0].Restarts)
	assert.Equal(t, "100m", pods[0].CPURequest)
	assert.Equal(t, "128Mi", pods[0].MemoryRequest)
	assert.Equal(t, "10m", pods[0].Age)
}

func TestGetPod(t *testing.T) {
	mockClient := new(MockKubernetesClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewKubernetesService(mockClient, nil, logger)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web-pod",
			Namespace: "production",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "web",
					Image: "cifo/web:v1",
					Ports: []corev1.ContainerPort{
						{ContainerPort: 80, Protocol: corev1.ProtocolTCP},
					},
				},
			},
		},
		Status: corev1.PodStatus{
			Phase:    corev1.PodRunning,
			QOSClass: corev1.PodQOSBurstable,
			HostIP:   "192.168.1.10",
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "web",
					Ready: true,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{},
					},
				},
			},
		},
	}

	mockClient.On("GetPod", mock.Anything, "production", "web-pod").Return(pod, nil)

	ctx := context.Background()
	detail, err := svc.GetPod(ctx, "production", "web-pod")
	assert.NoError(t, err)
	assert.Equal(t, "web-pod", detail.Name)
	assert.Equal(t, "Burstable", detail.QoSClass)
	assert.Len(t, detail.Containers, 1)
	assert.Equal(t, "80/TCP", detail.Containers[0].Ports[0])
	assert.Contains(t, detail.Conditions, "Ready")
}

func TestGetPodLogs(t *testing.T) {
	mockClient := new(MockKubernetesClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewKubernetesService(mockClient, nil, logger)

	mockReader := io.NopCloser(strings.NewReader("server started on :8080\nready to accept connections"))
	mockClient.On("GetPodLogs", mock.Anything, "default", "pod-1", "app", int64(100)).Return(mockReader, nil)

	ctx := context.Background()
	reader, err := svc.GetPodLogs(ctx, "default", "pod-1", "app", 100)
	assert.NoError(t, err)
	defer reader.Close()

	bytes, readErr := io.ReadAll(reader)
	assert.NoError(t, readErr)
	assert.Contains(t, string(bytes), "server started on :8080")
}

func TestListDeployments(t *testing.T) {
	mockClient := new(MockKubernetesClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewKubernetesService(mockClient, nil, logger)

	replicas := int32(3)
	deployList := &appsv1.DeploymentList{
		Items: []appsv1.Deployment{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "api-deployment",
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replicas,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Image: "cifo/api:v1.2"},
							},
						},
					},
				},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas:     3,
					AvailableReplicas: 3,
					UpdatedReplicas:   3,
				},
			},
		},
	}

	mockClient.On("ListDeployments", mock.Anything, "default").Return(deployList, nil)

	ctx := context.Background()
	deps, err := svc.ListDeployments(ctx, "default")
	assert.NoError(t, err)
	assert.Len(t, deps, 1)
	assert.Equal(t, "api-deployment", deps[0].Name)
	assert.Equal(t, int32(3), deps[0].Replicas)
	assert.Equal(t, int32(3), deps[0].ReadyReplicas)
	assert.Equal(t, "cifo/api:v1.2", deps[0].Images[0])
}

func TestRestartDeployment(t *testing.T) {
	mockClient := new(MockKubernetesClient)
	mockAudit := new(MockAuditRepo)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewKubernetesService(mockClient, mockAudit, logger)

	mockClient.On("RestartDeployment", mock.Anything, "default", "api-deployment").Return(nil)
	mockAudit.On("Create", mock.Anything, mock.MatchedBy(func(l *model.AuditLog) bool {
		return l.Action == "restart_k8s_deployment" && l.ResourceType == "kubernetes"
	})).Return(nil)

	ctx := context.Background()
	err := svc.RestartDeployment(ctx, "default", "api-deployment", "admin-user", "127.0.0.1")
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestScaleDeployment(t *testing.T) {
	mockClient := new(MockKubernetesClient)
	mockAudit := new(MockAuditRepo)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewKubernetesService(mockClient, mockAudit, logger)

	mockClient.On("ScaleDeployment", mock.Anything, "default", "api-deployment", int32(5)).Return(nil)
	mockAudit.On("Create", mock.Anything, mock.MatchedBy(func(l *model.AuditLog) bool {
		return l.Action == "scale_k8s_deployment"
	})).Return(nil)

	ctx := context.Background()
	err := svc.ScaleDeployment(ctx, "default", "api-deployment", 5, "devops-user", "127.0.0.1")
	assert.NoError(t, err)

	// test negative replicas error
	negErr := svc.ScaleDeployment(ctx, "default", "api-deployment", -1, "devops-user", "127.0.0.1")
	assert.Error(t, negErr)
}

func TestListNodes(t *testing.T) {
	mockClient := new(MockKubernetesClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewKubernetesService(mockClient, nil, logger)

	nodeList := &corev1.NodeList{
		Items: []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "k3d-server-0",
					Labels: map[string]string{"node-role.kubernetes.io/control-plane": "true"},
				},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
					},
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeInternalIP, Address: "172.18.0.2"},
					},
					NodeInfo: corev1.NodeSystemInfo{
						KubeletVersion: "v1.28.8+k3s1",
						OSImage:        "Ubuntu 22.04 LTS",
					},
				},
			},
		},
	}

	podList := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-1"},
				Spec:       corev1.PodSpec{NodeName: "k3d-server-0"},
			},
		},
	}

	mockClient.On("ListNodes", mock.Anything).Return(nodeList, nil)
	mockClient.On("ListPods", mock.Anything, "").Return(podList, nil)

	ctx := context.Background()
	nodes, err := svc.ListNodes(ctx)
	assert.NoError(t, err)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "k3d-server-0", nodes[0].Name)
	assert.Equal(t, "Ready", nodes[0].Status)
	assert.Equal(t, "control-plane", nodes[0].Roles)
	assert.Equal(t, 1, nodes[0].PodCount)
	assert.Equal(t, "172.18.0.2", nodes[0].InternalIP)
}

func TestGetClusterOverview(t *testing.T) {
	mockClient := new(MockKubernetesClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := NewKubernetesService(mockClient, nil, logger)

	nodeList := &corev1.NodeList{
		Items: []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "n1"},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}},
				},
			},
		},
	}

	podList := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "p1"},
				Spec:       corev1.PodSpec{NodeName: "n1"},
				Status:     corev1.PodStatus{Phase: corev1.PodRunning},
			},
		},
	}

	replicas := int32(1)
	deployList := &appsv1.DeploymentList{
		Items: []appsv1.Deployment{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "d1"},
				Spec:       appsv1.DeploymentSpec{Replicas: &replicas},
				Status:     appsv1.DeploymentStatus{ReadyReplicas: 1, Replicas: 1},
			},
		},
	}

	svcList := &corev1.ServiceList{
		Items: []corev1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "s1"},
				Spec:       corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP},
			},
		},
	}

	mockClient.On("ListNodes", mock.Anything).Return(nodeList, nil)
	mockClient.On("ListPods", mock.Anything, "").Return(podList, nil)
	mockClient.On("ListDeployments", mock.Anything, "").Return(deployList, nil)
	mockClient.On("ListServices", mock.Anything, "").Return(svcList, nil)

	ctx := context.Background()
	overview, err := svc.GetClusterOverview(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, overview.TotalNodes)
	assert.Equal(t, 1, overview.ReadyNodes)
	assert.Equal(t, 1, overview.TotalPods)
	assert.Equal(t, 1, overview.RunningPods)
	assert.Equal(t, 1, overview.TotalDeployments)
	assert.Equal(t, 1, overview.ReadyDeployments)
	assert.Equal(t, 1, overview.TotalServices)
}
