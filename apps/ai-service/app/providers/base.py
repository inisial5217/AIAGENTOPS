from abc import ABC, abstractmethod
from typing import Any
from pydantic import BaseModel, Field
from app.tools.base import ToolDefinition, ToolCallRequest


class ChatMessage(BaseModel):
    role: str
    content: str
    name: str | None = None


class ProviderResponse(BaseModel):
    content: str
    tool_calls: list[ToolCallRequest] = Field(default_factory=list)
    model_used: str
    provider_name: str
    input_tokens: int = 0
    output_tokens: int = 0
    estimated_cost_usd: float = 0.0


class LLMProvider(ABC):
    def __init__(self, provider_name: str, model_name: str) -> None:
        self.provider_name = provider_name
        self.model_name = model_name

    @abstractmethod
    async def chat(
        self,
        messages: list[ChatMessage],
        tools: list[ToolDefinition] | None = None,
        system_instruction: str = "",
    ) -> ProviderResponse:
        # abstract chat call
        pass

    @abstractmethod
    def estimate_cost(self, input_tokens: int, output_tokens: int) -> float:
        # compute cost
        pass
