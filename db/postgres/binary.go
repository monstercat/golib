package pgUtils

import (
	"database/sql/driver"
	"errors"
)

var (
	ErrByteFormatError = errors.New("unexpected format. Expecting a bit array")
)

// Postgres Binary requires a specific format when writing and reading.
// This is used to write and read it properly.
type Byte uint8

func (i *Byte) Scan(value interface{}) error {
	byt, ok := value.([]byte)
	if !ok {
		return ErrByteFormatError
	}
	*i = 0
	for idx, b := range byt {
		switch b {
		case '0':
		case '1':
			*i |= 1 << (7-idx)
		default:
			return ErrByteFormatError
		}
	}

	return nil
}

func (i Byte) Value() (driver.Value, error) {
	b := make([]byte, 0, 8)
	ii := byte(i)
	for test := byte(1 << 7); test > 0; test = test >> 1 {
		if ii >= test {
			ii -= test
			b = append(b, '1')
		} else {
			b = append(b, '0')
		}
	}
	return b, nil
}
