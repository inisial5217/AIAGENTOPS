import re
from typing import Any
from app.tools.base import ToolCallRequest
from app.tools import find_tool

DANGEROUS_PATTERNS: list[tuple[str, re.Pattern[str]]] = [
    ("prompt_override", re.compile(r"ignore\s+(?:all\s+)?previous\s+instructions", re.IGNORECASE)),
    ("forget_instructions", re.compile(r"forget\s+(?:all\s+)?your\s+instructions", re.IGNORECASE)),
    ("system_override", re.compile(r"system\s+override", re.IGNORECASE)),
    ("privilege_escalation", re.compile(r"(?:act\s+as|you\s+are\s+now)\s+root", re.IGNORECASE)),
    ("destructive_delete_all", re.compile(r"kubectl\s+delete\s+(?:namespace|node|all)", re.IGNORECASE)),
    ("destructive_prune", re.compile(r"docker\s+system\s+prune", re.IGNORECASE)),
    ("filesystem_wipe", re.compile(r"rm\s+-rf\s+[/~]", re.IGNORECASE)),
]


class PromptSanitizer:
    @staticmethod
    def detect_injection(text: str) -> tuple[bool, str | None]:
        # detect prompt injection
        for label, pattern in DANGEROUS_PATTERNS:
            if pattern.search(text):
                return True, label
        return False, None

    @staticmethod
    def sanitize_input(text: str) -> str:
        # strip raw control sequences
        sanitized = re.sub(r"[\x00-\x08\x0B\x0C\x0E-\x1F]", "", text)
        return sanitized.strip()

    @staticmethod
    def validate_tool_calls(tool_calls: list[ToolCallRequest]) -> list[ToolCallRequest]:
        # ensure tool is in allowlist
        valid_calls: list[ToolCallRequest] = []
        for tc in tool_calls:
            tool_def = find_tool(tc.name)
            if tool_def is None:
                # rejected unlisted tool
                continue

            # enforce canonical role and approval
            tc.requires_approval = tool_def.requires_approval
            tc.required_role = tool_def.required_role
            valid_calls.append(tc)
        return valid_calls
