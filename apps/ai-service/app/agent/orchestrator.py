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
from app.core.telemetry import get_tracer
from app.tools.base import ToolDefinition

logger = logging.getLogger("cifo.ai.orchestrator")


class ModelOrchestrator:
    def __init__(self) -> None:
        # initialize provider adapters
        self.rehydrate_providers()

        self.circuit_breakers: dict[str, CircuitBreaker] = {
            p.provider_name: CircuitBreaker(
                failure_threshold=settings.failure_threshold,
                recovery_time_seconds=settings.recovery_time_seconds,
            )
            for p in self.providers
        }

        self.active_provider_name: str = self.providers[0].provider_name
        self.model_switch_history: list[dict[str, Any]] = []

    def rehydrate_providers(self) -> None:
        # update provider keys dynamically
        google_key = settings.google_api_key or settings.gemini_api_key
        openai_key = settings.openai_api_key
        anthropic_key = settings.anthropic_api_key
        ollama_url = settings.ollama_base_url

        self.providers: list[LLMProvider] = [
            GoogleGeminiProvider(api_key=google_key, model_name=settings.default_model),
            OpenAIProvider(api_key=openai_key, model_name="gpt-4o-mini"),
            AnthropicProvider(api_key=anthropic_key),
            OllamaProvider(base_url=ollama_url),
            DeterministicMockProvider(),
        ]

    async def chat(
        self,
        messages: list[ChatMessage],
        tools: list[ToolDefinition] | None = None,
        system_instruction: str = "",
        preferred_provider: str = "",
        preferred_model: str = "",
    ) -> ProviderResponse:
        # always ensure fresh credentials
        self.rehydrate_providers()

        # sort providers: preferred first, then providers with keys, then mock
        ordered_providers = list(self.providers)
        if preferred_provider:
            p_match = [p for p in ordered_providers if p.provider_name.lower() == preferred_provider.lower()]
            p_rest = [p for p in ordered_providers if p.provider_name.lower() != preferred_provider.lower()]
            if p_match:
                if preferred_model:
                    p_match[0].model_name = preferred_model
                ordered_providers = p_match + p_rest

        errors: list[str] = []

        for provider in ordered_providers:
            p_name = provider.provider_name
            cb = self.circuit_breakers.get(p_name)

            if cb and not cb.can_execute():
                logger.warning("skipping tripped circuit", extra={"provider": p_name})
                continue

            # skip external providers with empty keys
            if p_name in ("google", "openai", "anthropic"):
                key = getattr(provider, "api_key", "")
                if not key:
                    continue

            tracer = get_tracer()
            try:
                if tracer:
                    with tracer.start_as_current_span(f"llm.{p_name}.generate") as span:
                        span.set_attribute("llm.provider", p_name)
                        span.set_attribute("llm.model", provider.model_name)
                        response = await provider.chat(
                            messages=messages,
                            tools=tools,
                            system_instruction=system_instruction,
                        )
                        span.set_attribute("llm.input_tokens", response.input_tokens)
                        span.set_attribute("llm.output_tokens", response.output_tokens)
                        span.set_attribute("llm.cost_usd", response.estimated_cost_usd)
                else:
                    response = await provider.chat(
                        messages=messages,
                        tools=tools,
                        system_instruction=system_instruction,
                    )

                if cb:
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
                if cb:
                    cb.record_failure()
                err_msg = f"{p_name}: {str(e)}"
                logger.error("provider failed", extra={"provider": p_name, "error": str(e)})
                errors.append(err_msg)

        # degraded mode response
        return ProviderResponse(
            content=(
                "Fitur AI sedang dalam mode terdegradasi karena provider model tidak tersedia. "
                "Silakan periksa konfigurasi API key atau gunakan dashboard pemantauan."
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
                    "circuit_state": self.circuit_breakers[p.provider_name].state.value
                    if p.provider_name in self.circuit_breakers
                    else "closed",
                }
                for p in self.providers
            ],
            "switch_history": self.model_switch_history[-10:],
        }
