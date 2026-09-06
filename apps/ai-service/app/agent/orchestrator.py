import logging
from typing import Any
from app.config.settings import settings
from app.providers.base import LLMProvider, ChatMessage, ProviderResponse
from app.providers.google_provider import GoogleGeminiProvider
from app.providers.openai_provider import OpenAIProvider
from app.providers.anthropic_provider import AnthropicProvider
from app.providers.ollama_provider import OllamaProvider
from app.providers.mock_provider import DeterministicMockProvider
from app.agent.circuit_breaker import CircuitBreaker, CircuitState
from app.tools.base import ToolDefinition

logger = logging.getLogger("cifo.ai.orchestrator")


class ModelOrchestrator:
    def __init__(self) -> None:
        # initialize provider adapters
        self.providers: list[LLMProvider] = [
            GoogleGeminiProvider(api_key=settings.google_api_key),
            OpenAIProvider(api_key=settings.openai_api_key),
            AnthropicProvider(api_key=settings.anthropic_api_key),
            OllamaProvider(base_url=settings.ollama_base_url),
            DeterministicMockProvider(),
        ]

        self.circuit_breakers: dict[str, CircuitBreaker] = {
            p.provider_name: CircuitBreaker(
                failure_threshold=settings.failure_threshold,
                recovery_time_seconds=settings.recovery_time_seconds,
            )
            for p in self.providers
        }

        self.active_provider_name: str = self.providers[0].provider_name
        self.model_switch_history: list[dict[str, Any]] = []

    async def chat(
        self,
        messages: list[ChatMessage],
        tools: list[ToolDefinition] | None = None,
        system_instruction: str = "",
    ) -> ProviderResponse:
        # route across providers
        errors: list[str] = []

        for provider in self.providers:
            p_name = provider.provider_name
            cb = self.circuit_breakers[p_name]

            if not cb.can_execute():
                logger.warning("skipping tripped circuit", extra={"provider": p_name})
                continue

            try:
                response = await provider.chat(
                    messages=messages,
                    tools=tools,
                    system_instruction=system_instruction,
                )
                cb.record_success()

                if self.active_provider_name != p_name:
                    logger.info(
                        "model failover switch",
                        extra={"from": self.active_provider_name, "to": p_name},
                    )
                    self.model_switch_history.append({
                        "from": self.active_provider_name,
                        "to": p_name,
                        "model": provider.model_name,
                    })
                    self.active_provider_name = p_name

                return response
            except Exception as e:
                cb.record_failure()
                err_msg = f"{p_name}: {str(e)}"
                logger.error("provider failed", extra={"provider": p_name, "error": str(e)})
                errors.append(err_msg)

        # degraded mode response
        return ProviderResponse(
            content=(
                "Fitur AI sedang dalam mode terdegradasi karena provider model tidak tersedia. "
                "Silakan gunakan dashboard manual untuk pemantauan dan operasional."
            ),
            tool_calls=[],
            model_used="degraded-mode",
            provider_name="system",
            input_tokens=0,
            output_tokens=0,
            estimated_cost_usd=0.0,
        )

    def get_status(self) -> dict[str, Any]:
        # return active status
        return {
            "active_provider": self.active_provider_name,
            "providers": [
                {
                    "name": p.provider_name,
                    "model": p.model_name,
                    "circuit_state": self.circuit_breakers[p.provider_name].state.value,
                }
                for p in self.providers
            ],
            "switch_history": self.model_switch_history[-10:],
        }
