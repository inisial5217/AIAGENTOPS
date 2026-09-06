from app.agent.circuit_breaker import CircuitBreaker, CircuitState


def test_circuit_breaker_trip_and_recover() -> None:
    # verify trip after threshold
    cb = CircuitBreaker(failure_threshold=3, recovery_time_seconds=1, window_seconds=10)
    assert cb.can_execute() is True
    assert cb.state == CircuitState.CLOSED

    cb.record_failure()
    assert cb.state == CircuitState.CLOSED
    cb.record_failure()
    assert cb.state == CircuitState.CLOSED
    cb.record_failure()

    # tripped
    assert cb.state == CircuitState.OPEN
    assert cb.can_execute() is False

    # simulate recovery elapsed
    cb.last_state_change -= 2
    assert cb.can_execute() is True
    assert cb.state == CircuitState.HALF_OPEN

    cb.record_success()
    assert cb.state == CircuitState.CLOSED
