package unlock

// SetMaxTxAttempts changes how often a write transaction is retried and
// returns a function that restores the previous value. Tests use 1 so that a
// deadlock fails the test instead of being retried away.
func SetMaxTxAttempts(n int) (restore func()) {
	previous := maxTxAttempts
	maxTxAttempts = n
	return func() { maxTxAttempts = previous }
}
