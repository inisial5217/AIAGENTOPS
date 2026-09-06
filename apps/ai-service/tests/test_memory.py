import time
from app.agent.memory import ConversationMemory


def test_conversation_memory_window() -> None:
    # verify max context bounded
    mem = ConversationMemory(session_id="test-session", max_messages=5, ttl_minutes=1)

    for i in range(10):
        mem.add_message(role="user", content=f"Message {i}")

    context = mem.get_context()
    assert len(context) == 5
    assert context[-1].content == "Message 9"


def test_conversation_memory_ttl() -> None:
    # verify ttl calculation
    mem = ConversationMemory(session_id="test-session", max_messages=5, ttl_minutes=1)
    assert mem.is_expired() is False

    mem.last_activity = time.time() - 70
    assert mem.is_expired() is True
