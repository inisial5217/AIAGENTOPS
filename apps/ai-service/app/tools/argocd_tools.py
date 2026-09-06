from app.tools.base import ToolDefinition

ARGOCD_TOOLS: list[ToolDefinition] = [
    ToolDefinition(
        name="get_argocd_app_status",
        description="Inspect ArgoCD GitOps application sync status, health status, and repository commit.",
        parameters={
            "type": "object",
            "properties": {
                "app_name": {"type": "string", "description": "Name of the ArgoCD application."},
            },
            "required": ["app_name"],
        },
        required_role="viewer",
        requires_approval=False,
    ),
    ToolDefinition(
        name="get_argocd_history",
        description="Retrieve deployment and sync revision history for an ArgoCD application.",
        parameters={
            "type": "object",
            "properties": {
                "app_name": {"type": "string", "description": "Name of the ArgoCD application."},
                "limit": {"type": "integer", "description": "Number of history items to return (default 10)."},
            },
            "required": ["app_name"],
        },
        required_role="viewer",
        requires_approval=False,
    ),
    ToolDefinition(
        name="sync_argocd_app",
        description="Trigger a manual GitOps sync for an ArgoCD application to reconcile desired state.",
        parameters={
            "type": "object",
            "properties": {
                "app_name": {"type": "string", "description": "Target ArgoCD application name."},
                "prune": {"type": "boolean", "description": "Whether to delete orphaned resources (default false)."},
            },
            "required": ["app_name"],
        },
        required_role="devops",
        requires_approval=True,
    ),
]
