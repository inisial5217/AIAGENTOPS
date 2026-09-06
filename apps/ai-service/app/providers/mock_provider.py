from app.providers.base import LLMProvider, ChatMessage, ProviderResponse
from app.tools.base import ToolDefinition, ToolCallRequest


class DeterministicMockProvider(LLMProvider):
    def __init__(self, model_name: str = "cifo-deterministic-mock") -> None:
        super().__init__(provider_name="mock", model_name=model_name)

    def estimate_cost(self, input_tokens: int, output_tokens: int) -> float:
        # zero cost test model
        return 0.0

    async def chat(
        self,
        messages: list[ChatMessage],
        tools: list[ToolDefinition] | None = None,
        system_instruction: str = "",
    ) -> ProviderResponse:
        last_msg = messages[-1].content.lower() if messages else ""

        # simulate tool call triggers
        tool_calls: list[ToolCallRequest] = []
        content = "CIFO SRE Assistant analysis completed."

        if "root cause" in last_msg or "rca" in last_msg or "diagnos" in last_msg:
            content = (
                "### 1. Incident Overview\n"
                "- **Status**: Incident Diagnosed\n\n"
                "### 2. Root Cause Hypothesis\n"
                "Analysis indicates container memory pressure triggering Linux OOM killer termination.\n\n"
                "### 3. Impact Assessment\n"
                "Transient service disruption on affected pod replicas.\n\n"
                "### 4. Recommended Remediation Steps\n"
                "1. Increase container memory limit to prevent future OOM events.\n"
                "2. Restart deployment to restore healthy pod status."
            )
        elif "pod" in last_msg or "status" in last_msg:
            tool_calls.append(
                ToolCallRequest(
                    name="get_pod_status",
                    parameters={"namespace": "default"},
                    requires_approval=False,
                    required_role="viewer",
                )
            )
            content = "Checking pod statuses in the default namespace..."
        elif "restart deployment" in last_msg:
            tool_calls.append(
                ToolCallRequest(
                    name="restart_deployment",
                    parameters={"namespace": "default", "deployment_name": "payment-gateway"},
                    requires_approval=True,
                    required_role="devops",
                )
            )
            content = "I recommend performing a rolling restart of the payment-gateway deployment."
        elif "log" in last_msg:
            tool_calls.append(
                ToolCallRequest(
                    name="get_container_logs",
                    parameters={"container_id": "cifo-payment-service", "tail_lines": 50},
                    requires_approval=False,
                    required_role="viewer",
                )
            )
            content = "Fetching recent container logs for inspection..."
        else:
            content = (
                "CIFO AIOps Assistant is online. All cluster services and Docker containers "
                "are monitored with real-time telemetry."
            )

        in_tokens = sum(len(m.content.split()) for m in messages) * 2
        out_tokens = len(content.split()) * 2

        return ProviderResponse(
            content=content,
            tool_calls=tool_calls,
            model_used=self.model_name,
            provider_name=self.provider_name,
            input_tokens=in_tokens,
            output_tokens=out_tokens,
            estimated_cost_usd=0.0,
        )
