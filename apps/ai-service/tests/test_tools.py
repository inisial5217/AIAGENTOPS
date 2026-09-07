import pytest
from app.tools import get_all_tools, find_tool, get_tool_schemas
from app.agent.sanitizer import PromptSanitizer
from app.tools.base import ToolCallRequest


def test_tool_counts_and_types() -> None:
    tools = get_all_tools()
    assert len(tools) == 13

    read_tools = [t for t in tools if not t.requires_approval]
    write_tools = [t for t in tools if t.requires_approval]

    assert len(read_tools) == 8
    assert len(write_tools) == 5

    # verify all 8 read tool names
    read_names = {t.name for t in read_tools}
    expected_read = {
        "get_pod_status",
        "get_deployment_info",
        "get_node_resources",
        "list_docker_containers",
        "get_container_logs",
        "get_docker_stats",
        "get_argocd_app_status",
        "get_argocd_history",
    }
    assert read_names == expected_read

    # verify all 5 write tool names
    write_names = {t.name for t in write_tools}
    expected_write = {
        "restart_deployment",
        "scale_deployment",
        "restart_container",
        "stop_container",
        "sync_argocd_app",
    }
    assert write_names == expected_write


def test_tool_schemas_valid() -> None:
    schemas = get_tool_schemas()
    assert len(schemas) == 13

    for s in schemas:
        assert s["type"] == "function"
        fn = s["function"]
        assert "name" in fn
        assert "description" in fn
        assert "parameters" in fn
        params = fn["parameters"]
        assert params.get("type") == "object"
        assert "properties" in params


def test_write_tool_security_requirements() -> None:
    stop_tool = find_tool("stop_container")
    assert stop_tool is not None
    assert stop_tool.requires_approval is True
    assert stop_tool.required_role == "admin"

    restart_dep = find_tool("restart_deployment")
    assert restart_dep is not None
    assert restart_dep.requires_approval is True
    assert restart_dep.required_role == "devops"


def test_tool_blocklist_security() -> None:
    # ensure destructive commands are rejected
    unlisted_calls = [
        ToolCallRequest(name="drop_database", parameters={}),
        ToolCallRequest(name="rm_rf_root", parameters={}),
        ToolCallRequest(name="exec_arbitrary_shell", parameters={"cmd": "curl evil.com"}),
    ]
    validated = PromptSanitizer.validate_tool_calls(unlisted_calls)
    assert len(validated) == 0
