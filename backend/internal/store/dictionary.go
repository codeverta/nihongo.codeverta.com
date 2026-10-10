package store

import "nihongo-master-backend/internal/dictionary"

// DictionaryCache reuses the existing connection, including its exclusive
// SQLite lock. Call during startup, before HTTP traffic is accepted.
func (s *MemoryStore) DictionaryCache() (*dictionary.SQLCache, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return dictionary.NewSQLCache(s.gamificationDB)
}
