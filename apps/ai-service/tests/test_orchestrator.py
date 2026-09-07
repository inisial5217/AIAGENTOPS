import asyncio
import pytest
from unittest.mock import AsyncMock
from app.agent.orchestrator import ModelOrchestrator
from app.agent.circuit_breaker import CircuitBreaker, CircuitState
from app.providers.base import ChatMessage, ProviderResponse, LLMProvider
from app.providers.mock_provider import DeterministicMockProvider


def test_orchestrator_fallback_to_mock() -> None:
    # initialize orchestrator
    orchestrator = ModelOrchestrator()

    # simulate real LLM provider failures so it falls back to DeterministicMockProvider
    for p in orchestrator.providers[:-1]:
        p.chat = AsyncMock(side_effect=Exception("API unavailable"))

    messages = [ChatMessage(role="user", content="Test health")]
    response = asyncio.run(orchestrator.chat(messages))

    assert response is not None
    assert response.provider_name == "mock"
    assert response.model_used == "cifo-deterministic-mock"
    assert orchestrator.active_provider_name == "mock"
    assert len(orchestrator.model_switch_history) > 0


def test_orchestrator_degraded_mode_when_all_fail() -> None:
    orchestrator = ModelOrchestrator()

    # make all providers fail, including mock
    for p in orchestrator.providers:
        p.chat = AsyncMock(side_effect=Exception("Complete outage"))

    messages = [ChatMessage(role="user", content="Hello")]
    response = asyncio.run(orchestrator.chat(messages))

    assert response is not None
    assert response.provider_name == "system"
    assert "terdegradasi" in response.content


def test_circuit_breaker_state_transitions() -> None:
    cb = CircuitBreaker(failure_threshold=2, recovery_time_seconds=1, window_seconds=60)
    assert cb.state == CircuitState.CLOSED
    assert cb.can_execute() is True

    # 1 failure -> still closed
    cb.record_failure()
    assert cb.state == CircuitState.CLOSED
    assert cb.can_execute() is True

    # 2 failures -> trips to OPEN
    cb.record_failure()
    assert cb.state == CircuitState.OPEN
    assert cb.can_execute() is False

    # simulate recovery time passage
    cb.last_state_change = cb.last_state_change - 2
    assert cb.can_execute() is True
    assert cb.state == CircuitState.HALF_OPEN

    # success in half-open resets to CLOSED
    cb.record_success()
    assert cb.state == CircuitState.CLOSED
    assert len(cb.failure_timestamps) == 0
