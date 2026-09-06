import httpx
from typing import Any
from app.providers.base import LLMProvider, ChatMessage, ProviderResponse
from app.tools.base import ToolDefinition, ToolCallRequest
from app.tools import find_tool


class AnthropicProvider(LLMProvider):
    def __init__(
        self,
        api_key: str = "",
        model_name: str = "claude-3-5-sonnet-20241022",
    ) -> None:
        super().__init__(provider_name="anthropic", model_name=model_name)
        self.api_key = api_key
        self.endpoint = "https://api.anthropic.com/v1/messages"

    def estimate_cost(self, input_tokens: int, output_tokens: int) -> float:
        # claude sonnet rate
        input_cost = (input_tokens / 1_000_000.0) * 3.00
        output_cost = (output_tokens / 1_000_000.0) * 15.00
        return round(input_cost + output_cost, 6)

    async def chat(
        self,
        messages: list[ChatMessage],
        tools: list[ToolDefinition] | None = None,
        system_instruction: str = "",
    ) -> ProviderResponse:
        # validate api key
        if not self.api_key:
            raise ValueError("Anthropic API key not configured")

        formatted_messages: list[dict[str, str]] = []
        for msg in messages:
            role = "assistant" if msg.role == "assistant" else "user"
            formatted_messages.append({"role": role, "content": msg.content})

        body: dict[str, Any] = {
            "model": self.model_name,
            "max_tokens": 2048,
            "messages": formatted_messages,
        }

        if system_instruction:
            body["system"] = system_instruction

        if tools:
            body["tools"] = [
                {
                    "name": t.name,
                    "description": t.description,
                    "input_schema": t.parameters,
                }
                for t in tools
            ]

        headers = {
            "x-api-key": self.api_key,
            "anthropic-version": "2023-06-01",
            "Content-Type": "application/json",
        }

        async with httpx.AsyncClient(timeout=30.0) as client:
            resp = await client.post(self.endpoint, headers=headers, json=body)
            if resp.status_code != 200:
                raise RuntimeError(f"Anthropic API error {resp.status_code}: {resp.text}")
            data = resp.json()

        text_content = ""
        tool_calls: list[ToolCallRequest] = []

        for block in data.get("content", []):
            if block.get("type") == "text":
                text_content += block.get("text", "")
            elif block.get("type") == "tool_use":
                name = block.get("name", "")
                args = block.get("input", {})
                tool_def = find_tool(name)
                req_approval = tool_def.requires_approval if tool_def else False
                req_role = tool_def.required_role if tool_def else "viewer"
                tool_calls.append(
                    ToolCallRequest(
                        name=name,
                        parameters=args,
                        requires_approval=req_approval,
                        required_role=req_role,
                    )
                )

        usage = data.get("usage", {})
        in_tokens = usage.get("input_tokens", 0)
        out_tokens = usage.get("output_tokens", 0)

        return ProviderResponse(
            content=text_content,
            tool_calls=tool_calls,
            model_used=self.model_name,
            provider_name=self.provider_name,
            input_tokens=in_tokens,
            output_tokens=out_tokens,
            estimated_cost_usd=self.estimate_cost(in_tokens, out_tokens),
        )
