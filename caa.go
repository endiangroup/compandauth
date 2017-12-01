package compandauth

type SessionCAA int64

type CAA interface {
	Lock()
	Unlock()
	IsLocked() bool

	IsValid(SessionCAA, int64) bool

	Revoke(int64)
	Issue() SessionCAA
	HasIssued() bool
}
