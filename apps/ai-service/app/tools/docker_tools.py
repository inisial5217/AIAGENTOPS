from app.tools.base import ToolDefinition

DOCKER_TOOLS: list[ToolDefinition] = [
    ToolDefinition(
        name="get_container_logs",
        description="Retrieve stdout/stderr logs from a specific Docker container.",
        parameters={
            "type": "object",
            "properties": {
                "container_id": {"type": "string", "description": "ID or name of the container."},
                "tail_lines": {"type": "integer", "description": "Number of recent lines to tail (default 100)."},
            },
            "required": ["container_id"],
        },
        required_role="viewer",
        requires_approval=False,
    ),
    ToolDefinition(
        name="list_docker_containers",
        description="List active or all Docker containers with image and status details.",
        parameters={
            "type": "object",
            "properties": {
                "status_filter": {
                    "type": "string",
                    "enum": ["running", "stopped", "all"],
                    "description": "Container status filter.",
                },
            },
        },
        required_role="viewer",
        requires_approval=False,
    ),
    ToolDefinition(
        name="get_docker_stats",
        description="Get live CPU and memory telemetry statistics for a Docker container.",
        parameters={
            "type": "object",
            "properties": {
                "container_id": {"type": "string", "description": "ID or name of the container."},
            },
            "required": ["container_id"],
        },
        required_role="viewer",
        requires_approval=False,
    ),
    ToolDefinition(
        name="restart_container",
        description="Safely restart a running or crashed Docker container.",
        parameters={
            "type": "object",
            "properties": {
                "container_id": {"type": "string", "description": "ID or name of the container to restart."},
            },
            "required": ["container_id"],
        },
        required_role="devops",
        requires_approval=True,
    ),
    ToolDefinition(
        name="stop_container",
        description="Gracefully stop an active Docker container.",
        parameters={
            "type": "object",
            "properties": {
                "container_id": {"type": "string", "description": "ID or name of the container to stop."},
            },
            "required": ["container_id"],
        },
        required_role="admin",
        requires_approval=True,
    ),
]
