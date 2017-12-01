package compandauth

import (
	"time"

	"github.com/adrianduke/compandauth/clock"
)

type TimeoutCAA int64

// Locks CAA to prevent validation of session CAA's.
func (caa TimeoutCAA) Lock() CAA {
	return -caa.abs()
}

// Unlocks CAA to allow validation of session CAA's.
func (caa TimeoutCAA) Unlock() CAA {
	return caa.abs()
}

func (caa TimeoutCAA) IsLocked() bool {
	return caa < 0
}

// Indicates if an incoming session CAA is considered valid. incomingCAA should
// be the CAA value retrieved from a session token (e.g. JWT). durationSecs represents
// number of seconds you would like to consider a session valid for.
func (caa TimeoutCAA) IsValid(incomingTimestamp int64, durationSecs int64) bool {
	incomingTimestamp = abs(incomingTimestamp)
	durationSecs = abs(durationSecs)
	expiryTimestamp := int64(caa)

	return !caa.IsLocked() &&
		caa.HasIssued() &&
		incomingTimestamp >= expiryTimestamp &&
		(incomingTimestamp+durationSecs) >= clock.Now().Unix()
}

// Utility function to convert time.Duration into int64 seconds
func ToSeconds(d time.Duration) int64 {
	return int64(d.Seconds())
}

// Invalidates all sessions issued before expiryTimstamp (which should be a unix
// timestamp in seconds). If CAA hasn't ever issued expiryTimstamp is ignored and the
// CAA is returned as is. If CAA is locked it will perform necessary
// conversions on expiryTimstamp. Set to now to invalid all previously issued sessions.
func (caa TimeoutCAA) Revoke(expiryTimestamp int64) CAA {
	if !caa.HasIssued() {
		return caa
	}

	if caa.IsLocked() {
		return TimeoutCAA(-abs(expiryTimestamp))
	}

	return TimeoutCAA(abs(expiryTimestamp))
}

// Issues the next CAA value to use in a distributed session and the CAA. If
// locked it will still return the next valid session CAA value. CAA is only
// set on first issue.
func (caa TimeoutCAA) Issue() (int64, CAA) {
	now := clock.Now().Unix()
	if !caa.HasIssued() {
		return now, TimeoutCAA(now)
	}

	return now, caa
}

// Indicates if the CAA has issued at least once, regardless if it has been
// locked.
func (caa TimeoutCAA) HasIssued() bool {
	return caa != 0
}

func (caa TimeoutCAA) abs() TimeoutCAA {
	return TimeoutCAA(abs(int64(caa)))
}

var _ = CAA(TimeoutCAA(0))
