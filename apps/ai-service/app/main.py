from fastapi import FastAPI
from app.config.settings import settings

app = FastAPI(
    title="CIFO AIOps Service",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
)


@app.get("/healthz")
async def health_check() -> dict[str, str]:
    # process health status
    return {"status": "ok", "service": settings.app_name}


@app.get("/readyz")
async def ready_check() -> dict[str, str]:
    # dependency readiness check
    return {"status": "ready", "environment": settings.environment}
