package compandauth

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/adrianduke/compandauth/clock"
	"github.com/stretchr/testify/assert"
)

func setTimeoutCAA(i int64) *TimeoutCAA {
	caa := NewTimeout()
	*caa = TimeoutCAA(i)

	return caa
}

func Test_IsLocked_ReturnsTrueWhenTimeoutCAAIsNegative(t *testing.T) {
	tests := []struct {
		CAA *TimeoutCAA
	}{
		{CAA: setTimeoutCAA(-1)},
		{CAA: setTimeoutCAA(-5)},
		{CAA: setTimeoutCAA(-math.MaxInt64)},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			assert.True(t, test.CAA.IsLocked())
		})
	}
}

func Test_IsLocked_ReturnsFalseWhenTimeoutCAAIsPostive(t *testing.T) {
	tests := []struct {
		CAA *TimeoutCAA
	}{
		{CAA: setTimeoutCAA(1)},
		{CAA: setTimeoutCAA(5)},
		{CAA: setTimeoutCAA(math.MaxInt64)},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			assert.False(t, test.CAA.IsLocked())
		})
	}
}

func Test_Lock_IsIdempotentForTimeoutCAA(t *testing.T) {
	tests := []struct {
		CAA *TimeoutCAA
	}{
		{CAA: setTimeoutCAA(-1)},
		{CAA: setTimeoutCAA(-5)},
		{CAA: setTimeoutCAA(-math.MaxInt64)},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			test.CAA.Lock()

			assert.True(t, test.CAA.IsLocked())
		})
	}
}

func Test_Lock_SetsNegativeTimeoutCAAValue(t *testing.T) {
	tests := []struct {
		CAA         *TimeoutCAA
		ExpectedCAA *TimeoutCAA
	}{
		{
			CAA:         setTimeoutCAA(0),
			ExpectedCAA: setTimeoutCAA(-0),
		},
		{
			CAA:         setTimeoutCAA(1),
			ExpectedCAA: setTimeoutCAA(-1),
		},
		{
			CAA:         setTimeoutCAA(math.MaxInt64),
			ExpectedCAA: setTimeoutCAA(-math.MaxInt64),
		},
	}

	for _, test := range tests {
		test.CAA.Lock()

		assert.Equal(t, test.ExpectedCAA, test.CAA)
	}
}

func Test_Unlock_IsIdempotentForTimeoutCAA(t *testing.T) {
	tests := []struct {
		CAA *TimeoutCAA
	}{
		{CAA: setTimeoutCAA(1)},
		{CAA: setTimeoutCAA(5)},
		{CAA: setTimeoutCAA(math.MaxInt64)},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			test.CAA.Unlock()

			assert.False(t, test.CAA.IsLocked())
		})
	}
}

func Test_IsValid_ReturnsFalseIfTimeoutCAAIsLocked(t *testing.T) {
	tests := []struct {
		CAA        *TimeoutCAA
		SessionCAA SessionCAA
		Delta      int64
	}{
		{CAA: setTimeoutCAA(-1), Delta: -1, SessionCAA: -1},
		{CAA: setTimeoutCAA(-1), Delta: -1, SessionCAA: 0},
		{CAA: setTimeoutCAA(-1), Delta: -1, SessionCAA: 1},
		{CAA: setTimeoutCAA(-1), Delta: 0, SessionCAA: -1},
		{CAA: setTimeoutCAA(-1), Delta: 0, SessionCAA: 0},
		{CAA: setTimeoutCAA(-1), Delta: 0, SessionCAA: 1},
		{CAA: setTimeoutCAA(-1), Delta: 1, SessionCAA: -1},
		{CAA: setTimeoutCAA(-1), Delta: 1, SessionCAA: 0},
		{CAA: setTimeoutCAA(-1), Delta: 1, SessionCAA: 1},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			assert.False(t, test.CAA.IsValid(test.SessionCAA, test.Delta))
		})
	}
}

func Test_IsValid_ReturnsFalseIfTimeoutCAAHasNotIssued(t *testing.T) {
	assert.False(t, NewTimeout().IsValid(0, 0))
}

func Test_IsValid_ReturnsFalseIfSessionCAAWasIssuedBeforeTimeoutCAA(t *testing.T) {
	now := SessionCAA(time.Now().Unix())

	assert.False(t, TimeoutCAA(1).IsValid(0, 1))
	assert.False(t, TimeoutCAA(now).IsValid(now-1, 10))
}

func Test_IsValid_ReturnsTrueIfSessionCAAPlusDeltaIsAfterOrEqualToNow(t *testing.T) {
	now := SessionCAA(time.Now().Unix())

	assert.True(t, TimeoutCAA(1).IsValid(now, 0))
	assert.True(t, TimeoutCAA(1).IsValid(now, 1))
	assert.True(t, TimeoutCAA(1).IsValid(now-5, 10))
	assert.True(t, TimeoutCAA(1).IsValid(now-10, 10))
}

func Test_Issue_SetsTimeCAAToNowOnFirstIssue(t *testing.T) {
	now := time.Now()
	clock.NowForce(now)
	defer clock.NowReset()
	caa := NewTimeout()

	caa.Issue()

	assert.Equal(t, now.Unix(), int64(*caa))
}

func Test_Issue_ReturnsNowAndCurrentTimeoutCAAAfterFirstIssue(t *testing.T) {
	now := time.Now()
	clock.NowForce(now)
	defer clock.NowReset()

	tests := []struct {
		CAA                *TimeoutCAA
		ExpectedCAA        *TimeoutCAA
		ExpectedSessionCAA int64
	}{
		{
			CAA:                setTimeoutCAA(0),
			ExpectedCAA:        setTimeoutCAA(now.Unix()),
			ExpectedSessionCAA: now.Unix(),
		},
		{
			CAA:                setTimeoutCAA(1),
			ExpectedCAA:        setTimeoutCAA(1),
			ExpectedSessionCAA: now.Unix(),
		},
		{
			CAA:                setTimeoutCAA(math.MaxInt64),
			ExpectedCAA:        setTimeoutCAA(math.MaxInt64),
			ExpectedSessionCAA: now.Unix(),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			sessionCAA := test.CAA.Issue()

			assert.Equal(t, SessionCAA(test.ExpectedSessionCAA), sessionCAA)
			assert.Equal(t, test.ExpectedCAA, test.CAA)
		})
	}
}

func Test_Revoke_HasNoEffectOnUnissuedTimeoutCAA(t *testing.T) {
	caa := NewTimeout()
	caa.Revoke(10)

	assert.Equal(t, NewTimeout(), caa)
}

func Test_Revoke_ReturnsNAsNegativeTimeCAAWhenLocked(t *testing.T) {
	tests := []struct {
		CAA         *TimeoutCAA
		ExpectedCAA *TimeoutCAA
		RevokeN     int64
	}{
		{
			CAA:         setTimeoutCAA(-1),
			ExpectedCAA: setTimeoutCAA(-2),
			RevokeN:     2,
		},
		{
			CAA:         setTimeoutCAA(-4),
			ExpectedCAA: setTimeoutCAA(-10),
			RevokeN:     -10,
		},
		{
			CAA:         setTimeoutCAA(-1),
			ExpectedCAA: setTimeoutCAA(-math.MaxInt64),
			RevokeN:     math.MaxInt64,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			test.CAA.Revoke(test.RevokeN)

			assert.Equal(t, test.ExpectedCAA, test.CAA)
		})
	}
}

func Test_Revoke_ReturnsNAsTimeCAA(t *testing.T) {
	tests := []struct {
		CAA         *TimeoutCAA
		ExpectedCAA *TimeoutCAA
		RevokeN     int64
	}{
		{
			CAA:         setTimeoutCAA(1),
			ExpectedCAA: setTimeoutCAA(1),
			RevokeN:     1,
		},
		{
			CAA:         setTimeoutCAA(4),
			ExpectedCAA: setTimeoutCAA(10),
			RevokeN:     10,
		},
		{
			CAA:         setTimeoutCAA(clock.Now().Unix()),
			ExpectedCAA: setTimeoutCAA(math.MaxInt64),
			RevokeN:     math.MaxInt64,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			test.CAA.Revoke(test.RevokeN)

			assert.Equal(t, test.ExpectedCAA, test.CAA)
		})
	}
}
