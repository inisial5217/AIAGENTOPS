# Test rate limit
$hit429 = $false
for ($i = 1; $i -le 110; $i++) {
    $code = curl.exe -s -w "%{http_code}" -o "d:\agent v2\scratch\ratelimit_resp.json" http://127.0.0.1:8080/healthz
    if ($code -eq "429") {
        Write-Host "Hit 429 Too Many Requests on request #$i"
        Get-Content "d:\agent v2\scratch\ratelimit_resp.json"
        $hit429 = $true
        break
    }
}
if (-not $hit429) {
    Write-Host "Did not hit 429 within 110 requests"
}
