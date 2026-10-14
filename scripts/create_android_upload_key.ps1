$ErrorActionPreference = 'Stop'

$repositoryRoot = Split-Path -Parent $PSScriptRoot
$androidRoot = Join-Path $repositoryRoot 'android'
$secretDirectory = Join-Path $androidRoot '.signing-secrets'
$keystorePath = Join-Path $secretDirectory 'upload-key.jks'
$alias = 'nihongo-master-upload'

if (Test-Path -LiteralPath $keystorePath) {
    throw "Keystore already exists at $keystorePath. Refusing to overwrite a signing identity."
}

$keytoolCommand = Get-Command keytool -ErrorAction SilentlyContinue
if (-not $keytoolCommand) {
    throw 'keytool was not found. Install Android Studio or a Java JDK, then run this script again.'
}

New-Item -ItemType Directory -Path $secretDirectory -Force | Out-Null

$random = [System.Security.Cryptography.RandomNumberGenerator]::Create()
$storeBytes = New-Object byte[] 32
$keyBytes = New-Object byte[] 32
$random.GetBytes($storeBytes)
$random.GetBytes($keyBytes)
$storePassword = [BitConverter]::ToString($storeBytes).Replace('-', '')
$keyPassword = [BitConverter]::ToString($keyBytes).Replace('-', '')
$random.Dispose()

$encoding = New-Object System.Text.UTF8Encoding($false)
$storePasswordFile = Join-Path $secretDirectory '.storepass.tmp'
$keyPasswordFile = Join-Path $secretDirectory '.keypass.tmp'
[System.IO.File]::WriteAllText($storePasswordFile, $storePassword, $encoding)
[System.IO.File]::WriteAllText($keyPasswordFile, $keyPassword, $encoding)

try {
    & $keytoolCommand.Source -genkeypair -keystore $keystorePath -storetype JKS `
        -storepass:file $storePasswordFile -keypass:file $keyPasswordFile `
        -alias $alias -keyalg RSA -keysize 2048 -validity 10000 `
        -dname 'CN=Nihongo Master, OU=Android, O=Nihongo Master, C=ID' -noprompt *> $null
    if ($LASTEXITCODE -ne 0) {
        throw "keytool failed with exit code $LASTEXITCODE."
    }

    $keystoreBase64 = [Convert]::ToBase64String([System.IO.File]::ReadAllBytes($keystorePath))
    $secrets = [ordered]@{
        ANDROID_KEYSTORE_BASE64 = $keystoreBase64
        ANDROID_KEYSTORE_PASSWORD = $storePassword
        ANDROID_KEY_ALIAS = $alias
        ANDROID_KEY_PASSWORD = $keyPassword
    }

    foreach ($name in $secrets.Keys) {
        [System.IO.File]::WriteAllText((Join-Path $secretDirectory ($name + '.txt')), [string]$secrets[$name], $encoding)
    }

    $secretLines = foreach ($name in $secrets.Keys) { "$name`r`n$($secrets[$name])`r`n" }
    [System.IO.File]::WriteAllText((Join-Path $secretDirectory 'github-actions-secrets.txt'), ($secretLines -join "`r`n"), $encoding)

    $absoluteKeystorePath = [System.IO.Path]::GetFullPath($keystorePath).Replace('\', '/')
    $properties = "storePassword=$storePassword`nkeyPassword=$keyPassword`nkeyAlias=$alias`nstoreFile=$absoluteKeystorePath`n"
    [System.IO.File]::WriteAllText((Join-Path $androidRoot 'key.properties'), $properties, $encoding)

    $certificate = & $keytoolCommand.Source -list -v -keystore $keystorePath -storetype JKS `
        -storepass:file $storePasswordFile -alias $alias 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw 'The generated keystore could not be verified.'
    }
    $certificateText = $certificate -join "`n"
    $sha1 = [regex]::Match($certificateText, 'SHA1:\s*([^\r\n]+)').Groups[1].Value.Trim()
    $sha256 = [regex]::Match($certificateText, 'SHA256:\s*([^\r\n]+)').Groups[1].Value.Trim()
    if (-not $sha1 -or -not $sha256) {
        throw 'The generated certificate fingerprints could not be read.'
    }
    [System.IO.File]::WriteAllText((Join-Path $secretDirectory 'upload-certificate-fingerprint.txt'), "SHA-1: $sha1`nSHA-256: $sha256`n", $encoding)
} finally {
    Remove-Item -LiteralPath $storePasswordFile, $keyPasswordFile -Force -ErrorAction SilentlyContinue
}

Write-Output 'Android upload keystore created and verified.'
Write-Output "Keystore: $keystorePath"
Write-Output "Alias: $alias"
Write-Output "GitHub Secrets: $($secrets.Keys -join ', ')"
Write-Output "Secrets bundle: $(Join-Path $secretDirectory 'github-actions-secrets.txt')"
Write-Output "Local signing config: $(Join-Path $androidRoot 'key.properties')"
Write-Output "Certificate SHA-1: $sha1"
Write-Output "Certificate SHA-256: $sha256"
