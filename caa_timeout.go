package compandauth

import (
	"time"

	"github.com/endiangroup/compandauth/clock"
)

type Timeout int64

func NewTimeout() *Timeout {
	return new(Timeout)
}

// Locks CAA to prevent validation of session CAA's.
func (caa *Timeout) Lock() {
	*caa = -caa.abs()
}

// Unlocks CAA to allow validation of session CAA's.
func (caa *Timeout) Unlock() {
	*caa = caa.abs()
}

func (caa Timeout) IsLocked() bool {
	return caa < 0
}

// Indicates if an session CAA is considered valid. s should be the CAA value
// retrieved from a session token (e.g. JWT). durationSecs represents
// number of seconds you would like to consider a session valid for.
func (caa Timeout) IsValid(s SessionCAA, durationSecs int64) bool {
	sessionTimestamp := abs(int64(s))
	durationSecs = abs(durationSecs)
	expiryTimestamp := int64(caa.abs())

	return !caa.IsLocked() &&
		caa.HasIssued() &&
		sessionTimestamp >= expiryTimestamp &&
		(sessionTimestamp+durationSecs) >= clock.Now().Unix()
}

// Utility function to convert time.Duration into int64 seconds
func ToSeconds(d time.Duration) int64 {
	return int64(d.Seconds())
}

// Invalidates all sessions issued before expiryTimstamp (which should be a unix
// timestamp in seconds). If CAA hasn't ever issued expiryTimstamp is ignored and the
// CAA is returned as is. If CAA is locked it will perform necessary
// conversions on expiryTimstamp. Set to now to invalid all previously issued sessions.
func (caa *Timeout) Revoke(expiryTimestamp int64) {
	if !caa.HasIssued() {
		return
	}

	caa.set(expiryTimestamp)
}

// Issues the next CAA value to use in a distributed session and the CAA. If
// locked it will still return the next valid session CAA value. CAA is only
// set on first issue.
func (caa *Timeout) Issue() SessionCAA {
	now := clock.Now().Unix()

	if !caa.HasIssued() {
		caa.set(now)
	}

	return SessionCAA(now)
}

// Indicates if the CAA has issued at least once, regardless if it has been
// locked.
func (caa Timeout) HasIssued() bool {
	return caa != 0
}

func (caa Timeout) abs() Timeout {
	return Timeout(abs(int64(caa)))
}

func (caa *Timeout) set(i int64) {
	if caa.IsLocked() {
		*caa = Timeout(-abs(i))
	} else {
		*caa = Timeout(abs(i))
	}
}

var _ = CAA(NewCounter())
