param(
    [string]$ApiBaseUrl,
    [string]$GoogleServerClientId,
    [switch]$Universal,
    [switch]$AppBundle
)

$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
Push-Location $projectRoot
try {
    $buildArgs = @('build', $(if ($AppBundle) { 'appbundle' } else { 'apk' }), '--release', '--split-debug-info=build/symbols')
    if (-not $AppBundle -and -not $Universal) { $buildArgs += '--split-per-abi' }
    if ($ApiBaseUrl) { $buildArgs += "--dart-define=API_BASE_URL=$ApiBaseUrl" }
    if ($GoogleServerClientId) { $buildArgs += "--dart-define=GOOGLE_SERVER_CLIENT_ID=$GoogleServerClientId" }
    & flutter @buildArgs
    if ($LASTEXITCODE -ne 0) { throw "Flutter build failed with exit code $LASTEXITCODE" }
} finally {
    Pop-Location
}
