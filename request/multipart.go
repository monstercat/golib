package request

import (
	"bytes"
)

// MultipartWriter wraps bytes.Buffer to make it a custom payload.
type MultipartBuffer struct {
	bytes.Buffer
	Boundary string
}

func (w *MultipartBuffer) ContentType() string {
	return "multipart/form-data; boundary=" + w.Boundary
}