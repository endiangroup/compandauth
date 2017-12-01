package compandauth

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsLocked_ReturnsTrueWhenCAAIsNegative(t *testing.T) {
	assert.True(t, CounterCAA(-1).IsLocked(), "-1 should be locked")
	assert.True(t, CounterCAA(-math.MaxInt64).IsLocked(), "-MaxInt64 should be locked")

	assert.True(t, CounterCAA(1).Lock().IsLocked(), "Locking 1 should be considered locked")
	assert.True(t, CounterCAA(math.MaxInt64).Lock().IsLocked(), "Locking MaxInt64 should be considered locked")
}

func Test_IsLocked_ReturnsFalseWhenCAAIsPostive(t *testing.T) {
	assert.False(t, CounterCAA(1).IsLocked(), "1 should be unlocked")
	assert.False(t, CounterCAA(5).IsLocked(), "5 should be unlocked")
	assert.False(t, CounterCAA(math.MaxInt64).IsLocked(), "MaxInt64 should be considered unlocked")

	assert.False(t, CounterCAA(-1).Unlock().IsLocked(), "Unlocking -1 should be considered unlocked")
	assert.False(t, CounterCAA(-math.MaxInt64).Unlock().IsLocked(), "Unlocking -MaxInt64 should be considered unlocked")
}

func Test_Lock_IsIdempotent(t *testing.T) {
	assert.True(t, CounterCAA(-4).Lock().IsLocked(), "Locking -4 should be considered locked")
	assert.True(t, CounterCAA(-math.MaxInt64).Lock().IsLocked(), "Locking -MaxInt64 should be considered locked")
}

func Test_Lock_ReturnsNegativeCurrentCAAValue(t *testing.T) {
	tests := []struct {
		CAA         CounterCAA
		ExpectedCAA CounterCAA
	}{
		{
			CAA:         CounterCAA(0),
			ExpectedCAA: CounterCAA(-0),
		},
		{
			CAA:         CounterCAA(1),
			ExpectedCAA: CounterCAA(-1),
		},
		{
			CAA:         CounterCAA(math.MaxInt64),
			ExpectedCAA: CounterCAA(-math.MaxInt64),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.ExpectedCAA, test.CAA.Lock())
	}
}

func Test_Unlock_IsIdempotent(t *testing.T) {
	assert.False(t, CounterCAA(0).Unlock().IsLocked(), "Unlocking 0 should be considered unlocked")
	assert.False(t, CounterCAA(1).Unlock().IsLocked(), "Unlocking 1 should be considered unlocked")
	assert.False(t, CounterCAA(5).Unlock().IsLocked(), "Unlocking 5 should be considered unlocked")
	assert.False(t, CounterCAA(math.MaxInt64).Unlock().IsLocked(), "Unlocking MaxInt64 should be considered unlocked")
}

func Test_IsValid_ReturnsFalseIfCAAIsLocked(t *testing.T) {
	assert.False(t, CounterCAA(-1).IsValid(0, 1))
	assert.False(t, CounterCAA(1).Lock().IsValid(0, 1))
	assert.False(t, CounterCAA(50).Lock().IsValid(45, 10))
	assert.False(t, CounterCAA(-math.MaxInt64).IsValid(1, 2))
}

func Test_IsValid_ReturnsFalseIfCAAHasNotIssued(t *testing.T) {
	assert.False(t, CounterCAA(0).IsValid(0, 0))
}

func Test_IsValid_ReturnsTrueIfIncomingCAAPlusDeltaIsGreaterThanOrEqualToCurrentCAA(t *testing.T) {
	assert.True(t, CounterCAA(1).IsValid(0, 1))
	assert.True(t, CounterCAA(50).IsValid(45, 10))
}

func Test_Issue_ReturnsSessionCAAValueAndIncrementedCAAValue(t *testing.T) {
	tests := []struct {
		CAA                CounterCAA
		ExpectedCAA        CounterCAA
		ExpectedSessionCAA int64
	}{
		{
			CAA:                CounterCAA(0),
			ExpectedCAA:        CounterCAA(1),
			ExpectedSessionCAA: 0,
		},
		{
			CAA:                CounterCAA(1),
			ExpectedCAA:        CounterCAA(2),
			ExpectedSessionCAA: 1,
		},
		{
			CAA:                CounterCAA(math.MaxInt64 - 1),
			ExpectedCAA:        CounterCAA(math.MaxInt64),
			ExpectedSessionCAA: math.MaxInt64 - 1,
		},
	}

	for _, test := range tests {
		sessionCAA, caa := test.CAA.Issue()

		assert.Equal(t, test.ExpectedSessionCAA, sessionCAA)
		assert.Equal(t, test.ExpectedCAA, caa)
	}
}

func Test_Issue_ReturnsValidSessionCAAAndIncrementedCAAWhenIsLocked(t *testing.T) {
	tests := []struct {
		CAA                CounterCAA
		ExpectedCAA        CounterCAA
		ExpectedSessionCAA int64
	}{
		{
			CAA:                CounterCAA(-1),
			ExpectedCAA:        CounterCAA(-2),
			ExpectedSessionCAA: 1,
		},
		{
			CAA:                CounterCAA(-2),
			ExpectedCAA:        CounterCAA(-3),
			ExpectedSessionCAA: 2,
		},
		{
			CAA:                CounterCAA(-math.MaxInt64 + 1),
			ExpectedCAA:        CounterCAA(-math.MaxInt64),
			ExpectedSessionCAA: math.MaxInt64 - 1,
		},
	}
	for _, test := range tests {
		sessionCAA, caa := test.CAA.Issue()

		assert.Equal(t, test.ExpectedSessionCAA, sessionCAA)
		assert.Equal(t, test.ExpectedCAA, caa)
	}
}

func Test_Revoke_ReturnsUnmodifiedCAAWhenItHasNeverIssued(t *testing.T) {
	assert.Equal(t, CounterCAA(0), CounterCAA(0).Revoke(10))
}

func Test_Revoke_ReturnsCAAWithRevocationsWhenLocked(t *testing.T) {
	tests := []struct {
		CAA         CounterCAA
		RevokeN     int64
		ExpectedCAA CounterCAA
	}{
		{
			CAA:         CounterCAA(-1),
			RevokeN:     1,
			ExpectedCAA: CounterCAA(-2),
		},
		{
			CAA:         CounterCAA(-4),
			RevokeN:     10,
			ExpectedCAA: CounterCAA(-14),
		},
		{
			CAA:         CounterCAA(-math.MaxInt64 + 1),
			RevokeN:     1,
			ExpectedCAA: CounterCAA(-math.MaxInt64),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.ExpectedCAA, test.CAA.Revoke(test.RevokeN))
	}
}

func Test_Revoke_ReturnsCAAWithRevocations(t *testing.T) {
	tests := []struct {
		CAA         CounterCAA
		RevokeN     int64
		ExpectedCAA CounterCAA
	}{
		{
			CAA:         CounterCAA(1),
			RevokeN:     1,
			ExpectedCAA: CounterCAA(2),
		},
		{
			CAA:         CounterCAA(4),
			RevokeN:     10,
			ExpectedCAA: CounterCAA(14),
		},
		{
			CAA:         CounterCAA(math.MaxInt64 - 1),
			RevokeN:     1,
			ExpectedCAA: CounterCAA(math.MaxInt64),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.ExpectedCAA, test.CAA.Revoke(test.RevokeN))
	}
}
