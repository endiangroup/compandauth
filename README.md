# CAA: Compare-and-Authenticate

[![Build Status](https://travis-ci.org/endiangroup/compandauth.svg?branch=master)](https://travis-ci.org/endiangroup/compandauth) [![Coverage Status](https://coveralls.io/repos/github/endiangroup/compandauth/badge.svg?branch=master)](https://coveralls.io/github/endiangroup/compandauth?branch=master) [![GoDoc](https://godoc.org/github.com/endiangroup/compandauth?status.svg)](https://godoc.org/github.com/endiangroup/compandauth) [![Go Report Card](https://goreportcard.com/badge/github.com/endiangroup/compandauth)](https://goreportcard.com/report/github.com/endiangroup/compandauth)

A single counter used to maintain the validity of a set number of distributed sessions. Inspired by CAS.

*For a more in depth look at how it was conceived and works, see: https://endian.io/articles/compandauth/*

### Features:

- *Fast; Really fast*, Nanosecond Issuing and validating of sessions.
- *Central revocation, locking and unlocking* of distributed sessions
- Tiny, *single int64 to be stored along with the entity* you wish to protect and *single int64 to store inside existing distributed session* (such as JWT or Cookie)
- [**Counter**] Can *maintain a number of concurrent active sessions* (lets say you want to allow a user to be able to login from 5 different browsers, or 1)
- [**Counter**] Can *dynamically change the number of concurrent sessions* server side with no need to update the distributed session
- [**Counter**] Can be *shoe horned into an existing system easily*, JWT's that don't contain a 'SessionCAA' value can be considered to have a 'SessionCAA' of '0' which is the first valid issued number
- [**Counter**] *Long lived sessions*, such as for mobile apps
- [**Timeout**] Can *manage the validity of a session based on some duration*
- [**Timeout**] Can *dynamically adjust the validity duration* server side
- [**Timeout**] Can *revoke all sessions before some timestamp* regardless if they are still within the valid duration or not

**What it doesn't do:**

- Lock or unlock sessions individually
	- Instead you'll lock an entity from doing what ever behaviour you have the CAA protecting, such as logging in or escalating privileges for example.
- Revoke sessions individually
	- [**Counter**] You can revoke the last N sessions but not a specific one
	- [**Timeout**] You can revoke all sessions before timestamp T
- Audit trail
	- No in built mechanism for tracking changes to CAA values, must be performed at a higher level
- Signing
	- You **MUST** be able to trust the incoming session CAA value, as such your session mechanism must at least sign its payload including the session CAA

### What problems does this package solve?

> You're building a service that allows some entity to authenticate against, however you want to limit the number of concurrent sessions it can maintain and centrally manage validity of issued tokens.

Vanilla JWT or Cookies (that is without a bulky server side session management system) don't have a mechanism for limiting the number of concurrent sessions a single entity may have. For example with a JWT or Cookie you can't say a single entity such as a user can only have 2 active sessions open at any time.

Additionally Cookies and JWT's cannot revoke access for already issued tokens. You can't for instance temporarily lock out all sessions for a given entity or revoke already issued sessions. For example a user wants to invalidate all their active sessions across devices, or internally you want to lock a users account temporarily whilst you investigate something.

**Possible solution:**

With the `Counter` you can do both of these things server side without having to touch already issued sessions. You add a `SessionCAA` to the existing struct you issue to your authenticating entites and a `CAA` implementation to the entity you want to protect.

---
> You're building a service that allows some entity to escalate its privileges, however you want it to do so only for some period of time, additionally you may want to increase that period of time during its lifetime

Both Cookies and JWT's support expiration times, however you can't increase an issued tokens expiration time without trading the token with the device that holds it (e.g. wait until the user makes a request to the server so you can trade the token with a new one with an increased expiration timestamp). For example when your user edits their settings you have them re-authenticate to escalate their privileges for a limited period of time, whilst the session is being used you keep the session alive until some fixed deadline.

**Possible solution:**

With the `Timeout` you can do all of these things with a combination of adjusting the `IsValid` duration and using the `Revoke` to set a hard deadline. You add a `SessionCAA` to the existing struct you issue to your authenticating entites and a `CAA` implementation to the entity you want to protect.

### Performance

Hot paths are blazingly fast, this package won't be the slowest link in the chain.

```
$ go test -run=^$ -bench=.
goos: darwin
goarch: amd64
pkg: github.com/endiangroup/compandauth
Benchmark_Counter_Issue-8       1000000000               2.88 ns/op
Benchmark_Counter_IsValid-8     200000000                7.08 ns/op
Benchmark_Timeout_IsValid-8     100000000               14.2 ns/op
Benchmark_Timeout_Issue-8       20000000                92.0 ns/op
PASS
ok      github.com/endiangroup/compandauth      8.785s
```

### Status

**Counter** - A previous incarnation has been used successfully in production with 15,000+ users since December 2016.
**Timeout** - Has not been used in a production environment that we are aware of yet.

### Usage:

- The `CAA` type is added to the entity being protected (e.g. user)
- A `SessionCAA` property is added to the session object (e.g. JWT)
- The session payload must be at least signed or encrypted
- When validating the session object, fetch the entity in question and check the validity of the incoming `SessionCAA` with `entity.CAA.IsValid(SessionCAA)`
- When issuing a new session for the entity set the sessions CAA value with `session.CAA = entity.CAA.Issue()`
- Ensure you update the entity after using `Revoke()`, `Issue()`, `Lock()` and `Unlock()` as they modify the CAA state

### Synchronisation

As this package was inspired by CAS, which itself is a synchronisation primitive, you do have to consider synchronisation. There are 3 situations that should be considered when using this package:

1. [Unlikely] is multiple goroutines during a single request, where you may spin off goroutines during the authentication flow, for that you can use the `caa.ThreadSafe` wrapper
2. [Likely] is a goroutine per request, where each incoming request gets a new goroutine, in that instance you should row level lock your entity for the duration of the authentication flow. (e.g. when fetching the User record, lock the User row [or ideally just their CAA] until you've ascertained the validity of their session or finished manipulating their CAA state)
3. [Likely] is multi-server, where there is a shared database between multiple servers storing the CAA value for an entity (e.g. horizontally scaled API servers calling a central SQL DB). see 2


You can get more specific read and write locking to increase performance, but We'll leave that to you to decide what works in your environment. See the `ThreadSafe` wrapper to understand when you need read and write locks.

### Examples:

**JWT Login**:

``` go
type User struct {
	//...
	compandauth.CAA
}

type JwtSession struct {
	jwt.StandardClaims
	CAA SessionCAA `json:"caa"`
}

func Login(incomingUsername, incomingPassword string) (JwtSession, error) {
	//... fetch the User ...
	if passwordsMatch(incomingPassword, user.Password) {
		newUserSession := JwtSession{...} // set standard claims

		newUserSession.CAA = user.CAA.Issue()

		if err := user.Update(); err != nil { // update user record with new issued CAA value
			return JwtSession{}, err
		}

		return newUserSession, nil
	}

	return JwtSession{}, errors.New("User login failed")
}
```

**JWT Counter Authentication**:

```go
type User struct {
	//...
	MaxActiveSessions uint
	CAA               compandauth.CAA
}

type JwtSession struct {
	jwt.StandardClaims
	CAA SessionCAA `json:"caa"`
}

func (j JwtSession) Valid() error {
	//... fetch the User from the session ...
	if !user.CAA.IsValid(j.CAA, user.MaxActiveSessions) {

		if user.CAA.IsLocked() {
			return errors.New("It appears your account has been locked")
		}

		return errors.New("Invalid session, please login again")
	}

	return nil
}
```

**JWT Timeout Authentication**:

```go
const SudoTimeout = 5 * time.Minute

type User struct {
	//...
	SudoCAA compandauth.CAA
}

type SudoSession struct {
	JwtSession
	SudoCAA SessionCAA `json:"sudo_caa"`
}

func (s SudoSession) Valid() error {
	if err := s.JwtSession.Valid(); err != nil {
		return err
	}

	//... fetch the User from the session ...
	if !user.SudoCAA.IsValid(s.SudoCAA, compandauth.ToSeconds(SudoTimeout)) {

		if user.SudoCAA.IsLocked() {
			return errors.New("It appears your locked out of sudo mode")
		}

		return errors.New("Invalid session, please login again")
	}

	return nil
}
```

**Locking**:

```go
type User struct {
	//...
	CAA compandauth.CAA
}

func (u *User) Lock() {
	u.CAA.Lock()
}
```

**Counter Revocation**:

```go
type User struct {
	//...
	MaxActiveSessions uint
	CAA               compandauth.CAA
}

func (u *User) LogoutAllSessions() {
	u.CAA.Revoke(u.MaxActiveSessions)
}
```

**Timeout Revocation**:

```go
type User struct {
	//...
	CAA compandauth.CAA
}

func (u *User) LogoutAllSessions() {
	u.CAA.Revoke(time.Now().Unix())
}
```

**Has Ever Logged In**

```go
type User struct {
	//...
	CAA compandauth.CAA
}

func (u *User) HasLoggedIn() bool {
	return u.CAA.HasIssued()
}
```
