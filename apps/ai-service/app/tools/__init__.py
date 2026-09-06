from typing import Any
from app.tools.base import ToolDefinition, ToolCallRequest
from app.tools.kubectl_tools import KUBECTL_TOOLS
from app.tools.docker_tools import DOCKER_TOOLS
from app.tools.argocd_tools import ARGOCD_TOOLS

ALL_TOOLS: list[ToolDefinition] = KUBECTL_TOOLS + DOCKER_TOOLS + ARGOCD_TOOLS
TOOL_REGISTRY: dict[str, ToolDefinition] = {t.name: t for t in ALL_TOOLS}


def get_all_tools() -> list[ToolDefinition]:
    # return all registered tools
    return ALL_TOOLS


def find_tool(name: str) -> ToolDefinition | None:
    # lookup tool definition
    return TOOL_REGISTRY.get(name)


def get_tool_schemas() -> list[dict[str, Any]]:
    # export tool schemas
    return [t.to_schema() for t in ALL_TOOLS]


__all__ = [
    "ToolDefinition",
    "ToolCallRequest",
    "ALL_TOOLS",
    "get_all_tools",
    "find_tool",
    "get_tool_schemas",
]
