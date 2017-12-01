package compandauth

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setCounterCAA(i int64) *CounterCAA {
	caa := NewCounter()
	*caa = CounterCAA(i)

	return caa
}

func Test_IsLocked_ReturnsTrueWhenCAAIsNegative(t *testing.T) {
	tests := []struct {
		CAA *CounterCAA
	}{
		{CAA: setCounterCAA(-1)},
		{CAA: setCounterCAA(-5)},
		{CAA: setCounterCAA(-math.MaxInt64)},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			assert.True(t, test.CAA.IsLocked())
		})
	}
}

func Test_IsLocked_ReturnsFalseWhenCAAIsPostive(t *testing.T) {
	tests := []struct {
		CAA *CounterCAA
	}{
		{CAA: setCounterCAA(1)},
		{CAA: setCounterCAA(5)},
		{CAA: setCounterCAA(math.MaxInt64)},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			assert.False(t, test.CAA.IsLocked())
		})
	}
}

func Test_Lock_IsIdempotent(t *testing.T) {
	tests := []struct {
		CAA *CounterCAA
	}{
		{CAA: setCounterCAA(-1)},
		{CAA: setCounterCAA(-5)},
		{CAA: setCounterCAA(-math.MaxInt64)},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			test.CAA.Lock()

			assert.True(t, test.CAA.IsLocked())
		})
	}
}

func Test_Lock_SetsNegativeCurrentCAAValue(t *testing.T) {
	tests := []struct {
		CAA         *CounterCAA
		ExpectedCAA *CounterCAA
	}{
		{
			CAA:         setCounterCAA(0),
			ExpectedCAA: setCounterCAA(-0),
		},
		{
			CAA:         setCounterCAA(1),
			ExpectedCAA: setCounterCAA(-1),
		},
		{
			CAA:         setCounterCAA(math.MaxInt64),
			ExpectedCAA: setCounterCAA(-math.MaxInt64),
		},
	}

	for _, test := range tests {
		test.CAA.Lock()

		assert.Equal(t, test.ExpectedCAA, test.CAA)
	}
}

func Test_Unlock_IsIdempotent(t *testing.T) {
	tests := []struct {
		CAA *CounterCAA
	}{
		{CAA: setCounterCAA(1)},
		{CAA: setCounterCAA(5)},
		{CAA: setCounterCAA(math.MaxInt64)},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			test.CAA.Unlock()

			assert.False(t, test.CAA.IsLocked())
		})
	}
}

func Test_IsValid_ReturnsFalseIfCAAIsLocked(t *testing.T) {
	tests := []struct {
		CAA        *CounterCAA
		SessionCAA SessionCAA
		Delta      int64
	}{
		{CAA: setCounterCAA(-1), Delta: -1, SessionCAA: -1},
		{CAA: setCounterCAA(-1), Delta: -1, SessionCAA: 0},
		{CAA: setCounterCAA(-1), Delta: -1, SessionCAA: 1},
		{CAA: setCounterCAA(-1), Delta: 0, SessionCAA: -1},
		{CAA: setCounterCAA(-1), Delta: 0, SessionCAA: 0},
		{CAA: setCounterCAA(-1), Delta: 0, SessionCAA: 1},
		{CAA: setCounterCAA(-1), Delta: 1, SessionCAA: -1},
		{CAA: setCounterCAA(-1), Delta: 1, SessionCAA: 0},
		{CAA: setCounterCAA(-1), Delta: 1, SessionCAA: 1},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			assert.False(t, test.CAA.IsValid(test.SessionCAA, test.Delta))
		})
	}
}

func Test_IsValid_ReturnsFalseIfCAAHasNotIssued(t *testing.T) {
	assert.False(t, NewCounter().IsValid(0, 0))
}

func Test_IsValid_ReturnsTrueIfSessionCAAPlusDeltaIsGreaterThanOrEqualToCurrentCAA(t *testing.T) {
	assert.True(t, setCounterCAA(1).IsValid(0, 1))
	assert.True(t, setCounterCAA(50).IsValid(45, 10))
}

func Test_Issue_ReturnsNextSessionCAAValueAndIncrementsCounterCAA(t *testing.T) {
	tests := []struct {
		CAA                *CounterCAA
		ExpectedCAA        *CounterCAA
		ExpectedSessionCAA SessionCAA
	}{
		{
			CAA:                setCounterCAA(0),
			ExpectedCAA:        setCounterCAA(1),
			ExpectedSessionCAA: 0,
		},
		{
			CAA:                setCounterCAA(1),
			ExpectedCAA:        setCounterCAA(2),
			ExpectedSessionCAA: 1,
		},
		{
			CAA:                setCounterCAA(math.MaxInt64 - 1),
			ExpectedCAA:        setCounterCAA(math.MaxInt64),
			ExpectedSessionCAA: math.MaxInt64 - 1,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			sessionCAA := test.CAA.Issue()

			assert.Equal(t, test.ExpectedSessionCAA, sessionCAA)
			assert.Equal(t, test.ExpectedCAA, test.CAA)
		})
	}
}

func Test_Issue_ReturnsNextSessionCAAValueAndIncrementedCAAWhenIsLocked(t *testing.T) {
	tests := []struct {
		CAA                *CounterCAA
		ExpectedCAA        *CounterCAA
		ExpectedSessionCAA SessionCAA
	}{
		{
			CAA:                setCounterCAA(-1),
			ExpectedCAA:        setCounterCAA(-2),
			ExpectedSessionCAA: 1,
		},
		{
			CAA:                setCounterCAA(-2),
			ExpectedCAA:        setCounterCAA(-3),
			ExpectedSessionCAA: 2,
		},
		{
			CAA:                setCounterCAA(-math.MaxInt64 + 1),
			ExpectedCAA:        setCounterCAA(-math.MaxInt64),
			ExpectedSessionCAA: math.MaxInt64 - 1,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			sessionCAA := test.CAA.Issue()

			assert.Equal(t, test.ExpectedSessionCAA, sessionCAA)
			assert.Equal(t, test.ExpectedCAA, test.CAA)
		})
	}
}

func Test_Revoke_HasNoEffectOnUnissuedCAA(t *testing.T) {
	caa := NewCounter()
	caa.Revoke(10)

	assert.Equal(t, NewCounter(), caa)
}

func Test_Revoke_IncrementsCAAWithRevocationsWhenLocked(t *testing.T) {
	tests := []struct {
		CAA         *CounterCAA
		ExpectedCAA *CounterCAA
		RevokeN     int64
	}{
		{
			CAA:         setCounterCAA(-1),
			ExpectedCAA: setCounterCAA(-2),
			RevokeN:     1,
		},
		{
			CAA:         setCounterCAA(-4),
			ExpectedCAA: setCounterCAA(-14),
			RevokeN:     10,
		},
		{
			CAA:         setCounterCAA(-math.MaxInt64 + 1),
			ExpectedCAA: setCounterCAA(-math.MaxInt64),
			RevokeN:     1,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			test.CAA.Revoke(test.RevokeN)

			assert.Equal(t, test.ExpectedCAA, test.CAA)
		})
	}
}

func Test_Revoke_ReturnsCAAWithRevocations(t *testing.T) {
	tests := []struct {
		CAA         *CounterCAA
		ExpectedCAA *CounterCAA
		RevokeN     int64
	}{
		{
			CAA:         setCounterCAA(1),
			ExpectedCAA: setCounterCAA(2),
			RevokeN:     1,
		},
		{
			CAA:         setCounterCAA(4),
			ExpectedCAA: setCounterCAA(14),
			RevokeN:     10,
		},
		{
			CAA:         setCounterCAA(math.MaxInt64 - 1),
			ExpectedCAA: setCounterCAA(math.MaxInt64),
			RevokeN:     1,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			test.CAA.Revoke(test.RevokeN)

			assert.Equal(t, test.ExpectedCAA, test.CAA)
		})
	}
}
