package compandauth_test

import (
	"fmt"
	"testing"

	"github.com/adrianduke/compandauth"
	"github.com/stretchr/testify/assert"
)

type Entity struct {
	Delta int64
	CAA   compandauth.CAA
}

type Session struct {
	CAA compandauth.SessionCAA
}

func setCounterCAA(i int64) *compandauth.CounterCAA {
	caa := compandauth.NewCounter()
	*caa = compandauth.CounterCAA(i)

	return caa
}

func Test_Counter_ItConsidersUnissuedCAAsAsAlwaysInvalid(t *testing.T) {
	tests := []struct {
		Delta      int64
		SessionCAA compandauth.SessionCAA
	}{
		{Delta: -1, SessionCAA: -1},
		{Delta: -1, SessionCAA: 0},
		{Delta: -1, SessionCAA: 1},
		{Delta: 0, SessionCAA: -1},
		{Delta: 0, SessionCAA: 0},
		{Delta: 0, SessionCAA: 1},
		{Delta: 1, SessionCAA: -1},
		{Delta: 1, SessionCAA: 0},
		{Delta: 1, SessionCAA: 1},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			entity := Entity{Delta: test.Delta, CAA: compandauth.NewCounter()}
			session := Session{CAA: test.SessionCAA}

			assert.False(t, entity.CAA.IsValid(session.CAA, entity.Delta))
		})
	}
}

func Test_Counter_ItConsidersLockedCAAAsAlwaysInvalid(t *testing.T) {
	tests := []struct {
		CAA        int64
		Delta      int64
		SessionCAA compandauth.SessionCAA
	}{
		{CAA: -1, Delta: -1, SessionCAA: -1},
		{CAA: -1, Delta: -1, SessionCAA: 0},
		{CAA: -1, Delta: -1, SessionCAA: 1},
		{CAA: -1, Delta: 0, SessionCAA: -1},
		{CAA: -1, Delta: 0, SessionCAA: 0},
		{CAA: -1, Delta: 0, SessionCAA: 1},
		{CAA: -1, Delta: 1, SessionCAA: -1},
		{CAA: -1, Delta: 1, SessionCAA: 0},
		{CAA: -1, Delta: 1, SessionCAA: 1},
		{CAA: 1, Delta: -1, SessionCAA: -1},
		{CAA: 1, Delta: -1, SessionCAA: 0},
		{CAA: 1, Delta: -1, SessionCAA: 1},
		{CAA: 1, Delta: 0, SessionCAA: -1},
		{CAA: 1, Delta: 0, SessionCAA: 0},
		{CAA: 1, Delta: 0, SessionCAA: 1},
		{CAA: 1, Delta: 1, SessionCAA: -1},
		{CAA: 1, Delta: 1, SessionCAA: 0},
		{CAA: 1, Delta: 1, SessionCAA: 1},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			session := Session{CAA: test.SessionCAA}
			entity := Entity{Delta: test.Delta, CAA: setCounterCAA(test.CAA)}

			entity.CAA.Lock()

			assert.False(t, entity.CAA.IsValid(session.CAA, entity.Delta))
		})
	}
}

func Test_Counter_OnlyTheLastDeltaSessionsAreConsideredValid(t *testing.T) {
	tests := []struct {
		Delta int64
		N     int
	}{
		{Delta: 0, N: 10},
		{Delta: 1, N: 100},
		{Delta: 5, N: 1000},
		{Delta: 10, N: 10000},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			entity := Entity{Delta: test.Delta, CAA: compandauth.NewCounter()}
			sessions := []Session{}

			for i := 0; i < test.N; i++ {
				newSession := Session{}
				newSession.CAA = entity.CAA.Issue()

				sessions = append(sessions, newSession)

				for j, session := range sessions {
					if int64(len(sessions)-j) > entity.Delta {
						assert.False(t, entity.CAA.IsValid(session.CAA, entity.Delta))
					} else {
						assert.True(t, entity.CAA.IsValid(session.CAA, entity.Delta))
					}
				}
			}
		})
	}
}

func Test_Counter_ItRevokesTheLastNSessions(t *testing.T) {
	tests := []struct {
		Delta            int64
		NumberOfSessions int64
		RevokeN          int64
	}{
		{Delta: 0, NumberOfSessions: 0, RevokeN: 0},
		{Delta: 0, NumberOfSessions: 0, RevokeN: 1},
		{Delta: 0, NumberOfSessions: 1, RevokeN: 1},
		{Delta: 1, NumberOfSessions: 1, RevokeN: 0},
		//{Delta: 1, NumberOfSessions: 0, RevokeN: 0},
		{Delta: 1, NumberOfSessions: 0, RevokeN: 1},
		{Delta: 1, NumberOfSessions: 1, RevokeN: 1},
		{Delta: 1, NumberOfSessions: 2, RevokeN: 2},
		{Delta: 10, NumberOfSessions: 10, RevokeN: 1},
		{Delta: 10, NumberOfSessions: 10, RevokeN: 2},
		{Delta: 10, NumberOfSessions: 10, RevokeN: 3},
		{Delta: 10, NumberOfSessions: 10, RevokeN: 4},
		{Delta: 10, NumberOfSessions: 10, RevokeN: 5},
		{Delta: 10, NumberOfSessions: 10, RevokeN: 6},
		{Delta: 10, NumberOfSessions: 10, RevokeN: 7},
		{Delta: 10, NumberOfSessions: 10, RevokeN: 8},
		{Delta: 10, NumberOfSessions: 10, RevokeN: 9},
		{Delta: 10, NumberOfSessions: 10, RevokeN: 10},
		{Delta: 10, NumberOfSessions: 10, RevokeN: 11},
		{Delta: 5, NumberOfSessions: 100, RevokeN: 50},
		{Delta: 5, NumberOfSessions: 100, RevokeN: 99},
		{Delta: 5, NumberOfSessions: 100, RevokeN: 150},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			entity := Entity{Delta: test.Delta, CAA: compandauth.NewCounter()}
			sessions := []Session{}

			for i := int64(0); i < test.NumberOfSessions; i++ {
				newSession := Session{}
				newSession.CAA = entity.CAA.Issue()

				sessions = append(sessions, newSession)
			}

			entity.CAA.Revoke(test.RevokeN)

			lastRevokedSessionOffset := min(
				test.NumberOfSessions-entity.Delta+test.RevokeN,
				test.NumberOfSessions,
			)
			expectedNumberOfValidSessions := max(
				entity.Delta-test.RevokeN,
				0,
			)
			assert.Len(t, sessions[lastRevokedSessionOffset:], int(expectedNumberOfValidSessions))

			for _, session := range sessions[lastRevokedSessionOffset:] {
				assert.True(t, entity.CAA.IsValid(session.CAA, entity.Delta))
			}

			for _, session := range sessions[:lastRevokedSessionOffset] {
				assert.False(t, entity.CAA.IsValid(session.CAA, entity.Delta))
			}

		})
	}
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
