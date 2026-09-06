# Test Phase 7 WebSocket Handshake & Auth
$ErrorActionPreference = "Continue"

Write-Host "=== TEST 1: WebSocket endpoint without token ===" -ForegroundColor Cyan
try {
    $resp = Invoke-WebRequest -Uri "http://127.0.0.1:8080/ws" -Method Get -ErrorAction Stop
    Write-Host "FAIL: Expected 401 but got $($resp.StatusCode)" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Host "PASS: Rejected with 401 Unauthorized" -ForegroundColor Green
    } else {
        Write-Host "FAIL: Unexpected status $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    }
}

Write-Host "`n=== TEST 2: WebSocket endpoint with invalid token ===" -ForegroundColor Cyan
try {
    $resp = Invoke-WebRequest -Uri "http://127.0.0.1:8080/ws?token=invalid-garbage-token" -Method Get -ErrorAction Stop
    Write-Host "FAIL: Expected 401 but got $($resp.StatusCode)" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Host "PASS: Rejected with 401 Unauthorized" -ForegroundColor Green
    } else {
        Write-Host "FAIL: Unexpected status $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    }
}

Write-Host "`n=== TEST 3: WebSocket client handshake with valid token ===" -ForegroundColor Cyan
$wsClient = New-Object System.Net.WebSockets.ClientWebSocket
$cts = New-Object System.Threading.CancellationTokenSource
$cts.CancelAfter(5000)

$uri = New-Object System.Uri("ws://127.0.0.1:8080/ws?token=dev-token-admin")
try {
    $connectTask = $wsClient.ConnectAsync($uri, $cts.Token)
    $connectTask.Wait()

    if ($wsClient.State -eq [System.Net.WebSockets.WebSocketState]::Open) {
        Write-Host "PASS: WebSocket connection OPEN and authenticated!" -ForegroundColor Green

        # Send subscribe to system_events
        $subMsg = '{"action":"subscribe","topic":"system_events"}'
        $bytes = [System.Text.Encoding]::UTF8.GetBytes($subMsg)
        $segment = New-Object System.ArraySegment[byte]($bytes, 0, $bytes.Length)
        $sendTask = $wsClient.SendAsync($segment, [System.Net.WebSockets.WebSocketMessageType]::Text, $true, $cts.Token)
        $sendTask.Wait()
        Write-Host "PASS: Sent subscribe action to topic 'system_events'" -ForegroundColor Green

        # Receive ack
        $buffer = New-Object byte[] 4096
        $recvSegment = New-Object System.ArraySegment[byte]($buffer, 0, $buffer.Length)
        $recvTask = $wsClient.ReceiveAsync($recvSegment, $cts.Token)
        $recvTask.Wait()
        $recvBytes = $recvTask.Result.Count
        $receivedText = [System.Text.Encoding]::UTF8.GetString($buffer, 0, $recvBytes)
        Write-Host "PASS: Received WebSocket frame from server: $receivedText" -ForegroundColor Green

        # Close
        $closeTask = $wsClient.CloseAsync([System.Net.WebSockets.WebSocketCloseStatus]::NormalClosure, "Test complete", $cts.Token)
        $closeTask.Wait()
        Write-Host "PASS: Clean disconnect verified" -ForegroundColor Green
    } else {
        Write-Host "FAIL: WebSocket state is $($wsClient.State)" -ForegroundColor Red
    }
} catch {
    Write-Host "FAIL: Exception connecting WebSocket: $($_.Exception.Message)" -ForegroundColor Red
} finally {
    $wsClient.Dispose()
    $cts.Dispose()
}
