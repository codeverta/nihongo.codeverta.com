package store

import (
	"os"
	"testing"
)

// Most store tests exercise the Lifetime Access gates, so they run with the
// open-access mode off. open_access_test.go turns it on explicitly.
func TestMain(m *testing.M) {
	openLearningAccess = false
	os.Exit(m.Run())
}
