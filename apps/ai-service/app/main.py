import os
from typing import Any
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from app.config.settings import settings
from app.agent.orchestrator import ModelOrchestrator
from app.agent.sanitizer import PromptSanitizer
from app.agent.memory import ConversationMemory
from app.providers.base import ChatMessage, ProviderResponse
from app.tools import ALL_TOOLS, get_tool_schemas

app = FastAPI(
    title="CIFO AIOps Service",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
)

orchestrator = ModelOrchestrator()
session_memories: dict[str, ConversationMemory] = {}


def load_prompt(filename: str) -> str:
    # read prompt file
    base_dir = os.path.dirname(__file__)
    prompt_path = os.path.join(base_dir, "prompts", filename)
    if os.path.exists(prompt_path):
        with open(prompt_path, "r", encoding="utf-8") as f:
            return f.read()
    return ""


SYSTEM_PROMPT = load_prompt("system_prompt.txt")
DIAGNOSIS_PROMPT = load_prompt("diagnosis_prompt.txt")


class ChatRequest(BaseModel):
    session_id: str
    user_id: str
    message: str
    user_role: str = "viewer"
    history: list[dict[str, str]] = Field(default_factory=list)


class ChatResponse(BaseModel):
    session_id: str
    content: str
    tool_calls: list[dict[str, Any]] = Field(default_factory=list)
    model_used: str
    provider_name: str
    input_tokens: int = 0
    output_tokens: int = 0
    estimated_cost_usd: float = 0.0
    security_flag: str | None = None


class DiagnoseRequest(BaseModel):
    incident_id: str
    alert_name: str
    severity: str
    resource: str
    namespace: str = "default"
    logs: str = ""
    metrics: dict[str, Any] = Field(default_factory=dict)


class DiagnoseResponse(BaseModel):
    incident_id: str
    rca_summary: str
    model_used: str
    provider_name: str
    input_tokens: int = 0
    output_tokens: int = 0
    estimated_cost_usd: float = 0.0


@app.get("/healthz")
async def health_check() -> dict[str, str]:
    # process health status
    return {"status": "ok", "service": settings.app_name}


@app.get("/readyz")
async def ready_check() -> dict[str, str]:
    # dependency readiness check
    return {"status": "ready", "environment": settings.environment}


@app.get("/api/v1/tools")
async def list_tools() -> list[dict[str, Any]]:
    # list registered tools
    return [
        {
            "name": t.name,
            "description": t.description,
            "parameters": t.parameters,
            "required_role": t.required_role,
            "requires_approval": t.requires_approval,
        }
        for t in ALL_TOOLS
    ]


@app.get("/api/v1/models")
async def model_status() -> dict[str, Any]:
    # get active models
    return orchestrator.get_status()


@app.post("/api/v1/chat", response_model=ChatResponse)
async def chat_endpoint(req: ChatRequest) -> ChatResponse:
    # handle chat message
    clean_message = PromptSanitizer.sanitize_input(req.message)

    # detect prompt injection
    is_injection, injection_type = PromptSanitizer.detect_injection(clean_message)
    if is_injection:
        return ChatResponse(
            session_id=req.session_id,
            content=(
                f"Peringatan Keamanan: Perintah Anda mengandung pola terlarang ('{injection_type}') "
                "dan dibatalkan demi keamanan cluster."
            ),
            tool_calls=[],
            model_used="security-filter",
            provider_name="system",
            security_flag=injection_type,
        )

    # retrieve or initialize session memory
    if req.session_id not in session_memories:
        session_memories[req.session_id] = ConversationMemory(
            session_id=req.session_id,
            max_messages=settings.max_context_messages,
            ttl_minutes=settings.session_ttl_minutes,
        )

    memory = session_memories[req.session_id]

    # hydrate history if memory empty
    if not memory.messages and req.history:
        for item in req.history:
            memory.add_message(role=item.get("role", "user"), content=item.get("content", ""))

    memory.add_message(role="user", content=clean_message)
    context_msgs = memory.get_context()

    # route through multi-model orchestrator
    response: ProviderResponse = await orchestrator.chat(
        messages=context_msgs,
        tools=ALL_TOOLS,
        system_instruction=SYSTEM_PROMPT,
    )

    # validate tool calls
    valid_tool_calls = PromptSanitizer.validate_tool_calls(response.tool_calls)

    # record assistant response in memory
    memory.add_message(role="assistant", content=response.content)

    return ChatResponse(
        session_id=req.session_id,
        content=response.content,
        tool_calls=[tc.model_dump() for tc in valid_tool_calls],
        model_used=response.model_used,
        provider_name=response.provider_name,
        input_tokens=response.input_tokens,
        output_tokens=response.output_tokens,
        estimated_cost_usd=response.estimated_cost_usd,
        security_flag=None,
    )


@app.post("/api/v1/diagnose", response_model=DiagnoseResponse)
async def diagnose_endpoint(req: DiagnoseRequest) -> DiagnoseResponse:
    # generate automated rca
    prompt_text = DIAGNOSIS_PROMPT.format(
        alert_name=req.alert_name,
        resource=req.resource,
        namespace=req.namespace,
        severity=req.severity,
    )

    full_query = f"{prompt_text}\n\nRecent Telemetry and Logs:\n{req.logs if req.logs else 'No raw logs attached.'}"
    diag_messages = [ChatMessage(role="user", content=full_query)]

    response = await orchestrator.chat(
        messages=diag_messages,
        tools=None,
        system_instruction=SYSTEM_PROMPT,
    )

    return DiagnoseResponse(
        incident_id=req.incident_id,
        rca_summary=response.content,
        model_used=response.model_used,
        provider_name=response.provider_name,
        input_tokens=response.input_tokens,
        output_tokens=response.output_tokens,
        estimated_cost_usd=response.estimated_cost_usd,
    )
