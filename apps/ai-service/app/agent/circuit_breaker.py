import time
from enum import Enum


class CircuitState(str, Enum):
    CLOSED = "closed"
    OPEN = "open"
    HALF_OPEN = "half_open"


class CircuitBreaker:
    def __init__(
        self,
        failure_threshold: int = 3,
        recovery_time_seconds: int = 120,
        window_seconds: int = 60,
    ) -> None:
        self.failure_threshold = failure_threshold
        self.recovery_time_seconds = recovery_time_seconds
        self.window_seconds = window_seconds

        self.state: CircuitState = CircuitState.CLOSED
        self.failure_timestamps: list[float] = []
        self.last_state_change: float = time.time()

    def can_execute(self) -> bool:
        # evaluate circuit availability
        now = time.time()
        if self.state == CircuitState.CLOSED:
            return True
        if self.state == CircuitState.OPEN:
            if now - self.last_state_change >= self.recovery_time_seconds:
                self.state = CircuitState.HALF_OPEN
                self.last_state_change = now
                return True
            return False
        if self.state == CircuitState.HALF_OPEN:
            return True
        return False

    def record_success(self) -> None:
        # reset on success
        self.state = CircuitState.CLOSED
        self.failure_timestamps.clear()
        self.last_state_change = time.time()

    def record_failure(self) -> None:
        # record failure occurrence
        now = time.time()
        self.failure_timestamps = [
            t for t in self.failure_timestamps if now - t <= self.window_seconds
        ]
        self.failure_timestamps.append(now)

        if len(self.failure_timestamps) >= self.failure_threshold:
            self.state = CircuitState.OPEN
            self.last_state_change = now
