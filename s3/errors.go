package s3util

import "errors"

var (
	ErrUploadNotFound             = errors.New("upload not found")
	ErrResumedUploadOffsetInvalid = errors.New("offset does not match uploaded status")
	ErrTimeout                    = errors.New("timeout")
)

type InitError struct {
	error error
}
func (e InitError) Error() string {
	return "upload initialization error: " + e.error.Error()
}

type UploadError struct {
	error error
}
func (e UploadError) Error() string {
	return "upload part error: " + e.error.Error()
}