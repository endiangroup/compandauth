package compandauth_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/endiangroup/compandauth"
	"github.com/endiangroup/compandauth/clock"
)

var isValidResult bool
var sessionCAAResult compandauth.SessionCAA

func Benchmark_Counter_Issue(b *testing.B) {
	b.StopTimer()
	entity := CounterEntity{Delta: 10, CAA: compandauth.NewCounter()}

	var sessionCAA compandauth.SessionCAA
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sessionCAA = entity.CAA.Issue()
	}

	sessionCAAResult = sessionCAA
}

func Benchmark_Counter_IsValid(b *testing.B) {
	b.StopTimer()
	entity := CounterEntity{Delta: 10, CAA: compandauth.NewCounter()}
	numberOfSessions := 100000

	sessions := []Session{}
	for i := 0; i < numberOfSessions; i++ {
		newSession := Session{}
		newSession.CAA = compandauth.SessionCAA(rand.Intn(numberOfSessions))

		if rand.Intn(numberOfSessions)%2 == 0 {
			entity.CAA.Issue()

			// Randomly lock ~50% of the sessions
			newSession.CAA = compandauth.SessionCAA(-1) * newSession.CAA
		}

		sessions = append(sessions, newSession)
	}

	var isValid bool
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		isValid = entity.CAA.IsValid(sessions[i%numberOfSessions].CAA, entity.Delta)
	}

	isValidResult = isValid
}

func Benchmark_Timeout_IsValid(b *testing.B) {
	b.StopTimer()

	start := time.Now()
	clock.NowForce(start)
	defer clock.NowReset()

	entity := TimeoutEntity{Timeout: 30 * time.Second, CAA: compandauth.NewTimeout()}
	numberOfSessions := 100000
	offset := start.Unix() - int64(numberOfSessions/2)

	sessions := []Session{}
	for i := 0; i < numberOfSessions; i++ {
		newSession := Session{}
		newSession.CAA = compandauth.SessionCAA(offset + int64(rand.Intn(numberOfSessions)))

		if rand.Intn(numberOfSessions)%2 == 0 {
			entity.CAA.Issue()

			// Randomly lock ~50% of the sessions
			newSession.CAA = compandauth.SessionCAA(-1) * newSession.CAA
		}

		sessions = append(sessions, newSession)
	}

	var isValid bool
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		isValid = entity.CAA.IsValid(sessions[i%numberOfSessions].CAA, compandauth.ToSeconds(entity.Timeout))
	}

	isValidResult = isValid
}

func Benchmark_Timeout_Issue(b *testing.B) {
	b.StopTimer()
	entity := TimeoutEntity{Timeout: 30 * time.Second, CAA: compandauth.NewTimeout()}

	var sessionCAA compandauth.SessionCAA
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sessionCAA = entity.CAA.Issue()
	}

	sessionCAAResult = sessionCAA
}
