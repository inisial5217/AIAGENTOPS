import httpx
from typing import Any
from app.providers.base import LLMProvider, ChatMessage, ProviderResponse
from app.tools.base import ToolDefinition, ToolCallRequest
from app.tools import find_tool


class OllamaProvider(LLMProvider):
    def __init__(
        self,
        base_url: str = "http://localhost:11434",
        model_name: str = "llama3",
    ) -> None:
        super().__init__(provider_name="ollama", model_name=model_name)
        self.base_url = base_url.rstrip("/")

    def estimate_cost(self, input_tokens: int, output_tokens: int) -> float:
        # local inference zero cost
        return 0.0

    async def chat(
        self,
        messages: list[ChatMessage],
        tools: list[ToolDefinition] | None = None,
        system_instruction: str = "",
    ) -> ProviderResponse:
        url = f"{self.base_url}/api/chat"

        formatted_messages: list[dict[str, str]] = []
        if system_instruction:
            formatted_messages.append({"role": "system", "content": system_instruction})

        for msg in messages:
            formatted_messages.append({"role": msg.role, "content": msg.content})

        body: dict[str, Any] = {
            "model": self.model_name,
            "messages": formatted_messages,
            "stream": False,
        }

        async with httpx.AsyncClient(timeout=30.0) as client:
            resp = await client.post(url, json=body)
            if resp.status_code != 200:
                raise RuntimeError(f"Ollama error {resp.status_code}: {resp.text}")
            data = resp.json()

        message = data.get("message", {})
        text_content = message.get("content", "")

        prompt_eval_count = data.get("prompt_eval_count", 0)
        eval_count = data.get("eval_count", 0)

        return ProviderResponse(
            content=text_content,
            tool_calls=[],
            model_used=self.model_name,
            provider_name=self.provider_name,
            input_tokens=prompt_eval_count,
            output_tokens=eval_count,
            estimated_cost_usd=0.0,
        )
