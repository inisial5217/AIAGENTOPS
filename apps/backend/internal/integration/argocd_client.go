package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/pkg/apperror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var argoAppGVR = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "applications",
}

const defaultArgoNamespace = "argocd"

// ArgoCDClient interface for argocd operations
type ArgoCDClient interface {
	ListApplications(ctx context.Context, namespace string) ([]model.ArgoApplicationSummary, error)
	GetApplication(ctx context.Context, namespace string, name string) (*model.ArgoApplicationDetail, error)
	SyncApplication(ctx context.Context, namespace string, name string, req model.ArgoSyncRequest) error
}

type argoCDClientImpl struct {
	dynClient dynamic.Interface
}

// NewArgoCDClient creates dynamic argocd client
func NewArgoCDClient(config *rest.Config) (ArgoCDClient, error) {
	if config == nil {
		return nil, apperror.NewInternal("rest config is nil")
	}
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("init dynamic client: %v", err))
	}
	return &argoCDClientImpl{
		dynClient: dynClient,
	}, nil
}

func resolveNamespace(namespace string) string {
	if namespace == "" {
		return defaultArgoNamespace
	}
	return namespace
}

// ListApplications retrieves application summaries
func (c *argoCDClientImpl) ListApplications(ctx context.Context, namespace string) ([]model.ArgoApplicationSummary, error) {
	ns := resolveNamespace(namespace)
	list, err := c.dynClient.Resource(argoAppGVR).Namespace(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("list argo applications: %v", err))
	}

	result := make([]model.ArgoApplicationSummary, 0, len(list.Items))
	for _, item := range list.Items {
		summary := parseApplicationSummary(&item)
		result = append(result, summary)
	}
	return result, nil
}

// GetApplication retrieves detailed application
func (c *argoCDClientImpl) GetApplication(ctx context.Context, namespace string, name string) (*model.ArgoApplicationDetail, error) {
	ns := resolveNamespace(namespace)
	item, err := c.dynClient.Resource(argoAppGVR).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, apperror.NewNotFound("argo application not found")
	}

	summary := parseApplicationSummary(item)
	detail := &model.ArgoApplicationDetail{
		ArgoApplicationSummary: summary,
		Resources:              make([]model.ArgoResourceStatus, 0),
		History:                make([]model.ArgoDeploymentRevision, 0),
	}

	if syncPolicy, found, _ := unstructured.NestedMap(item.Object, "spec", "syncPolicy"); found {
		if automated, ok := syncPolicy["automated"].(map[string]interface{}); ok {
			detail.AutomatedSync = true
			if prune, ok := automated["prune"].(bool); ok {
				detail.Prune = prune
			}
			if selfHeal, ok := automated["selfHeal"].(bool); ok {
				detail.SelfHeal = selfHeal
			}
		}
	}

	if rawResources, found, _ := unstructured.NestedSlice(item.Object, "status", "resources"); found {
		for _, raw := range rawResources {
			resMap, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			res := model.ArgoResourceStatus{
				Group:     getString(resMap, "group"),
				Version:   getString(resMap, "version"),
				Kind:      getString(resMap, "kind"),
				Namespace: getString(resMap, "namespace"),
				Name:      getString(resMap, "name"),
				Status:    getString(resMap, "status"),
				Hook:      getString(resMap, "hook"),
			}
			if healthMap, ok := resMap["health"].(map[string]interface{}); ok {
				res.Health = getString(healthMap, "status")
			}
			detail.Resources = append(detail.Resources, res)
		}
	}

	if rawHistory, found, _ := unstructured.NestedSlice(item.Object, "status", "history"); found {
		for _, raw := range rawHistory {
			histMap, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			hist := model.ArgoDeploymentRevision{
				Revision:   getString(histMap, "revision"),
				DeployedAt: getString(histMap, "deployedAt"),
			}
			if idVal, ok := histMap["id"].(int64); ok {
				hist.ID = idVal
			} else if idFloat, ok := histMap["id"].(float64); ok {
				hist.ID = int64(idFloat)
			}
			if sourceMap, ok := histMap["source"].(map[string]interface{}); ok {
				hist.Source = getString(sourceMap, "path")
				if hist.Source == "" {
					hist.Source = getString(sourceMap, "repoURL")
				}
			}
			detail.History = append(detail.History, hist)
		}
	}

	return detail, nil
}

// SyncApplication triggers sync operation
func (c *argoCDClientImpl) SyncApplication(ctx context.Context, namespace string, name string, req model.ArgoSyncRequest) error {
	ns := resolveNamespace(namespace)
	patchMap := map[string]interface{}{
		"operation": map[string]interface{}{
			"sync": map[string]interface{}{
				"prune":  req.Prune,
				"dryRun": req.DryRun,
				"syncStrategy": map[string]interface{}{
					"apply": map[string]interface{}{
						"force": false,
					},
				},
			},
			"initiatedBy": map[string]interface{}{
				"username":  "cifo-dashboard",
				"automated": false,
			},
		},
	}

	patchBytes, err := json.Marshal(patchMap)
	if err != nil {
		return apperror.NewInternal(fmt.Sprintf("marshal sync patch: %v", err))
	}

	_, err = c.dynClient.Resource(argoAppGVR).Namespace(ns).Patch(ctx, name, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return apperror.NewInternal(fmt.Sprintf("trigger argo sync: %v", err))
	}
	return nil
}

func parseApplicationSummary(item *unstructured.Unstructured) model.ArgoApplicationSummary {
	summary := model.ArgoApplicationSummary{
		Name:                 item.GetName(),
		Project:              getNestedString(item.Object, "spec", "project"),
		RepoURL:              getNestedString(item.Object, "spec", "source", "repoURL"),
		Path:                 getNestedString(item.Object, "spec", "source", "path"),
		TargetRevision:       getNestedString(item.Object, "spec", "source", "targetRevision"),
		DestinationServer:    getNestedString(item.Object, "spec", "destination", "server"),
		DestinationNamespace: getNestedString(item.Object, "spec", "destination", "namespace"),
		SyncStatus:           getNestedString(item.Object, "status", "sync", "status"),
		HealthStatus:         getNestedString(item.Object, "status", "health", "status"),
		SyncMessage:          getNestedString(item.Object, "status", "operationState", "message"),
		HealthMessage:        getNestedString(item.Object, "status", "health", "message"),
		CreatedAt:            item.GetCreationTimestamp().Time.Format(time.RFC3339),
	}

	if summary.SyncStatus == "" {
		summary.SyncStatus = "Unknown"
	}
	if summary.HealthStatus == "" {
		summary.HealthStatus = "Unknown"
	}

	if images, found, _ := unstructured.NestedStringSlice(item.Object, "status", "summary", "images"); found {
		summary.Images = images
	}

	return summary
}

func getNestedString(obj map[string]interface{}, fields ...string) string {
	val, found, _ := unstructured.NestedString(obj, fields...)
	if !found {
		return ""
	}
	return val
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
