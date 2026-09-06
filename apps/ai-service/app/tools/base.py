from typing import Any
from pydantic import BaseModel, Field


class ToolDefinition(BaseModel):
    name: str
    description: str
    parameters: dict[str, Any]
    required_role: str = "viewer"
    requires_approval: bool = False

    def to_schema(self) -> dict[str, Any]:
        # return json schema
        return {
            "type": "function",
            "function": {
                "name": self.name,
                "description": self.description,
                "parameters": self.parameters,
            },
        }


class ToolCallRequest(BaseModel):
    name: str
    parameters: dict[str, Any] = Field(default_factory=dict)
    requires_approval: bool = False
    required_role: str = "viewer"
