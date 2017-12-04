# CAA: Compare-and-Authenticate

[![Build Status](https://travis-ci.org/endiangroup/compandauth.svg?branch=master)](https://travis-ci.org/endiangroup/compandauth) [![Coverage Status](https://coveralls.io/repos/github/endiangroup/compandauth/badge.svg?branch=master)](https://coveralls.io/github/endiangroup/compandauth?branch=master) [![GoDoc](https://godoc.org/github.com/endiangroup/compandauth?status.svg)](https://godoc.org/github.com/endiangroup/compandauth)

A single counter used to maintain the validity of a set number of distributed sessions. Inspired by CAS counters.

### Features:

- Central revocation, locking and unlocking of distributed sessions
- Tiny, single int64 to be stored along with the entity you wish to protect and single int64 to store inside distributed session
- Can maintain a number of concurrent active sessions (lets say you want to allow a user to be able to login from 5 different browsers, or 1)
- Can dynamically change the number of concurrent sessions with no need to update the distributed session
- Can be shoe horned into an existing system easily, JWT's that don't contain a 'CAA' value can be considered to have a 'CAA' of '0' which is the first valid issued number
- Long lived sessions, such as for mobile apps
- Naturally acts as a 'number of logins' counter
- Doubles as a nonce, as every issued session will have a unique CAA value

**What it doesn't do:**

- Lock or unlock sessions individually (you would need a CAA per thing you'd want to manage e.g. laptop sessions, mobile sessions... etc)
- Revoke sessions individually (again would need individual CAA's)
- Audit trails
- Time limited sessions

### CAA Pre-requisites:

- A column or property is added to the entity being protected (e.g. a user), of equivalent type BIGINT, defaulted to 0
- The `CAA` type is added to the user entity object in code
- A `CAA` claim is added to the session object (e.g. JWT) in code, of type int64 (don't use the CAA type provided by this package!)

### CAA Usage:

JWT **Login**:

1. Whilst preparing a JWT for a successfully authenticated user, store the current user CAA value in the session CAA claim
2. Increment the user CAA value by 1
3. Update user record and issue JWT

Example code, `Issue()` returns both the incremented CAA and current CAA value:
``` go
type User struct {
	//...
	compandauth.CAA
}

type JwtSession struct {
	jwt.StandardClaims
	CAA int64 `json:"caa"`
}

func Login(incomingUsername, incomingPassword string) (JwtSession, error) {
	//... fetch the User ...
	if passwordsMatch(incomingPassword, user.Password) {
		newUserSession := JwtSession{...} // set standard claims

		newUserSession.CAA, user.CAA = user.CAA.Issue()

		if err := user.Update(); err != nil { // update user record with new issued CAA value
			return JwtSession{}, err
		}

		return newUserSession, nil
	}

	return JwtSession{}, errors.New("User login failed")
}
```

**Authentication**:

1. Extract the CAA value from the incoming session
2. Fetch the user record related to the session
3. Compare (session CAA + maximum concurrent sessions) >= user CAA
	1. _True_: session can be considered valid
	2. _False_: session can be considered invalid

```go
type User struct {
	//...
	MaxActiveSessions uint
	CAA               compandauth.CAA
}

type JwtSession struct {
	CAA int64 `json:"caa"`
	jwt.StandardClaims
}

func (j JwtSession) Valid() error {
	//... fetch the User from the session ...
	if !user.CAA.IsValid(j.CAA, user.MaxActiveSessions) {

		if user.CAA.IsLocked() {
			return errors.New("It appears your account has been locked")
		}

		return errors.New("Invalid session, please login again")
	}
}
```

**Locking**:

1. Fetch the user record for which you want to lock
2. Flip the signed bit on the user CAA (turn the positive integer into a negative one)
3. Update the user record

```go
type User struct {
	//...
	CAA compandauth.CAA
}

func (u *User) Lock() {
	u.CAA = u.CAA.Lock()
}
```

**Revocation**:

1. Fetch the user record for which you want to force logout
2. Increment the user CAA by the maximum number of concurrent sessions they can have
3. Update the user record

```go
type User struct {
	//...
	MaxActiveSessions uint
	CAA               compandauth.CAA
}

func (u *User) LogoutAllSessions() {
	u.CAA = u.CAA.Revoke(u.MaxActiveSessions)
}
```

**Has Ever Logged In**

1. Fetch the user record for which you want to check
2. Check if CAA is not 0

```go
type User struct {
	//...
	CAA compandauth.CAA
}

func (u *User) HasLoggedIn() bool {
	return u.CAA.HasIssued()
}
```

### Advanced Scenarios:

**Blessed session**

Typical usecase would be to have a user re-login to access their account settings, like GitHub or LinkedIn.

Add an additional CAA to your entity (e.g. `BlessedCAA compandauth.CAA`) separate to your regular login CAA. In the handler that deals with promoting privileges, have it issue your BlessedCAA with a background job to `Revoke` after some period of time.
