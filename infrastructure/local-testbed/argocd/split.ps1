# Split install.yaml into crds.yaml and components.yaml
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$installFile = Join-Path $scriptDir "install.yaml"
$crdsFile = Join-Path $scriptDir "crds.yaml"
$compFile = Join-Path $scriptDir "components.yaml"

$lines = [System.IO.File]::ReadAllLines($installFile)
Write-Host "Total lines: $($lines.Length)"

# Line 20511 is the separator before kind: ServiceAccount
$splitIndex = 20510

$crdLines = $lines[0..($splitIndex - 1)]
$compLines = $lines[$splitIndex..($lines.Length - 1)]

[System.IO.File]::WriteAllLines($crdsFile, $crdLines)
[System.IO.File]::WriteAllLines($compFile, $compLines)

Write-Host "Written crds.yaml: $($crdLines.Length) lines" -ForegroundColor Green
Write-Host "Written components.yaml: $($compLines.Length) lines" -ForegroundColor Green
