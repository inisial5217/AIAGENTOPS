import httpx
from typing import Any
from app.providers.base import LLMProvider, ChatMessage, ProviderResponse
from app.tools.base import ToolDefinition, ToolCallRequest
from app.tools import find_tool


class GoogleGeminiProvider(LLMProvider):
    def __init__(
        self,
        api_key: str = "",
        model_name: str = "gemini-2.0-flash",
    ) -> None:
        super().__init__(provider_name="google", model_name=model_name)
        self.api_key = api_key
        self.base_url = "https://generativelanguage.googleapis.com/v1beta/models"

    def estimate_cost(self, input_tokens: int, output_tokens: int) -> float:
        # gemini flash rate
        input_cost = (input_tokens / 1_000_000.0) * 0.075
        output_cost = (output_tokens / 1_000_000.0) * 0.30
        return round(input_cost + output_cost, 6)

    async def chat(
        self,
        messages: list[ChatMessage],
        tools: list[ToolDefinition] | None = None,
        system_instruction: str = "",
    ) -> ProviderResponse:
        # validate api key
        if not self.api_key:
            raise ValueError("Google API key not configured")

        url = f"{self.base_url}/{self.model_name}:generateContent?key={self.api_key}"

        # format contents payload
        contents: list[dict[str, Any]] = []
        for msg in messages:
            role = "model" if msg.role == "assistant" else "user"
            contents.append({"role": role, "parts": [{"text": msg.content}]})

        body: dict[str, Any] = {"contents": contents}

        if system_instruction:
            body["system_instruction"] = {"parts": [{"text": system_instruction}]}

        if tools:
            function_declarations = []
            for t in tools:
                function_declarations.append({
                    "name": t.name,
                    "description": t.description,
                    "parameters": t.parameters,
                })
            body["tools"] = [{"function_declarations": function_declarations}]

        async with httpx.AsyncClient(timeout=30.0) as client:
            resp = await client.post(url, json=body)
            if resp.status_code != 200:
                raise RuntimeError(f"Gemini API error {resp.status_code}: {resp.text}")
            data = resp.json()

        # extract output
        candidates = data.get("candidates", [])
        if not candidates:
            raise RuntimeError("Gemini returned empty candidates")

        first_candidate = candidates[0]
        content_parts = first_candidate.get("content", {}).get("parts", [])

        text_content = ""
        tool_calls: list[ToolCallRequest] = []

        for part in content_parts:
            if "text" in part:
                text_content += part["text"]
            if "functionCall" in part:
                fc = part["functionCall"]
                name = fc.get("name", "")
                args = fc.get("args", {})
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

        usage = data.get("usageMetadata", {})
        in_tokens = usage.get("promptTokenCount", 0)
        out_tokens = usage.get("candidatesTokenCount", 0)

        return ProviderResponse(
            content=text_content,
            tool_calls=tool_calls,
            model_used=self.model_name,
            provider_name=self.provider_name,
            input_tokens=in_tokens,
            output_tokens=out_tokens,
            estimated_cost_usd=self.estimate_cost(in_tokens, out_tokens),
        )
