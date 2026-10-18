package store

import (
	"errors"
	"gorm.io/gorm"
)

// Jobs uses independent tables and durable per-domain leases, not the cached
// account-state writer. No account or purchase state is exposed.
func (s *MemoryStore) JobsDatabase() (*gorm.DB, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ormDB == nil || s.persistenceClosed {
		return nil, errors.New("MariaDB persistence is unavailable")
	}
	return s.ormDB, nil
}
