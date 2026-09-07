$env:VAULT_ADDR = "http://127.0.0.1:8200"
$env:VAULT_TOKEN = "cifo-vault-root-token"
$env:VAULT_ENABLED = "true"
Set-Location "d:\agent v2\apps\ai-service"
Write-Host "Running AI service unit tests with pytest..."
& "C:\Users\rin2r\AppData\Local\Programs\Python\Python312\python.exe" -m pytest -v tests/
