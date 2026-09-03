from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    # server config
    app_name: str = "cifo-ai-service"
    environment: str = "development"
    http_port: int = 8000
    grpc_port: int = 50051

    # model keys
    google_api_key: str = ""
    openai_api_key: str = ""
    anthropic_api_key: str = ""
    ollama_base_url: str = "http://localhost:11434"

    # circuit breaker
    failure_threshold: int = 3
    recovery_time_seconds: int = 120

    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")


settings = Settings()
