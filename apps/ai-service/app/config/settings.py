from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    # server config
    app_name: str = "cifo-ai-service"
    environment: str = "development"
    http_port: int = 8000
    grpc_port: int = 50051
    backend_api_url: str = "http://localhost:8080/api/v1"

    # model keys
    google_api_key: str = ""
    openai_api_key: str = ""
    anthropic_api_key: str = ""
    ollama_base_url: str = "http://localhost:11434"

    # model defaults
    default_model: str = "gemini-2.0-flash"
    fallback_models: list[str] = ["gpt-4o", "claude-3-5-sonnet", "ollama-llama3"]
    model_timeout_seconds: int = 30

    # circuit breaker
    failure_threshold: int = 3
    recovery_time_seconds: int = 120

    # memory & context
    max_context_messages: int = 20
    session_ttl_minutes: int = 30

    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")


settings = Settings()
