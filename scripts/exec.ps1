# Execute command with refreshed PATH
$m = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
$u = [System.Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$m;$u"
$cmd = $args -join " "
Invoke-Expression $cmd
