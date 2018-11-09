package compandauth

// Compare-and-authenticate
//
// A single counter used to maintain a set number of distributed sessions. Each
// time a session is issued the Counter is incremnented and a value is returned to
// be stored in the distributed session. When the distributed session is
// received to be validated the value stored in the session is compared with
// the Counter and if it is within the delta then the session can be considered
// valid. The Counter can be locked and unlocked at will with out a need to update
// the distributed sessions. Finally distributed sessions can be revoked in
// chronological order (although typically you would revoke all existing
// sessions) with no need to have access to them.
type Counter int64

func NewCounter() *Counter {
	return new(Counter)
}

// Locks CAA to prevent validation of session CAA's.
func (caa *Counter) Lock() {
	*caa = -caa.abs()
}

// Unlocks CAA to allow validation of session CAA's.
func (caa *Counter) Unlock() {
	*caa = caa.abs()
}

func (caa Counter) IsLocked() bool {
	return caa < 0
}

// Indicates if an incoming session CAA is considered valid. s should
// be the CAA value retrieved from a distributed session. delta represents
// number of active distributed sessions you would like to maintain per CAA.
func (caa Counter) IsValid(s SessionCAA, delta int64) bool {
	sessionCAA := abs(int64(s))
	delta = abs(delta)

	return !caa.IsLocked() &&
		caa.HasIssued() &&
		(sessionCAA+delta) >= int64(caa.abs())
}

// Invalidates the oldest n sessions. Set n to delta to invalidate all active
// sessions. If the CAA has never issued it has no effect. If the CAA has been
// locked it will still perform the revocations which will come into effect
// when the CAA is unlocked.
func (caa *Counter) Revoke(n int64) {
	if !caa.HasIssued() {
		return
	}

	caa.step(n)
}

// Issues the next CAA value to use in a distributed session and the
// incremented CAA. If locked it will return the next valid session CAA value
// and progress the CAA with out unlocking it (the session will be considered
// invalid whilst the CAA remains locked).
func (caa *Counter) Issue() SessionCAA {
	sessionCAA := SessionCAA(caa.abs())
	caa.step(1)

	return sessionCAA
}

// Indicates if the CAA has issued at least once, regardless if it has been
// locked.
func (caa Counter) HasIssued() bool {
	return caa != 0
}

func (caa Counter) abs() Counter {
	return Counter(abs(int64(caa)))
}

func (caa *Counter) step(n int64) {
	if caa.IsLocked() {
		caa.decrement(n)
	} else {
		caa.increment(n)
	}
}

func (caa *Counter) increment(n int64) {
	*caa += Counter(n)
}
func (caa *Counter) decrement(n int64) {
	*caa -= Counter(n)
}

var _ = CAA(NewCounter())
