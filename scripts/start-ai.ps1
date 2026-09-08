# Start CIFO AI Service (FastAPI)
$ErrorActionPreference = 'Stop'

$env:HTTP_PORT = "8000"
$env:ENVIRONMENT = "development"
$env:VAULT_ADDR = "http://127.0.0.1:8200"
$env:VAULT_TOKEN = "cifo-vault-root-token"
$env:VAULT_ENABLED = "true"

# Load overrides from root .env if present
$rootEnv = "d:\agent v2\.env"
if (Test-Path $rootEnv) {
    Get-Content $rootEnv | ForEach-Object {
        $line = $_.Trim()
        if ($line -and -not $line.StartsWith("#") -and $line.Contains("=")) {
            $parts = $line.Split("=", 2)
            $k = $parts[0].Trim()
            $v = $parts[1].Trim()
            if ($v -and -not $v.StartsWith("your_")) {
                [System.Environment]::SetEnvironmentVariable($k, $v, "Process")
            }
        }
    }
}

Set-Location "d:\agent v2\apps\ai-service"
Write-Host "Starting CIFO AI Service on port 8000..." -ForegroundColor Cyan

$py = "C:\Users\rin2r\AppData\Local\Programs\Python\Python312\python.exe"
if (Test-Path $py) {
    & $py -m uvicorn app.main:app --host 127.0.0.1 --port 8000 --reload
} else {
    python -m uvicorn app.main:app --host 127.0.0.1 --port 8000 --reload
}
