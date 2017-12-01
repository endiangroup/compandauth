package compandauth

type CAA interface {
	Lock() CAA
	Unlock() CAA
	IsLocked() bool

	IsValid(int64, int64) bool

	Revoke(int64) CAA
	Issue() (int64, CAA)
	HasIssued() bool
}
