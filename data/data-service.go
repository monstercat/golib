package data

import (
	"errors"
	"io"
)

var (
	ErrUploadNotFound             = errors.New("upload not found")

)

type UploadStatusCode string

func OkUploadStatus() UploadStatus {
	return UploadStatus{
		Code: UploadStatusCodeOk,
	}
}
func ErrUploadStatus(err error) UploadStatus {
	return UploadStatus{
		Code:  UploadStatusCodeError,
		Error: err,
	}
}

var (
	UploadStatusCodeOk       UploadStatusCode = "Ok"
	UploadStatusCodeProgress UploadStatusCode = "Progress"
	UploadStatusCodeError    UploadStatusCode = "Error"
)

type UploadStatus struct {
	Code    UploadStatusCode
	Message string
	Error   error
}

type Service interface {
	Exists(filepath string) (bool, error)
	Get(filepath string) (io.ReadCloser, error)
	Download(filepath string, w io.WriterAt) error
	Delete(filepath string) error

	Put(filepath string, filesize int, chunks int, r io.Reader) error
}

type Upload interface {
	GetFilepath() string
	GetUploaded() int
}

type ChunkService interface{
	Service
	PutWithStatus(filepath string, filesize int, chunks int, r io.Reader) chan UploadStatus
	ResumePutWithStatus(filepath string, offset int, r io.Reader) (chan UploadStatus, error)
	GetIncompleteUpload(filepath string) Upload
	GetMinChunkSize() int
}