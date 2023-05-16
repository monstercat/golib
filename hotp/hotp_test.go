package hotp

import (
	"crypto/sha1"
	"encoding/hex"
	"testing"
)

// Test values for HOTP are derived from https://datatracker.ietf.org/doc/html/rfc4226#page-32
func TestGenerate(t *testing.T) {
	secretHex, err := hex.DecodeString("3132333435363738393031323334353637383930")
	if err != nil {
		t.Fatal(err)
	}

	gen := &Generator{
		Hash:      sha1.New,
		NumDigits: 6,
	}

	hotpVals := []string{
		"755224",
		"287082",
		"359152",
		"969429",
		"338314",
		"254676",
		"287922",
		"162583",
		"399871",
		"520489",
	}
	for i, v := range hotpVals {
		actual, err := gen.Generate(secretHex, uint64(i))
		if err != nil {
			t.Fatal(err)
		}
		if actual != v {
			t.Errorf("For %d, expect %s. Got %s", i, v, actual)
		}
	}
}

// Test values for HOTP are derived from https://datatracker.ietf.org/doc/html/rfc4226#page-32
func TestTruncate(t *testing.T) {
	testVals := []struct {
		Sum   string
		Value string
	}{
		{
			Sum:   "cc93cf18508d94934c64b65d8ba7667fb7cde4b0",
			Value: "755224",
		},
		{
			Sum:   "75a48a19d4cbe100644e8ac1397eea747a2d33ab",
			Value: "287082",
		},
		{
			Sum:   "0bacb7fa082fef30782211938bc1c5e70416ff44",
			Value: "359152",
		},
		{
			Sum:   "66c28227d03a2d5529262ff016a1e6ef76557ece",
			Value: "969429",
		},
		{
			Sum:   "a904c900a64b35909874b33e61c5938a8e15ed1c",
			Value: "338314",
		},
		{
			Sum:   "a37e783d7b7233c083d4f62926c7a25f238d0316",
			Value: "254676",
		},
		{
			Sum:   "bc9cd28561042c83f219324d3c607256c03272ae",
			Value: "287922",
		},
		{
			Sum:   "a4fb960c0bc06e1eabb804e5b397cdc4b45596fa",
			Value: "162583",
		},
		{
			Sum:   "1b3c89f65e6c9e883012052823443f048b4332db",
			Value: "399871",
		},
		{
			Sum:   "1637409809a679dc698207310c8c7fc07290d9e5",
			Value: "520489",
		},
	}

	gen := &Generator{
		NumDigits: 6,
	}

	for _, val := range testVals {
		hexVal, err := hex.DecodeString(val.Sum)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := gen.Truncate(hexVal)
		if err != nil {
			t.Fatal(err)
		}
		if actual != val.Value {
			t.Errorf("For %x, expect %s. Got %s", hexVal, val.Value, actual)
		}
	}
}

func TestValidate(t *testing.T) {
	secretHex, err := hex.DecodeString("3132333435363738393031323334353637383930")
	if err != nil {
		t.Fatal(err)
	}

	gen := &Generator{
		Hash:      sha1.New,
		NumDigits: 6,
	}

	validVals := []string{
		"755224",
		"287082",
		"359152",
		"969429",
		"338314",
		"254676",
		"287922",
		"162583",
		"399871",
		"520489",
	}
	for i, v := range validVals {
		ok, err := gen.Validate(v, secretHex, uint64(i))
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Errorf("Expected %s to be the code for %d.", v, i)
		}
	}

	invalidVals := []string{
		"755225",
		"287182",
		"357152",
		"963429",
		"331314",
		"252676",
		"282922",
		"165583",
		"394871",
		"522489",
	}
	for i, v := range invalidVals {
		ok, err := gen.Validate(v, secretHex, uint64(i))
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Errorf("Expected %s to be an invalid code for %d.", v, i)
		}
	}
}