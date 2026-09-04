# Test TCP connection to Postgres and Redis
$tcpPg = New-Object System.Net.Sockets.TcpClient
try {
    $tcpPg.Connect('127.0.0.1', 5432)
    Write-Host "Postgres 127.0.0.1:5432 is OPEN: $($tcpPg.Connected)"
} catch {
    Write-Host "Postgres connection error: $_"
} finally {
    $tcpPg.Close()
}

$tcpRedis = New-Object System.Net.Sockets.TcpClient
try {
    $tcpRedis.Connect('127.0.0.1', 6379)
    Write-Host "Redis 127.0.0.1:6379 is OPEN: $($tcpRedis.Connected)"
} catch {
    Write-Host "Redis connection error: $_"
} finally {
    $tcpRedis.Close()
}
