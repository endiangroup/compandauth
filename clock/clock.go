package clock

import "time"

var (
	// Force to UTC so no possible timzone conflicts across servers
	Now = func() time.Time { return time.Now().UTC() }
)

func NowForce(t time.Time) {
	Now = func() time.Time { return t }
}

func NowReset() {
	Now = func() time.Time { return time.Now().UTC() }
}
