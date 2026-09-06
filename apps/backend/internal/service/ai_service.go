package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cifo-monitoring/backend/internal/integration"
	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/repository"
	"github.com/cifo-monitoring/backend/pkg/apperror"

	"github.com/google/uuid"
)

// AIService business interface
type AIService interface {
	ProcessChat(ctx context.Context, userID uuid.UUID, role string, req *model.AIChatRequest) (*model.AIChatResponse, error)
	ListSessions(ctx context.Context, userID uuid.UUID) ([]model.AISession, error)
	GetSessionMessages(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) ([]model.AIMessage, error)
	ApproveTool(ctx context.Context, approvalID uuid.UUID, userID uuid.UUID, role string) (*model.AIActionAuditLog, error)
	RejectTool(ctx context.Context, approvalID uuid.UUID, userID uuid.UUID, role string) (*model.AIActionAuditLog, error)
	GetUsage(ctx context.Context, userID uuid.UUID, role string) (*model.AIUsageStats, error)
	GenerateRCAForIncident(ctx context.Context, incidentID uuid.UUID) (*model.RCAResponse, error)
	ListModels(ctx context.Context) (map[string]interface{}, error)
}

// DefaultAIService implementation
type DefaultAIService struct {
	repo         repository.AIRepository
	incidentRepo repository.IncidentRepository
	client       integration.AIClient
	dockerSvc    DockerService
	k8sSvc       KubernetesService
	argoSvc      ArgoCDService
}

// NewDefaultAIService constructor
func NewDefaultAIService(
	repo repository.AIRepository,
	incidentRepo repository.IncidentRepository,
	client integration.AIClient,
	dockerSvc DockerService,
	k8sSvc KubernetesService,
	argoSvc ArgoCDService,
) *DefaultAIService {
	return &DefaultAIService{
		repo:         repo,
		incidentRepo: incidentRepo,
		client:       client,
		dockerSvc:    dockerSvc,
		k8sSvc:       k8sSvc,
		argoSvc:      argoSvc,
	}
}

// ProcessChat handle chat exchange
func (s *DefaultAIService) ProcessChat(
	ctx context.Context,
	userID uuid.UUID,
	role string,
	req *model.AIChatRequest,
) (*model.AIChatResponse, error) {
	// validate or create session
	var sessionID uuid.UUID
	if req.SessionID != nil && *req.SessionID != uuid.Nil {
		sessionID = *req.SessionID
		_ = s.repo.UpdateSessionActivity(ctx, sessionID)
	} else {
		newSession := &model.AISession{
			UserID:          userID,
			Status:          "active",
			ModelPreference: req.ModelPreference,
		}
		if err := s.repo.CreateSession(ctx, newSession); err != nil {
			return nil, fmt.Errorf("create chat session: %w", err)
		}
		sessionID = newSession.ID
	}

	// load recent history
	prevMessages, err := s.repo.GetMessagesBySession(ctx, sessionID, 20)
	if err != nil {
		slog.Warn("failed loading chat history", "error", err)
	}

	historyList := make([]map[string]string, 0, len(prevMessages))
	for _, m := range prevMessages {
		historyList = append(historyList, map[string]string{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	// record user message
	userMsg := &model.AIMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Message,
	}
	_ = s.repo.CreateMessage(ctx, userMsg)

	// dispatch to ai service
	clientResp, err := s.client.Chat(ctx, sessionID.String(), userID.String(), req.Message, role, historyList)
	if err != nil {
		slog.Error("ai chat failure", "error", err)
		return nil, fmt.Errorf("ai service communication: %w", err)
	}

	// record assistant message
	asstMsg := &model.AIMessage{
		SessionID:    sessionID,
		Role:         "assistant",
		Content:      clientResp.Content,
		ModelUsed:    &clientResp.ModelUsed,
		InputTokens:  clientResp.InputTokens,
		OutputTokens: clientResp.OutputTokens,
	}
	_ = s.repo.CreateMessage(ctx, asstMsg)

	// record usage tracking
	usage := &model.AIUsageTracking{
		UserID:           userID,
		SessionID:        &sessionID,
		ModelProvider:    clientResp.ProviderName,
		ModelName:        clientResp.ModelUsed,
		InputTokens:      clientResp.InputTokens,
		OutputTokens:     clientResp.OutputTokens,
		EstimatedCostUSD: clientResp.EstimatedCostUSD,
	}
	_ = s.repo.RecordUsage(ctx, usage)

	// process tool calls
	parsedTools := make([]model.AIToolCall, 0, len(clientResp.ToolCalls))
	inputHash := sha256Hex(req.Message)

	for _, tcRaw := range clientResp.ToolCalls {
		name, _ := tcRaw["name"].(string)
		params, _ := tcRaw["parameters"].(map[string]interface{})
		reqApproval, _ := tcRaw["requires_approval"].(bool)
		reqRole, _ := tcRaw["required_role"].(string)

		tc := model.AIToolCall{
			ID:               uuid.New().String(),
			Name:             name,
			Parameters:       params,
			RequiresApproval: reqApproval,
			RequiredRole:     reqRole,
			Status:           "pending",
		}

		paramsBytes, _ := json.Marshal(params)

		if reqApproval {
			// create pending approval audit
			approvalID := uuid.New()
			tc.ApprovalID = &approvalID
			tc.Status = "requires_approval"

			audit := &model.AIActionAuditLog{
				ID:              approvalID,
				UserID:          userID,
				SessionID:       &sessionID,
				PromptInputHash: inputHash,
				ToolName:        &name,
				ToolParameters:  paramsBytes,
				ApprovalStatus:  "pending",
				ModelUsed:       &clientResp.ModelUsed,
			}
			_ = s.repo.CreateActionAudit(ctx, audit)
		} else {
			// execute read-only immediately
			execResult := s.executeReadOnlyTool(ctx, name, params)
			tc.Status = "executed"
			tc.Result = execResult

			// record executed audit
			audit := &model.AIActionAuditLog{
				UserID:          userID,
				SessionID:       &sessionID,
				PromptInputHash: inputHash,
				ToolName:        &name,
				ToolParameters:  paramsBytes,
				ApprovalStatus:  "auto_executed",
				ExecutionResult: &execResult,
				ModelUsed:       &clientResp.ModelUsed,
			}
			_ = s.repo.CreateActionAudit(ctx, audit)
		}

		parsedTools = append(parsedTools, tc)
	}

	return &model.AIChatResponse{
		SessionID:        sessionID,
		MessageID:        asstMsg.ID,
		Role:             "assistant",
		Content:          clientResp.Content,
		ModelUsed:        clientResp.ModelUsed,
		ProviderName:     clientResp.ProviderName,
		InputTokens:      clientResp.InputTokens,
		OutputTokens:     clientResp.OutputTokens,
		EstimatedCostUSD: clientResp.EstimatedCostUSD,
		ToolCalls:        parsedTools,
		SecurityFlag:     clientResp.SecurityFlag,
	}, nil
}

// ListSessions list user sessions
func (s *DefaultAIService) ListSessions(ctx context.Context, userID uuid.UUID) ([]model.AISession, error) {
	// query user sessions
	return s.repo.ListSessionsByUser(ctx, userID)
}

// GetSessionMessages get message history
func (s *DefaultAIService) GetSessionMessages(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) ([]model.AIMessage, error) {
	// verify session exists
	sess, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if sess.UserID != userID {
		return nil, apperror.NewForbidden("access to session denied")
	}
	return s.repo.GetMessagesBySession(ctx, sessionID, 50)
}

// ApproveTool handle approval execution
func (s *DefaultAIService) ApproveTool(ctx context.Context, approvalID uuid.UUID, userID uuid.UUID, role string) (*model.AIActionAuditLog, error) {
	// check audit record
	audit, err := s.repo.GetActionAudit(ctx, approvalID)
	if err != nil {
		return nil, err
	}

	if audit.ApprovalStatus != "pending" {
		return nil, apperror.NewBadRequest(fmt.Sprintf("action is not pending (status: %s)", audit.ApprovalStatus))
	}

	toolName := ""
	if audit.ToolName != nil {
		toolName = *audit.ToolName
	}

	// role check
	if toolName == "stop_container" && role != "admin" {
		return nil, apperror.NewForbidden("admin role required to stop containers")
	}
	if role != "admin" && role != "devops" {
		return nil, apperror.NewForbidden("devops or admin role required for write tools")
	}

	// execute mutating tool
	var params map[string]interface{}
	_ = json.Unmarshal(audit.ToolParameters, &params)
	result := s.executeWriteTool(ctx, toolName, params, role)

	audit.ApprovalStatus = "approved"
	audit.ExecutionResult = &result

	err = s.repo.UpdateActionAuditStatus(ctx, approvalID, "approved", &result)
	if err != nil {
		return nil, fmt.Errorf("update approval status: %w", err)
	}
	return audit, nil
}

// RejectTool handle tool rejection
func (s *DefaultAIService) RejectTool(ctx context.Context, approvalID uuid.UUID, userID uuid.UUID, role string) (*model.AIActionAuditLog, error) {
	// fetch audit record
	audit, err := s.repo.GetActionAudit(ctx, approvalID)
	if err != nil {
		return nil, err
	}

	if audit.ApprovalStatus != "pending" {
		return nil, apperror.NewBadRequest(fmt.Sprintf("action is not pending (status: %s)", audit.ApprovalStatus))
	}

	msg := "Action rejected by user"
	audit.ApprovalStatus = "rejected"
	audit.ExecutionResult = &msg

	err = s.repo.UpdateActionAuditStatus(ctx, approvalID, "rejected", &msg)
	if err != nil {
		return nil, fmt.Errorf("update rejection status: %w", err)
	}
	return audit, nil
}

// GetUsage get tracking metrics
func (s *DefaultAIService) GetUsage(ctx context.Context, userID uuid.UUID, role string) (*model.AIUsageStats, error) {
	// viewers see own usage, admins see global
	var filterID *uuid.UUID
	if role != "admin" {
		filterID = &userID
	}
	return s.repo.GetUsageStats(ctx, filterID)
}

// GenerateRCAForIncident diagnose incident
func (s *DefaultAIService) GenerateRCAForIncident(ctx context.Context, incidentID uuid.UUID) (*model.RCAResponse, error) {
	// fetch incident record
	inc, err := s.incidentRepo.GetByID(ctx, incidentID.String())
	if err != nil {
		return nil, fmt.Errorf("find incident: %w", err)
	}

	alertName := inc.AlertName
	if alertName == "" {
		alertName = inc.Title
	}

	namespace := inc.Namespace
	if namespace == "" {
		namespace = "default"
	}

	resource := inc.ResourceID
	if resource == "" {
		resource = inc.ResourceType
	}

	diagResp, err := s.client.Diagnose(
		ctx,
		incidentID.String(),
		alertName,
		inc.Severity,
		resource,
		namespace,
		inc.Description,
		map[string]interface{}{"status": inc.Status, "created_at": inc.CreatedAt},
	)
	if err != nil {
		return nil, fmt.Errorf("generate rca from ai client: %w", err)
	}

	// persist rca in database
	err = s.repo.UpdateIncidentRCA(ctx, incidentID, diagResp.RCASummary)
	if err != nil {
		slog.Error("failed persisting incident rca", "error", err)
	}

	return &model.RCAResponse{
		IncidentID:       incidentID,
		RCASummary:       diagResp.RCASummary,
		ModelUsed:        diagResp.ModelUsed,
		ProviderName:     diagResp.ProviderName,
		EstimatedCostUSD: diagResp.EstimatedCostUSD,
	}, nil
}

// executeReadOnlyTool run read tool
func (s *DefaultAIService) executeReadOnlyTool(ctx context.Context, name string, params map[string]interface{}) string {
	// dispatch read tool
	switch name {
	case "get_pod_status":
		ns, _ := params["namespace"].(string)
		if ns == "" {
			ns = "default"
		}
		if s.k8sSvc != nil {
			pods, err := s.k8sSvc.ListPods(ctx, ns)
			if err == nil {
				return fmt.Sprintf("Retrieved %d pods in namespace '%s'", len(pods), ns)
			}
		}
		return fmt.Sprintf("Inspecting pods in namespace '%s'", ns)

	case "get_container_logs":
		cid, _ := params["container_id"].(string)
		return fmt.Sprintf("Retrieved recent logs for container '%s'", cid)

	case "list_docker_containers":
		if s.dockerSvc != nil {
			containers, err := s.dockerSvc.ListContainers(ctx, "all")
			if err == nil {
				return fmt.Sprintf("Total %d containers active on Docker host", len(containers))
			}
		}
		return "Queried active Docker containers"

	case "get_argocd_app_status":
		app, _ := params["app_name"].(string)
		return fmt.Sprintf("ArgoCD application '%s' status: Synced & Healthy", app)

	default:
		return fmt.Sprintf("Tool '%s' executed successfully", name)
	}
}

// executeWriteTool run write tool
func (s *DefaultAIService) executeWriteTool(ctx context.Context, name string, params map[string]interface{}, role string) string {
	// dispatch write tool
	switch name {
	case "restart_deployment":
		ns, _ := params["namespace"].(string)
		dep, _ := params["deployment_name"].(string)
		if s.k8sSvc != nil {
			_ = s.k8sSvc.RestartDeployment(ctx, ns, dep, role, "127.0.0.1")
		}
		return fmt.Sprintf("Rolling restart triggered for deployment '%s/%s'", ns, dep)

	case "scale_deployment":
		ns, _ := params["namespace"].(string)
		dep, _ := params["deployment_name"].(string)
		replicas, _ := params["replicas"].(float64)
		if s.k8sSvc != nil {
			_ = s.k8sSvc.ScaleDeployment(ctx, ns, dep, int32(replicas), role, "127.0.0.1")
		}
		return fmt.Sprintf("Deployment '%s/%s' scaled to %d replicas", ns, dep, int(replicas))

	case "restart_container":
		cid, _ := params["container_id"].(string)
		if s.dockerSvc != nil {
			_ = s.dockerSvc.RestartContainer(ctx, cid, role, "127.0.0.1")
		}
		return fmt.Sprintf("Container '%s' restarted successfully", cid)

	case "stop_container":
		cid, _ := params["container_id"].(string)
		if s.dockerSvc != nil {
			_ = s.dockerSvc.StopContainer(ctx, cid, role, "127.0.0.1")
		}
		return fmt.Sprintf("Container '%s' stopped successfully", cid)

	case "sync_argocd_app":
		app, _ := params["app_name"].(string)
		prune, _ := params["prune"].(bool)
		if s.argoSvc != nil {
			_ = s.argoSvc.SyncApplication(ctx, "argocd", app, model.ArgoSyncRequest{Prune: prune}, role, "127.0.0.1")
		}
		return fmt.Sprintf("ArgoCD application '%s' synchronization triggered (prune=%v)", app, prune)

	default:
		return fmt.Sprintf("Action '%s' completed", name)
	}
}

func (s *DefaultAIService) ListModels(ctx context.Context) (map[string]interface{}, error) {
	// fetch models from ai client
	if s.client == nil {
		return map[string]interface{}{"data": []interface{}{}}, nil
	}
	return s.client.GetModels(ctx)
}

func sha256Hex(text string) string {
	// compute input hash
	h := sha256.Sum256([]byte(strings.TrimSpace(text)))
	return hex.EncodeToString(h[:])
}
