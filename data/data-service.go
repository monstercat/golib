package data

import (
	"errors"
	"io"
	"time"
)

var (
	ErrUploadNotFound = errors.New("upload not found")
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
	Put(filepath string, r io.Reader) error

	Delete(filepath string) error
}

type HeadService interface {
	Head(filepath string) (*HeadInfo, error)
}

type SignedUrlService interface {
	SignedUrl(filepath string, tm time.Duration, cfg *SignedUrlConfig) (string, error)
}

type HeadInfo struct {
	Exists        bool
	LastModified  time.Time
	ContentLength int64
}

type Upload interface {
	GetFilepath() string
	GetUploaded() int
}

type ChunkService interface {
	Service
	PutWithStatus(filepath string, filesize int, chunks int, r io.Reader) chan UploadStatus
	ResumePutWithStatus(filepath string, offset int, r io.Reader) (chan UploadStatus, error)
	GetIncompleteUpload(filepath string) Upload
	GetChunkSize() int
}

type ParallelService interface {
	Download(filepath string, w io.WriterAt) error
	Upload(filepath string, filesize int, chunks int, r io.Reader) error
}
