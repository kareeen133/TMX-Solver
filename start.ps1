param(
  [switch]$Ngrok,
  [string]$Addr = ":8080",
  [string]$ApiKey = "",
  [string]$AdminUser = "admin",
  [string]$AdminPass = "anees3232@",
  [string]$Proxy = ""
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$port = ($Addr -replace '^.*:', '')
if (-not $port) { $port = "8080" }

$serverArgs = @("run", "-buildvcs=false", ".", "-serve", "-addr", $Addr, "-admin-user", $AdminUser, "-admin-pass", $AdminPass)
if ($ApiKey) { $serverArgs += @("-api-key", $ApiKey) }
if ($Proxy)  { $serverArgs += @("-proxy", $Proxy) }

Write-Host "[start] launching TMX server: go $($serverArgs -join ' ')" -ForegroundColor Cyan
$server = Start-Process -FilePath "go" -ArgumentList $serverArgs -PassThru -NoNewWindow

if ($Ngrok) {
  Start-Sleep -Seconds 2
  Write-Host "[start] launching ngrok tunnel -> http://localhost:$port" -ForegroundColor Cyan
  $ngrok = Start-Process -FilePath "ngrok" -ArgumentList @("http", $port, "--log=stdout") -PassThru -NoNewWindow
  Start-Sleep -Seconds 3
  try {
    $tun = (Invoke-RestMethod "http://127.0.0.1:4040/api/tunnels").tunnels |
           Where-Object { $_.proto -eq "https" } | Select-Object -First 1
    if ($tun) {
      Write-Host ""
      Write-Host "  Public URL : $($tun.public_url)" -ForegroundColor Green
      Write-Host "  Admin      : $($tun.public_url)/admin" -ForegroundColor Green
      Write-Host "  Docs       : $($tun.public_url)/docs" -ForegroundColor Green
      Write-Host "  ngrok web  : http://127.0.0.1:4040" -ForegroundColor Green
      Write-Host ""
    }
  } catch {
    Write-Host "[start] ngrok started; inspect tunnels at http://127.0.0.1:4040" -ForegroundColor Yellow
  }
}

Write-Host "[start] local admin: http://localhost:$port/admin (user: $AdminUser)" -ForegroundColor Cyan
Write-Host "[start] Ctrl+C to stop." -ForegroundColor DarkGray

try {
  Wait-Process -Id $server.Id
} finally {
  if ($Ngrok -and $ngrok -and -not $ngrok.HasExited) { Stop-Process -Id $ngrok.Id -Force -ErrorAction SilentlyContinue }
  if ($server -and -not $server.HasExited) { Stop-Process -Id $server.Id -Force -ErrorAction SilentlyContinue }
}
