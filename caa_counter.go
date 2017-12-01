package compandauth

// Compare-and-authenticate
//
// A single counter used to maintain a set number of distributed sessions. Each
// time a session is issued the CounterCAA is incremnented and a value is returned to
// be stored in the distributed session. When the distributed session is
// received to be validated the value stored in the session is compared with
// the CounterCAA and if it is within the delta then the session can be considered
// valid. The CounterCAA can be locked and unlocked at will with out a need to update
// the distributed sessions. Finally distributed sessions can be revoked in
// chronological order (although typically you would revoke all existing
// sessions) with no need to have access to them.
type CounterCAA int64

func NewCounter() *CounterCAA {
	return new(CounterCAA)
}

// Locks CAA to prevent validation of session CAA's.
func (caa *CounterCAA) Lock() {
	*caa = -caa.abs()
}

// Unlocks CAA to allow validation of session CAA's.
func (caa *CounterCAA) Unlock() {
	*caa = caa.abs()
}

func (caa CounterCAA) IsLocked() bool {
	return caa < 0
}

// Indicates if an incoming session CAA is considered valid. s should
// be the CAA value retrieved from a distributed session. delta represents
// number of active distributed sessions you would like to maintain per CAA.
func (caa CounterCAA) IsValid(s SessionCAA, delta int64) bool {
	sessionCAA := abs(int64(s))
	delta = abs(delta)

	return !caa.IsLocked() &&
		caa.HasIssued() &&
		(sessionCAA+delta) >= int64(caa)
}

// Invalidates the oldest n sessions. Set n to delta to invalidate all active
// sessions. If the CAA has never issued it has no effect. If the CAA has been
// locked it will still perform the revocations which will come into effect
// when the CAA is unlocked.
func (caa *CounterCAA) Revoke(n int64) {
	if !caa.HasIssued() {
		return
	}

	caa.step(n)
}

// Issues the next CAA value to use in a distributed session and the
// incremented CAA. If locked it will return the next valid session CAA value
// and progress the CAA with out unlocking it (the session will be considered
// invalid whilst the CAA remains locked).
func (caa *CounterCAA) Issue() SessionCAA {
	defer caa.step(1)

	return SessionCAA(caa.abs())
}

// Indicates if the CAA has issued at least once, regardless if it has been
// locked.
func (caa CounterCAA) HasIssued() bool {
	return caa != 0
}

func (caa CounterCAA) abs() CounterCAA {
	return CounterCAA(abs(int64(caa)))
}

func (caa *CounterCAA) step(n int64) {
	if caa.IsLocked() {
		caa.decrement(n)
	} else {
		caa.increment(n)
	}
}

func (caa *CounterCAA) increment(n int64) {
	*caa += CounterCAA(n)
}
func (caa *CounterCAA) decrement(n int64) {
	*caa -= CounterCAA(n)
}

var _ = CAA(NewCounter())
