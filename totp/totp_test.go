package totp

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"testing"
	"time"
)

// Tests are derived from https://datatracker.ietf.org/doc/html/rfc6238#appendix-B
//
//	 +-------------+--------------+------------------+----------+--------+
//	|  Time (sec) |   UTC Time   | Value of T (hex) |   TOTP   |  Mode  |
//	+-------------+--------------+------------------+----------+--------+
//	|      59     |  1970-01-01  | 0000000000000001 | 94287082 |  SHA1  |
//	|             |   00:00:59   |                  |          |        |
//	|      59     |  1970-01-01  | 0000000000000001 | 46119246 | SHA256 |
//	|             |   00:00:59   |                  |          |        |
//	|      59     |  1970-01-01  | 0000000000000001 | 90693936 | SHA512 |
//	|             |   00:00:59   |                  |          |        |
//	|  1111111109 |  2005-03-18  | 00000000023523EC | 07081804 |  SHA1  |
//	|             |   01:58:29   |                  |          |        |
//	|  1111111109 |  2005-03-18  | 00000000023523EC | 68084774 | SHA256 |
//	|             |   01:58:29   |                  |          |        |
//	|  1111111109 |  2005-03-18  | 00000000023523EC | 25091201 | SHA512 |
//	|             |   01:58:29   |                  |          |        |
//	|  1111111111 |  2005-03-18  | 00000000023523ED | 14050471 |  SHA1  |
//	|             |   01:58:31   |                  |          |        |
//	|  1111111111 |  2005-03-18  | 00000000023523ED | 67062674 | SHA256 |
//	|             |   01:58:31   |                  |          |        |
//	|  1111111111 |  2005-03-18  | 00000000023523ED | 99943326 | SHA512 |
//	|             |   01:58:31   |                  |          |        |
//	|  1234567890 |  2009-02-13  | 000000000273EF07 | 89005924 |  SHA1  |
//	|             |   23:31:30   |                  |          |        |
//	|  1234567890 |  2009-02-13  | 000000000273EF07 | 91819424 | SHA256 |
//	|             |   23:31:30   |                  |          |        |
//	|  1234567890 |  2009-02-13  | 000000000273EF07 | 93441116 | SHA512 |
//	|             |   23:31:30   |                  |          |        |
//	|  2000000000 |  2033-05-18  | 0000000003F940AA | 69279037 |  SHA1  |
//	|             |   03:33:20   |                  |          |        |
//	|  2000000000 |  2033-05-18  | 0000000003F940AA | 90698825 | SHA256 |
//	|             |   03:33:20   |                  |          |        |
//	|  2000000000 |  2033-05-18  | 0000000003F940AA | 38618901 | SHA512 |
//	|             |   03:33:20   |                  |          |        |
//	| 20000000000 |  2603-10-11  | 0000000027BC86AA | 65353130 |  SHA1  |
//	|             |   11:33:20   |                  |          |        |
//	| 20000000000 |  2603-10-11  | 0000000027BC86AA | 77737706 | SHA256 |
//	|             |   11:33:20   |                  |          |        |
//	| 20000000000 |  2603-10-11  | 0000000027BC86AA | 47863826 | SHA512 |
//	|             |   11:33:20   |                  |          |        |
//	+-------------+--------------+------------------+----------+--------+
func TestGenerate(t *testing.T) {
	secretSha1, err := hex.DecodeString("3132333435363738393031323334353637383930")
	if err != nil {
		t.Fatal(err)
	}
	secretSha256, err := hex.DecodeString("3132333435363738393031323334353637383930313233343536373839303132")
	if err != nil {
		t.Fatal(err)
	}
	secretSha512, err := hex.DecodeString("31323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334")
	if err != nil {
		t.Fatal(err)
	}

	secrets := map[string][]byte{
		"SHA1":   secretSha1,
		"SHA256": secretSha256,
		"SHA512": secretSha512,
	}

	period := 30

	tests := []struct {
		// Number of seconds since epoch
		Secs int64

		// Expected time index.
		TimeIdx uint64

		// Expected Code
		ExpectedCode string

		// Name of the hash for logging and to extract the correct secret length.
		HashName string

		// Hash function to use.
		Hash func() hash.Hash
	}{
		{
			Secs:         59,
			TimeIdx:      0x0000000000000001,
			ExpectedCode: "94287082",
			HashName:     "SHA1",
			Hash:         sha1.New,
		},
		{
			Secs:         59,
			TimeIdx:      0x0000000000000001,
			ExpectedCode: "46119246",
			HashName:     "SHA256",
			Hash:         sha256.New,
		},
		{
			Secs:         59,
			TimeIdx:      0x0000000000000001,
			ExpectedCode: "90693936",
			HashName:     "SHA512",
			Hash:         sha512.New,
		},
		{
			Secs:         1111111109,
			TimeIdx:      0x00000000023523EC,
			ExpectedCode: "07081804",
			HashName:     "SHA1",
			Hash:         sha1.New,
		},
		{
			Secs:         1111111109,
			TimeIdx:      0x00000000023523EC,
			ExpectedCode: "68084774",
			HashName:     "SHA256",
			Hash:         sha256.New,
		},
		{
			Secs:         1111111109,
			TimeIdx:      0x00000000023523EC,
			ExpectedCode: "25091201",
			HashName:     "SHA512",
			Hash:         sha512.New,
		},
		{
			Secs:         1111111111,
			TimeIdx:      0x00000000023523ED,
			ExpectedCode: "14050471",
			HashName:     "SHA1",
			Hash:         sha1.New,
		},
		{
			Secs:         1111111111,
			TimeIdx:      0x00000000023523ED,
			ExpectedCode: "67062674",
			HashName:     "SHA256",
			Hash:         sha256.New,
		},
		{
			Secs:         1111111111,
			TimeIdx:      0x00000000023523ED,
			ExpectedCode: "99943326",
			HashName:     "SHA512",
			Hash:         sha512.New,
		},
		{
			Secs:         1234567890,
			TimeIdx:      0x000000000273EF07,
			ExpectedCode: "89005924",
			HashName:     "SHA1",
			Hash:         sha1.New,
		},
		{
			Secs:         1234567890,
			TimeIdx:      0x000000000273EF07,
			ExpectedCode: "91819424",
			HashName:     "SHA256",
			Hash:         sha256.New,
		},
		{
			Secs:         1234567890,
			TimeIdx:      0x000000000273EF07,
			ExpectedCode: "93441116",
			HashName:     "SHA512",
			Hash:         sha512.New,
		},
		{
			Secs:         2000000000,
			TimeIdx:      0x0000000003F940AA,
			ExpectedCode: "69279037",
			HashName:     "SHA1",
			Hash:         sha1.New,
		},
		{
			Secs:         2000000000,
			TimeIdx:      0x0000000003F940AA,
			ExpectedCode: "90698825",
			HashName:     "SHA256",
			Hash:         sha256.New,
		},
		{
			Secs:         2000000000,
			TimeIdx:      0x0000000003F940AA,
			ExpectedCode: "38618901",
			HashName:     "SHA512",
			Hash:         sha512.New,
		},
		{
			Secs:         20000000000,
			TimeIdx:      0x0000000027BC86AA,
			ExpectedCode: "65353130",
			HashName:     "SHA1",
			Hash:         sha1.New,
		},
		{
			Secs:         20000000000,
			TimeIdx:      0x0000000027BC86AA,
			ExpectedCode: "77737706",
			HashName:     "SHA256",
			Hash:         sha256.New,
		},
		{
			Secs:         20000000000,
			TimeIdx:      0x0000000027BC86AA,
			ExpectedCode: "47863826",
			HashName:     "SHA512",
			Hash:         sha512.New,
		},
	}

	for _, test := range tests {
		secretHex := secrets[test.HashName]

		// Time to test
		tm := time.Unix(test.Secs, 0).UTC()

		actualIdx := CalculateTime(tm, Period(period))
		if actualIdx != test.TimeIdx {
			t.Errorf("For %s (%s), expected idx %d. Got %d", tm.Format(time.RFC3339), test.HashName, test.TimeIdx, actualIdx)
		}

		p := &GenerateParams{
			Hash:      test.Hash,
			NumDigits: 8,
			Period:    Period(period),
		}
		actualCode, err := Generate(secretHex, tm, p)
		if err != nil {
			t.Fatal(err)
		}
		if actualCode != test.ExpectedCode {
			t.Errorf("For %s (%s), expected code %s. Got %s", tm.Format(time.RFC3339), test.HashName, test.ExpectedCode, actualCode)
		}

		ok, err := Validate(actualCode, secretHex, tm, p)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Errorf("For %s (%s), expected code %s to be valid", tm.Format(time.RFC3339), test.HashName, actualCode)
		}

		ok, err = ValidateWithDelayWindow(actualCode, secretHex, tm, 2, p)
		if err != nil {
			return
		}

		if !ok {
			t.Errorf("For %s (%s), expected code %s to be valid with delay window", tm.Format(time.RFC3339), test.HashName, actualCode)
		}
	}
}
