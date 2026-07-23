# PowerShell Script - Nihongo Master Scaffold Generator
# Menghasilkan struktur Feature-First Clean Architecture di folder lib/

$libPath = ".\lib"

$folders = @(
    # CORE
    "$libPath\core\constants",
    "$libPath\core\network",
    "$libPath\core\routing",
    "$libPath\core\theme",
    "$libPath\core\utils",
    
    # SHARED
    "$libPath\shared\widgets",
    
    # FEATURES - Auth
    "$libPath\features\auth\data\datasources",
    "$libPath\features\auth\data\models",
    "$libPath\features\auth\data\repositories",
    "$libPath\features\auth\domain\entities",
    "$libPath\features\auth\domain\repositories",
    "$libPath\features\auth\domain\usecases",
    "$libPath\features\auth\presentation\controllers",
    "$libPath\features\auth\presentation\screens",
    "$libPath\features\auth\presentation\widgets",
    
    # FEATURES - Onboarding
    "$libPath\features\onboarding\data\datasources",
    "$libPath\features\onboarding\data\models",
    "$libPath\features\onboarding\data\repositories",
    "$libPath\features\onboarding\domain\entities",
    "$libPath\features\onboarding\domain\repositories",
    "$libPath\features\onboarding\presentation\controllers",
    "$libPath\features\onboarding\presentation\screens",
    "$libPath\features\onboarding\presentation\widgets",

    # FEATURES - Home
    "$libPath\features\home\data\datasources",
    "$libPath\features\home\data\models",
    "$libPath\features\home\data\repositories",
    "$libPath\features\home\domain\entities",
    "$libPath\features\home\domain\repositories",
    "$libPath\features\home\presentation\controllers",
    "$libPath\features\home\presentation\screens",
    "$libPath\features\home\presentation\widgets",

    # FEATURES - Learn (Kana, Kanji, Grammar, Vocabulary)
    "$libPath\features\learn\data\datasources",
    "$libPath\features\learn\data\models",
    "$libPath\features\learn\data\repositories",
    "$libPath\features\learn\domain\entities",
    "$libPath\features\learn\domain\repositories",
    "$libPath\features\learn\presentation\controllers",
    "$libPath\features\learn\presentation\screens",
    "$libPath\features\learn\presentation\widgets",

    # FEATURES - Exam (Quiz, JLPT Mock Tests)
    "$libPath\features\exam\data\datasources",
    "$libPath\features\exam\data\models",
    "$libPath\features\exam\data\repositories",
    "$libPath\features\exam\domain\entities",
    "$libPath\features\exam\domain\repositories",
    "$libPath\features\exam\presentation\controllers",
    "$libPath\features\exam\presentation\screens",
    "$libPath\features\exam\presentation\widgets",

    # FEATURES - Profile
    "$libPath\features\profile\data\datasources",
    "$libPath\features\profile\data\models",
    "$libPath\features\profile\data\repositories",
    "$libPath\features\profile\domain\entities",
    "$libPath\features\profile\domain\repositories",
    "$libPath\features\profile\presentation\controllers",
    "$libPath\features\profile\presentation\screens",
    "$libPath\features\profile\presentation\widgets"
)

foreach ($folder in $folders) {
    if (-not (Test-Path $folder)) {
        New-Item -ItemType Directory -Path $folder -Force | Out-Null
        Write-Host "[CREATED] Folder: $folder" -ForegroundColor Green
    }
}

Write-Host "`nStruktur direktori Nihongo Master berhasil dibuat!" -ForegroundColor Cyan
