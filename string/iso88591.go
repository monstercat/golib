package strutil

import (
	"bytes"

	"github.com/paulrosania/go-charset/charset"
	_ "github.com/paulrosania/go-charset/data"
)

func ToISO_8859_1(str string) (string, error) {
	buf := bytes.NewBuffer(nil)
	w, err := charset.NewWriter("iso-8859-1", buf)
	if err != nil {
		return "", err
	}
	defer w.Close()

	if _, err := w.Write([]byte(str)); err != nil {
		return "", err
	}
	return buf.String(), nil
}
