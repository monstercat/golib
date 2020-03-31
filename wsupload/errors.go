package wsupload

import "errors"

var (
	ErrInvalidMessageType = errors.New("invalid message type")
	ErrInitUpload = errors.New("could not initialize upload")
	ErrResumeUpload = errors.New("could not resume upload")
)

type DataError struct {
	error error
}

func (e DataError) Error() string {
	return "could not receive data for upload: " + e.error.Error()
}
