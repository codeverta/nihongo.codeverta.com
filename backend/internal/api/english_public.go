package api

import (
	"os"
	"strings"
)

const (
	englishPublicContentFlag = "PUBLIC_ENGLISH_CONTENT_ENABLED"
	englishQAPreviewFlag     = "ENGLISH_QA_PREVIEW_ENABLED"
)

func publicEnglishContentEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv(englishPublicContentFlag)), "true")
}

func englishLearningContentEnabled() bool {
	return publicEnglishContentEnabled() || strings.EqualFold(strings.TrimSpace(os.Getenv(englishQAPreviewFlag)), "true")
}
