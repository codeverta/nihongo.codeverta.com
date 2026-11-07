package api_test

import (
	"os"
	"testing"

	"nihongo-master-backend/internal/store"
)

// The API tests cover the Lifetime Access gates, so they run with open access off.
func TestMain(m *testing.M) {
	store.SetOpenLearningAccess(false)
	os.Exit(m.Run())
}
