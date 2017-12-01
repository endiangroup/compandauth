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
	CAA int64
}

func Test_Counter_ItConsidersUnissuedCAAsAsAlwaysInvalid(t *testing.T) {
	tests := []struct {
		Delta      int64
		SessionCAA int64
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
		SessionCAA int64
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
			entity := Entity{Delta: test.Delta, CAA: compandauth.CounterCAA(test.CAA)}

			entity.CAA = entity.CAA.Lock()

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
				newSession.CAA, entity.CAA = entity.CAA.Issue()

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
