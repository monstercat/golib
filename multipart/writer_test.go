package multipart

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"testing"
)

type Test struct {
	A int         `multipart:"a"`     // A should be 'a' in the multipart fields.
	B string      `multipart:"-"`     // B should be ignored
	C io.Reader   `multipart:"file"`  // C should be a form file
	D *TestReader `multipart:"file2"` // D is something that implements io.Reader. Should also be a form file.
}

type TestReader struct {
	data []byte
}

func (r *TestReader) Read(b []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := copy(b, r.data)
	r.data = r.data[n:]
	return n, nil
}

func TestWriter(t *testing.T) {

	test := &Test{
		A: 123,
		B: "123",
		C: bytes.NewReader([]byte("123")),
		D: &TestReader{
			data: []byte("123"),
		},
	}

	var b bytes.Buffer
	w := NewWriter(&b)

	if err := w.Marshal(test); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	r := multipart.NewReader(&b, w.Boundary())

	var numParts int
	for ; ; numParts++ {
		p, err := r.NextPart()
		if err == io.EOF {
			break
		}

		data, err := ioutil.ReadAll(p)
		if err != nil {
			t.Fatal(err)
		}

		switch p.FormName() {
		case "file":
			fallthrough
		case "file2":
			if p.Header.Get("Content-Type") != "application/octet-stream" {
				t.Fatal("Should be a file type")
			}
			fallthrough
		case "a":
			if string(data) != "123" {
				return
			}
		case "-":
			t.Fatal("- is present in the multipart form")
		}
	}

	if numParts > 3 {
		t.Error("Should only have 3 parts")
	}

}
