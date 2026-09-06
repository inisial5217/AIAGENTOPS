# Test monitoring endpoints
$ErrorActionPreference = "Stop"

$ready = Invoke-RestMethod -Uri "http://127.0.0.1:8080/readyz"
Write-Host "Readyz:" ($ready | ConvertTo-Json -Compress)

$login = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/auth/login" -Method Post -ContentType "application/json" -Body '{"username":"admin@cifo.local","password":"admin123"}'
$token = $login.access_token
Write-Host "Login Token Received: " ($token -ne $null)

$headers = @{ Authorization = "Bearer $token" }

$stats = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/monitoring/stats" -Headers $headers
Write-Host "Stats:" ($stats | ConvertTo-Json -Compress)

$cpu = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/monitoring/metrics/cpu" -Headers $headers
Write-Host "CPU Points count:" $cpu.data.Count
Write-Host "CPU[0]:" ($cpu.data[0] | ConvertTo-Json -Compress)

$mem = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/monitoring/metrics/memory" -Headers $headers
Write-Host "Mem Points count:" $mem.data.Count
Write-Host "Mem[0]:" ($mem.data[0] | ConvertTo-Json -Compress)

$net = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/monitoring/metrics/network" -Headers $headers
Write-Host "Net Points count:" $net.data.Count
Write-Host "Net[0]:" ($net.data[0] | ConvertTo-Json -Compress)
