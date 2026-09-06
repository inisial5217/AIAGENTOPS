from app.tools.base import ToolDefinition

KUBECTL_TOOLS: list[ToolDefinition] = [
    ToolDefinition(
        name="get_pod_status",
        description="Get real-time pod statuses, health states, and restart counts across Kubernetes namespaces.",
        parameters={
            "type": "object",
            "properties": {
                "namespace": {
                    "type": "string",
                    "description": "Kubernetes namespace (e.g. default, kube-system, argocd).",
                },
                "pod_name": {
                    "type": "string",
                    "description": "Optional specific pod name.",
                },
            },
            "required": ["namespace"],
        },
        required_role="viewer",
        requires_approval=False,
    ),
    ToolDefinition(
        name="get_deployment_info",
        description="Get Kubernetes deployment specifications, available replicas, and status.",
        parameters={
            "type": "object",
            "properties": {
                "namespace": {"type": "string", "description": "Kubernetes namespace."},
                "deployment_name": {"type": "string", "description": "Deployment name to inspect."},
            },
            "required": ["namespace", "deployment_name"],
        },
        required_role="viewer",
        requires_approval=False,
    ),
    ToolDefinition(
        name="get_node_resources",
        description="Inspect cluster node resource capacity, CPU, and memory allocation.",
        parameters={
            "type": "object",
            "properties": {
                "node_name": {"type": "string", "description": "Optional specific node name."},
            },
        },
        required_role="viewer",
        requires_approval=False,
    ),
    ToolDefinition(
        name="restart_deployment",
        description="Perform a rolling restart of a Kubernetes deployment to recover failing pods.",
        parameters={
            "type": "object",
            "properties": {
                "namespace": {"type": "string", "description": "Kubernetes namespace."},
                "deployment_name": {"type": "string", "description": "Target deployment name."},
            },
            "required": ["namespace", "deployment_name"],
        },
        required_role="devops",
        requires_approval=True,
    ),
    ToolDefinition(
        name="scale_deployment",
        description="Scale the number of replica pods for a Kubernetes deployment.",
        parameters={
            "type": "object",
            "properties": {
                "namespace": {"type": "string", "description": "Kubernetes namespace."},
                "deployment_name": {"type": "string", "description": "Target deployment name."},
                "replicas": {"type": "integer", "description": "Desired replica count (e.g. 1-10)."},
            },
            "required": ["namespace", "deployment_name", "replicas"],
        },
        required_role="devops",
        requires_approval=True,
    ),
]
