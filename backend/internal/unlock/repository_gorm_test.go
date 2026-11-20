package unlock_test

import (
	"testing"

	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/testdb"
	"nihongo-master-backend/internal/unlock"
)

func newGormRepo(t *testing.T) *unlock.GormRepository {
	t.Helper()
	db, err := database.Open(testdb.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	repo, err := unlock.NewGormRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	return repo
}

// The MariaDB repository must behave exactly like the in-memory one, and its
// conditional updates and unique index are what make the limits atomic.
func TestGormRepository(t *testing.T) {
	// Deadlocks must not happen at all, so the last-resort retry is off here:
	// a single deadlock surfaces as an error and fails the scenario.
	defer unlock.SetMaxTxAttempts(1)()
	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) { tc.run(t, newHarness(t, newGormRepo(t))) })
	}
}
