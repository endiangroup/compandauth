package compandauth

import "sync"

func NewThreadSafe(caa CAA) ThreadSafe {
	return ThreadSafe{
		CAA: caa,
	}
}

// Only to be used if the goroutine which fetches the CAA
// starts new co-routines sharing the same CAA. E.g.
// The routine which handles an incoming request fetches
// an entity with a CAA attached, it then proceeds to
// spin off go routines with that entity which might affect
// the CAA
type ThreadSafe struct {
	CAA
	mu sync.RWMutex
}

func (t *ThreadSafe) Lock() {
	t.mu.Lock()
	t.Lock()
	t.mu.Unlock()
}

func (t *ThreadSafe) Unlock() {
	t.mu.Lock()
	t.Unlock()
	t.mu.Unlock()
}

func (t *ThreadSafe) IsLocked() bool {
	t.mu.RLock()
	isLocked := t.IsLocked()
	t.mu.RUnlock()

	return isLocked
}

func (t *ThreadSafe) IsValid(s SessionCAA, n int64) bool {
	t.mu.RLock()
	isValid := t.IsValid(s, n)
	t.mu.RUnlock()

	return isValid
}

func (t *ThreadSafe) Revoke(n int64) {
	t.mu.Lock()
	t.Revoke(n)
	t.mu.Unlock()
}

func (t *ThreadSafe) Issue() SessionCAA {
	t.mu.Lock()
	sessionCaa := t.Issue()
	t.mu.Unlock()

	return sessionCaa
}

func (t *ThreadSafe) HasIssued() bool {
	t.mu.RLock()
	hasIssued := t.HasIssued()
	t.mu.RUnlock()

	return hasIssued
}
