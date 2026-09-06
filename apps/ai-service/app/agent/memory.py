import time
from app.providers.base import ChatMessage


class ConversationMemory:
    def __init__(
        self,
        session_id: str,
        max_messages: int = 20,
        ttl_minutes: int = 30,
    ) -> None:
        self.session_id = session_id
        self.max_messages = max_messages
        self.ttl_seconds = ttl_minutes * 60
        self.messages: list[ChatMessage] = []
        self.summary: str = ""
        self.last_activity: float = time.time()

    def touch(self) -> None:
        # update last activity
        self.last_activity = time.time()

    def is_expired(self) -> bool:
        # verify session ttl
        return (time.time() - self.last_activity) > self.ttl_seconds

    def add_message(self, role: str, content: str) -> None:
        # record new message
        self.messages.append(ChatMessage(role=role, content=content))
        self.touch()

    def get_context(self) -> list[ChatMessage]:
        # return bounded context
        self.touch()
        context_msgs: list[ChatMessage] = []

        if self.summary:
            context_msgs.append(
                ChatMessage(
                    role="system",
                    content=f"Previous conversation summary: {self.summary}",
                )
            )

        # sliding window
        window = self.messages[-self.max_messages :]
        context_msgs.extend(window)
        return context_msgs

    def should_summarize(self) -> bool:
        # check if needs compression
        return len(self.messages) > (self.max_messages + 5)
