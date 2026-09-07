import pytest
from unittest.mock import MagicMock, patch
from app.config.settings import Settings
from app.core.vault import get_vault_secrets, apply_vault_secrets_to_settings


def test_get_vault_secrets_success():
    mock_resp = MagicMock()
    mock_resp.status_code = 200
    mock_resp.json.return_value = {
        "data": {
            "data": {
                "GOOGLE_API_KEY": "mocked-gemini-key",
                "OPENAI_API_KEY": "mocked-openai-key",
            }
        }
    }

    with patch("httpx.Client.get", return_value=mock_resp):
        secrets = get_vault_secrets("cifo/ai-service")
        assert secrets.get("GOOGLE_API_KEY") == "mocked-gemini-key"
        assert secrets.get("OPENAI_API_KEY") == "mocked-openai-key"


def test_get_vault_secrets_failure_fallback():
    with patch("httpx.Client.get", side_effect=Exception("Connection refused")):
        secrets = get_vault_secrets("cifo/ai-service")
        assert secrets == {}


def test_apply_vault_secrets_to_settings():
    test_settings = Settings()
    mock_secrets = {
        "GOOGLE_API_KEY": "vault-key-test",
        "OPENAI_API_KEY": "vault-openai-test",
        "ANTHROPIC_API_KEY": "vault-claude-test",
        "OLLAMA_BASE_URL": "http://vault-host:11434",
    }

    with patch("app.core.vault.get_vault_secrets", return_value=mock_secrets):
        apply_vault_secrets_to_settings(test_settings)
        assert test_settings.google_api_key == "vault-key-test"
        assert test_settings.openai_api_key == "vault-openai-test"
        assert test_settings.anthropic_api_key == "vault-claude-test"
        assert test_settings.ollama_base_url == "http://vault-host:11434"
