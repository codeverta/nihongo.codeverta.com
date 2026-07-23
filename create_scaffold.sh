#!/bin/bash
# Bash Script - Nihongo Master Scaffold Generator
# Menghasilkan struktur Feature-First Clean Architecture di folder lib/

LIB_PATH="./lib"

folders=(
  # CORE
  "$LIB_PATH/core/constants"
  "$LIB_PATH/core/network"
  "$LIB_PATH/core/routing"
  "$LIB_PATH/core/theme"
  "$LIB_PATH/core/utils"

  # SHARED
  "$LIB_PATH/shared/widgets"

  # FEATURES
  "$LIB_PATH/features/auth/data/datasources"
  "$LIB_PATH/features/auth/data/models"
  "$LIB_PATH/features/auth/data/repositories"
  "$LIB_PATH/features/auth/domain/entities"
  "$LIB_PATH/features/auth/domain/repositories"
  "$LIB_PATH/features/auth/domain/usecases"
  "$LIB_PATH/features/auth/presentation/controllers"
  "$LIB_PATH/features/auth/presentation/screens"
  "$LIB_PATH/features/auth/presentation/widgets"

  "$LIB_PATH/features/onboarding/data/datasources"
  "$LIB_PATH/features/onboarding/data/models"
  "$LIB_PATH/features/onboarding/data/repositories"
  "$LIB_PATH/features/onboarding/domain/entities"
  "$LIB_PATH/features/onboarding/domain/repositories"
  "$LIB_PATH/features/onboarding/presentation/controllers"
  "$LIB_PATH/features/onboarding/presentation/screens"
  "$LIB_PATH/features/onboarding/presentation/widgets"

  "$LIB_PATH/features/home/data/datasources"
  "$LIB_PATH/features/home/data/models"
  "$LIB_PATH/features/home/data/repositories"
  "$LIB_PATH/features/home/domain/entities"
  "$LIB_PATH/features/home/domain/repositories"
  "$LIB_PATH/features/home/presentation/controllers"
  "$LIB_PATH/features/home/presentation/screens"
  "$LIB_PATH/features/home/presentation/widgets"

  "$LIB_PATH/features/learn/data/datasources"
  "$LIB_PATH/features/learn/data/models"
  "$LIB_PATH/features/learn/data/repositories"
  "$LIB_PATH/features/learn/domain/entities"
  "$LIB_PATH/features/learn/domain/repositories"
  "$LIB_PATH/features/learn/presentation/controllers"
  "$LIB_PATH/features/learn/presentation/screens"
  "$LIB_PATH/features/learn/presentation/widgets"

  "$LIB_PATH/features/exam/data/datasources"
  "$LIB_PATH/features/exam/data/models"
  "$LIB_PATH/features/exam/data/repositories"
  "$LIB_PATH/features/exam/domain/entities"
  "$LIB_PATH/features/exam/domain/repositories"
  "$LIB_PATH/features/exam/presentation/controllers"
  "$LIB_PATH/features/exam/presentation/screens"
  "$LIB_PATH/features/exam/presentation/widgets"

  "$LIB_PATH/features/profile/data/datasources"
  "$LIB_PATH/features/profile/data/models"
  "$LIB_PATH/features/profile/data/repositories"
  "$LIB_PATH/features/profile/domain/entities"
  "$LIB_PATH/features/profile/domain/repositories"
  "$LIB_PATH/features/profile/presentation/controllers"
  "$LIB_PATH/features/profile/presentation/screens"
  "$LIB_PATH/features/profile/presentation/widgets"
)

for folder in "${folders[@]}"; do
  mkdir -p "$folder"
  echo "[CREATED] Folder: $folder"
done

echo -e "\nStruktur direktori Nihongo Master berhasil dibuat!"
