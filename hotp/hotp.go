package hotp

import (
	"crypto/hmac"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"math"
	"strings"
)

var (
	ErrInvalidSumLength = errors.New("sum must be greater than or equal to 160 bits")
)

// Generator is a generic generator for HOTP codes.
type Generator struct {
	// Hash function used to generate the HOTP codes. Note that if implementing STRICTLY for HOTP, this hash must be an
	// SHA-1 hash. However, HOTP is also used for TOTP implementations, which allows for SHA256 and SHA512.
	Hash func() hash.Hash

	// Number of digits to return as the code. This number must be between 6-8.
	NumDigits int
}

// Generate will generate an OTP from a secret and a count. In the HOTP spec, count must be an 8-byte number.
// Thus, we are using uint64
func (g *Generator) Generate(secret []byte, count uint64) (string, error) {
	// Hmac algorithm.
	mac := hmac.New(g.Hash, secret)

	// Write the byte representation of count
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, count)

	// Write to the mac.
	mac.Write(buf)
	sum := mac.Sum(nil)

	// Truncate.
	return g.Truncate(sum)
}

// Truncate implements the Dynamic Truncation algorithm in https://datatracker.ietf.org/doc/html/rfc4226#section-5.4
// This algorithm not only truncates but also returns it as a numeric string.
//
// What it does is use the lowest 4 bits of the string as the offset for selecting a location to truncate. It then
// takes 4 bytes from that byte value and applies MOD 10^(num-digits) to end up with a numeric string.
//
// In the case of HOTP, this "sum" would be a 20 byte string. However, this will not be true for SHA256 and SHA512.
// Based on the example code in https://datatracker.ietf.org/doc/html/rfc6238#page-13, what is used as offset is just
// the final digits regardless of string length.
func (g *Generator) Truncate(sum []byte) (string, error) {
	if len(sum) < 20 {
		return "", ErrInvalidSumLength
	}

	// Get the last 4 bits of the sum as the offset.
	offset := sum[len(sum)-1] & 0xf

	// The next 4 bytes (from the offset) are the required bytes of the truncated value, except that the most
	// significant bit is masked just to remove the sign.
	value := sum[offset : offset+4]
	value[0] &= 0x7F

	// Set it to BigEndian 32-byte integer
	intVal := binary.BigEndian.Uint32(value)

	// set the modulus
	intVal %= uint32(math.Pow10(g.NumDigits))

	return fmt.Sprintf(fmt.Sprintf("%%0%dd", g.NumDigits), intVal), nil
}

// Validate validates a provided code for a certain secret and count.
func (g *Generator) Validate(code string, secret []byte, count uint64) (bool, error) {
	code = strings.TrimSpace(code)
	if len(code) != g.NumDigits {
		return false, nil
	}

	expected, err := g.Generate(secret, count)
	if err != nil {
		return false, err
	}

	return code == expected, nil
}