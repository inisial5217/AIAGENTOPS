import logging
import os
from typing import Any
import httpx
from app.config.settings import Settings

logger = logging.getLogger("cifo-ai-service.vault")


def get_vault_secrets(
    path: str = "cifo/ai-service",
    vault_addr: str | None = None,
    vault_token: str | None = None,
) -> dict[str, Any]:
    """Fetch KV v2 secrets from HashiCorp Vault with resilient error handling."""
    addr = (vault_addr or os.environ.get("VAULT_ADDR", "http://127.0.0.1:8200")).rstrip("/")
    token = vault_token or os.environ.get("VAULT_TOKEN", "cifo-vault-root-token")

    clean_path = path.lstrip("/")
    if not clean_path.startswith("secret/data/"):
        if clean_path.startswith("secret/"):
            clean_path = clean_path.replace("secret/", "secret/data/", 1)
        else:
            clean_path = f"secret/data/{clean_path}"

    url = f"{addr}/v1/{clean_path}"
    headers = {"X-Vault-Token": token}

    try:
        with httpx.Client(timeout=3.0) as client:
            resp = client.get(url, headers=headers)
            if resp.status_code == 200:
                body = resp.json()
                data = body.get("data", {}).get("data", {})
                logger.info(f"Loaded {len(data)} secrets from Vault at {path}")
                return data
            else:
                logger.warning(
                    f"Vault returned status {resp.status_code} for path {path}: {resp.text}"
                )
                return {}
    except Exception as exc:
        logger.warning(f"Could not connect to Vault at {addr} ({exc}). Using environment fallback.")
        return {}


def apply_vault_secrets_to_settings(target_settings: Settings) -> None:
    """Read secrets from Vault and apply to settings with fallback."""
    vault_enabled = os.environ.get("VAULT_ENABLED", "true").lower() in ("true", "1", "yes")
    if not vault_enabled:
        logger.info("Vault integration disabled via VAULT_ENABLED=false")
        return

    secrets = get_vault_secrets("cifo/ai-service")
    if not secrets:
        return

    if "GOOGLE_API_KEY" in secrets and secrets["GOOGLE_API_KEY"]:
        target_settings.google_api_key = str(secrets["GOOGLE_API_KEY"])
    if "OPENAI_API_KEY" in secrets and secrets["OPENAI_API_KEY"]:
        target_settings.openai_api_key = str(secrets["OPENAI_API_KEY"])
    if "ANTHROPIC_API_KEY" in secrets and secrets["ANTHROPIC_API_KEY"]:
        target_settings.anthropic_api_key = str(secrets["ANTHROPIC_API_KEY"])
    if "OLLAMA_BASE_URL" in secrets and secrets["OLLAMA_BASE_URL"]:
        target_settings.ollama_base_url = str(secrets["OLLAMA_BASE_URL"])
    logger.info("Applied Vault credentials to AI service runtime configuration")
