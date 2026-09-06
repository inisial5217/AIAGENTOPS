import json
import httpx
from typing import Any
from app.providers.base import LLMProvider, ChatMessage, ProviderResponse
from app.tools.base import ToolDefinition, ToolCallRequest
from app.tools import find_tool


class OpenAIProvider(LLMProvider):
    def __init__(
        self,
        api_key: str = "",
        model_name: str = "gpt-4o",
    ) -> None:
        super().__init__(provider_name="openai", model_name=model_name)
        self.api_key = api_key
        self.endpoint = "https://api.openai.com/v1/chat/completions"

    def estimate_cost(self, input_tokens: int, output_tokens: int) -> float:
        # gpt-4o rate
        input_cost = (input_tokens / 1_000_000.0) * 2.50
        output_cost = (output_tokens / 1_000_000.0) * 10.00
        return round(input_cost + output_cost, 6)

    async def chat(
        self,
        messages: list[ChatMessage],
        tools: list[ToolDefinition] | None = None,
        system_instruction: str = "",
    ) -> ProviderResponse:
        # validate api key
        if not self.api_key:
            raise ValueError("OpenAI API key not configured")

        formatted_messages: list[dict[str, str]] = []
        if system_instruction:
            formatted_messages.append({"role": "system", "content": system_instruction})

        for msg in messages:
            formatted_messages.append({"role": msg.role, "content": msg.content})

        body: dict[str, Any] = {
            "model": self.model_name,
            "messages": formatted_messages,
        }

        if tools:
            body["tools"] = [t.to_schema() for t in tools]
            body["tool_choice"] = "auto"

        headers = {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json",
        }

        async with httpx.AsyncClient(timeout=30.0) as client:
            resp = await client.post(self.endpoint, headers=headers, json=body)
            if resp.status_code != 200:
                raise RuntimeError(f"OpenAI API error {resp.status_code}: {resp.text}")
            data = resp.json()

        choice = data.get("choices", [{}])[0]
        message = choice.get("message", {})
        text_content = message.get("content") or ""

        tool_calls: list[ToolCallRequest] = []
        raw_tool_calls = message.get("tool_calls", [])
        for tc in raw_tool_calls:
            func = tc.get("function", {})
            name = func.get("name", "")
            args_str = func.get("arguments", "{}")
            try:
                args = json.loads(args_str)
            except Exception:
                args = {}

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
        in_tokens = usage.get("prompt_tokens", 0)
        out_tokens = usage.get("completion_tokens", 0)

        return ProviderResponse(
            content=text_content,
            tool_calls=tool_calls,
            model_used=self.model_name,
            provider_name=self.provider_name,
            input_tokens=in_tokens,
            output_tokens=out_tokens,
            estimated_cost_usd=self.estimate_cost(in_tokens, out_tokens),
        )
