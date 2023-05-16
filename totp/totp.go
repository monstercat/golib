package totp

import (
	"crypto/sha1"
	"hash"
	"math"
	"time"

	"github.com/monstercat/golib/hotp"
)

type Period int

var (
	Period30 Period = 30
	Period60 Period = 60
)

// GenerateParams are parameters for generating a code.
type GenerateParams struct {
	// Hash function used to generate the TOTP codes. This is normally SHA-1, SHA-256 or SHA-512
	Hash func() hash.Hash

	// Number of digits to return as the code. This number must be between 6-8.
	NumDigits int

	// Period is the number of seconds per OTP code.
	Period Period
}

// ApplyDefaults applies default values for the params
func (p *GenerateParams) ApplyDefaults() {
	if p.Hash == nil {
		p.Hash = sha1.New
	}
	if p.NumDigits < 6 || p.NumDigits > 8 {
		p.NumDigits = 6
	}
	if p.Period != 15 && p.Period != 30 && p.Period != 60 {
		p.Period = 30
	}
}

// CalculateTime calculates the time "index" based on the provided period.
func CalculateTime(t time.Time, period Period) uint64 {
	return uint64(math.Floor(float64(t.Unix()) / float64(period)))
}

// Generate generates a code for a certain time value.
func Generate(secret []byte, t time.Time, p *GenerateParams) (string, error) {
	p.ApplyDefaults()

	count := CalculateTime(t, p.Period)
	gen := &hotp.Generator{
		Hash:      p.Hash,
		NumDigits: p.NumDigits,
	}
	return gen.Generate(secret, count)
}

// ValidateWithDelayWindow validates the provided code with a delay window which is defined by window. Each offset
// is a single period. For example, if window = 2 and Period = 30, it will check codes within the last 60 seconds.
func ValidateWithDelayWindow(code string, secret []byte, t time.Time, window int, p *GenerateParams) (bool, error) {
	p.ApplyDefaults()

	count := CalculateTime(t, p.Period)
	gen := &hotp.Generator{
		Hash:      p.Hash,
		NumDigits: p.NumDigits,
	}

	if ok, err := gen.Validate(code, secret, count); ok {
		return true, nil
	} else if err != nil {
		return false, err
	}

	for i := 1; i < window; i++ {
		if count < uint64(i) {
			return false, nil
		}
		ok, err := gen.Validate(code, secret, count-uint64(i))
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// Validate validates a code for a specific time value.
func Validate(code string, secret []byte, t time.Time, p *GenerateParams) (bool, error) {
	p.ApplyDefaults()

	count := CalculateTime(t, p.Period)
	gen := &hotp.Generator{
		Hash:      p.Hash,
		NumDigits: p.NumDigits,
	}
	return gen.Validate(code, secret, count)
}
