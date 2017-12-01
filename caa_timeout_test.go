package compandauth

import (
	"math"
	"testing"
	"time"

	"github.com/adrianduke/compandauth/clock"
	"github.com/stretchr/testify/assert"
)

func Test_IsLocked_ReturnsTrueWhenTimeCAAIsNegative(t *testing.T) {
	assert.True(t, TimeoutCAA(-1).IsLocked(), "-1 should be locked")
	assert.True(t, TimeoutCAA(-math.MaxInt64).IsLocked(), "-MaxInt64 should be locked")
}

func Test_IsLocked_ReturnsFalseWhenTimeCAAIsPostive(t *testing.T) {
	assert.False(t, TimeoutCAA(1).IsLocked(), "1 should be unlocked")
	assert.False(t, TimeoutCAA(5).IsLocked(), "5 should be unlocked")
	assert.False(t, TimeoutCAA(math.MaxInt64).IsLocked(), "MaxInt64 should be considered unlocked")
}

func Test_Lock_IsIdempotentTimeCAA(t *testing.T) {
	assert.True(t, TimeoutCAA(-4).Lock().IsLocked(), "Locking -4 should be considered locked")
	assert.True(t, TimeoutCAA(-math.MaxInt64).Lock().IsLocked(), "Locking -MaxInt64 should be considered locked")
}

func Test_Lock_ReturnsNegativeCurrentTimeCAAValue(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		CAA         TimeoutCAA
		ExpectedCAA TimeoutCAA
	}{
		{
			CAA:         TimeoutCAA(0),
			ExpectedCAA: TimeoutCAA(-0),
		},
		{
			CAA:         TimeoutCAA(now),
			ExpectedCAA: TimeoutCAA(-now),
		},
		{
			CAA:         TimeoutCAA(math.MaxInt64),
			ExpectedCAA: TimeoutCAA(-math.MaxInt64),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.ExpectedCAA, test.CAA.Lock())
	}
}

func Test_Unlock_IsIdempotentTimeCAA(t *testing.T) {
	assert.False(t, TimeoutCAA(0).Unlock().IsLocked(), "Unlocking 0 should be considered unlocked")
	assert.False(t, TimeoutCAA(1).Unlock().IsLocked(), "Unlocking 1 should be considered unlocked")
	assert.False(t, TimeoutCAA(time.Now().Unix()).Unlock().IsLocked(), "Unlocking 5 should be considered unlocked")
	assert.False(t, TimeoutCAA(math.MaxInt64).Unlock().IsLocked(), "Unlocking MaxInt64 should be considered unlocked")
}

func Test_IsValid_ReturnsFalseIfTimeCAAIsLocked(t *testing.T) {
	now := time.Now().Unix()

	assert.False(t, TimeoutCAA(-1).IsValid(0, 1))
	assert.False(t, TimeoutCAA(1).Lock().IsValid(0, 1))
	assert.False(t, TimeoutCAA(now).Lock().IsValid(now-5, 10))
	assert.False(t, TimeoutCAA(-math.MaxInt64).IsValid(1, 2))
}

func Test_IsValid_ReturnsFalseIfTimeCAAHasNotIssued(t *testing.T) {
	assert.False(t, TimeoutCAA(0).IsValid(0, 0))
}

func Test_IsValid_ReturnsFalseIfIncomingCAAWasIssuedBeforeTimeCAA(t *testing.T) {
	now := time.Now().Unix()

	assert.False(t, TimeoutCAA(1).IsValid(0, 1))
	assert.False(t, TimeoutCAA(now).IsValid(now-1, 10))
}

func Test_IsValid_ReturnsTrueIfIncomingCAAPlusDeltaIsAfterOrEqualToNow(t *testing.T) {
	now := time.Now().Unix()

	assert.True(t, TimeoutCAA(1).IsValid(now, 0))
	assert.True(t, TimeoutCAA(1).IsValid(now, 1))
	assert.True(t, TimeoutCAA(1).IsValid(now-5, 10))
	assert.True(t, TimeoutCAA(1).IsValid(now-10, 10))
}

func Test_Issue_SetsTimeCAAToNowOnFirstIssue(t *testing.T) {
	now := time.Now()
	clock.NowForce(now)
	defer clock.NowReset()

	_, timeCAA := TimeoutCAA(0).Issue()

	assert.Equal(t, now.Unix(), int64(timeCAA.(TimeoutCAA)))
}

func Test_Issue_ReturnsNowAndCurrentTimeCAAAfterFirstIssue(t *testing.T) {
	now := time.Now()
	clock.NowForce(now)
	defer clock.NowReset()

	tests := []struct {
		CAA                TimeoutCAA
		ExpectedCAA        TimeoutCAA
		ExpectedSessionCAA int64
	}{
		{
			CAA:                TimeoutCAA(1),
			ExpectedCAA:        TimeoutCAA(1),
			ExpectedSessionCAA: now.Unix(),
		},
		{
			CAA:                TimeoutCAA(now.Unix() - 500),
			ExpectedCAA:        TimeoutCAA(now.Unix() - 500),
			ExpectedSessionCAA: now.Unix(),
		},
		{
			CAA:                TimeoutCAA(math.MaxInt64),
			ExpectedCAA:        TimeoutCAA(math.MaxInt64),
			ExpectedSessionCAA: now.Unix(),
		},
		{
			CAA:                TimeoutCAA(-math.MaxInt64),
			ExpectedCAA:        TimeoutCAA(-math.MaxInt64),
			ExpectedSessionCAA: now.Unix(),
		},
	}

	for _, test := range tests {
		sessionCAA, caa := test.CAA.Issue()

		assert.Equal(t, test.ExpectedSessionCAA, sessionCAA)
		assert.Equal(t, test.ExpectedCAA, caa)
	}
}

func Test_Revoke_ReturnsUnmodifiedTimeCAAWhenItHasNeverIssued(t *testing.T) {
	assert.Equal(t, TimeoutCAA(0), TimeoutCAA(0).Revoke(10))
}

func Test_Revoke_ReturnsNAsNegativeTimeCAAWhenLocked(t *testing.T) {
	tests := []struct {
		CAA         TimeoutCAA
		RevokeN     int64
		ExpectedCAA TimeoutCAA
	}{
		{
			CAA:         TimeoutCAA(-1),
			RevokeN:     2,
			ExpectedCAA: TimeoutCAA(-2),
		},
		{
			CAA:         TimeoutCAA(-4),
			RevokeN:     -10,
			ExpectedCAA: TimeoutCAA(-10),
		},
		{
			CAA:         TimeoutCAA(-1),
			RevokeN:     math.MaxInt64,
			ExpectedCAA: TimeoutCAA(-math.MaxInt64),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.ExpectedCAA, test.CAA.Revoke(test.RevokeN))
	}
}

func Test_Revoke_ReturnsNAsTimeCAA(t *testing.T) {
	tests := []struct {
		CAA         TimeoutCAA
		RevokeN     int64
		ExpectedCAA TimeoutCAA
	}{
		{
			CAA:         TimeoutCAA(1),
			RevokeN:     1,
			ExpectedCAA: TimeoutCAA(1),
		},
		{
			CAA:         TimeoutCAA(4),
			RevokeN:     10,
			ExpectedCAA: TimeoutCAA(10),
		},
		{
			CAA:         TimeoutCAA(clock.Now().Unix()),
			RevokeN:     math.MaxInt64,
			ExpectedCAA: TimeoutCAA(math.MaxInt64),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.ExpectedCAA, test.CAA.Revoke(test.RevokeN))
	}
}
