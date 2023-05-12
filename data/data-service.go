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

// ServiceWithTimeout allows a timeout to be set for a service. Typically, this should *not* be included in the service
// description, but optionally set if the service allows it.
type ServiceWithTimeout interface {
	SetTimeout(tm time.Duration)
}

// ServiceWithConcurrency allows a concurrency to be set for the service. This should be used for services which
// implement ParallelService. Typically, this should *not* be included in the service description, but optionally set if
// the service allows it.
type ServiceWithConcurrency interface {
	SetConcurrency(concurrency int)
}

// ServiceWithMimeType allows to set the MimeType/ContentType for a specific file that has been uploaded.
// This service is optional.
type ServiceWithMimeType interface {
	SetMimeType(filepath, mimeType string) error
}

// Service defines a basic data service. It allows the user to Get, Put, Delete, and check that a certain piece of
// data exists. Categorization is by filepath (not necessarily / delimited). To use this service, encapsulate it in
// another interface.
//
// For example:
//
//	type FileService interface {
//	    data.Service
//	    data.HeadService
//	    // ....
//	}
type Service interface {
	Exists(filepath string) (bool, error)
	Get(filepath string) (io.ReadCloser, error)
	Put(filepath string, r io.Reader) error

	Delete(filepath string) error
}

// HeadService provides a way for a data service to return statistical information such as the content length and
// last modified date from the data service.
type HeadService interface {
	Head(filepath string) (*HeadInfo, error)
}

// SignedUrlService allows a service to return a signed URL for a filepath.
type SignedUrlService interface {
	SignedUrl(filepath string, tm time.Duration, cfg *SignedUrlConfig) (string, error)
}

// HeadInfo is the information returned by the HeadService.
type HeadInfo struct {
	Exists        bool
	LastModified  time.Time
	ContentLength int64
}

// Upload refers to an instance of upload. It allows for chunked and resumable uploads.
type Upload interface {
	GetFilepath() string
	GetUploaded() int
}

// ChunkService provides the methods required to upload a file in chunks.
type ChunkService interface {
	PutWithStatus(filepath string, filesize int, r io.Reader) chan UploadStatus
	ResumePutWithStatus(filepath string, r io.Reader) (chan UploadStatus, error)
	GetIncompleteUpload(filepath string) Upload
	GetChunkSize() int
}

// ChunkManagementService is an interface that defines an uploader which is managed. This means that the uploader runs
// a goroutine which receives parts of files through a channel and uploads them sequentially, rather than pushing the
// updates directly. By doing this, we can be certain that parts are uploaded sequentially.
type ChunkManagementService interface {
	RunUploader()
	Shutdown()
}

// DownloadParams are params specific to ParallelService.Download. The implementation should be able to handle null
// properly. These should just be custom values and input as needed.
type DownloadParams struct {
	// A custom concurrency.
	Concurrency int
}

// ParallelService allows upload and download of files in parallel. This is not to be confused with uploading in chunks.
// Uploading in chunks assumes that the provided reader only has a part of the information, whereas for parallel Upload,
// the reader contains all the data to be uploaded. Only the act of uploading is done in paralle.
//
// A parallel service should also satisfy the `ServiceWithConcurrency` interface.
type ParallelService interface {
	Download(filepath string, w io.WriterAt, p *DownloadParams) error
	Upload(filepath string, filesize int, r io.Reader) error
}

// RangeService allows for a part of the file to be downloaded. Dictate the
// [start, finish) of the download, and the result will be written into
// io.WriterAt.
//
// For example, if start=0 and finish=5, the function should return bytes 0-4.
type RangeService interface {
	DownloadRange(filepath string, w io.WriterAt, start, finish int) error
}

// StreamService allows for data to be streamed. Pass in the writer which should be streamed to.
type StreamService interface {
	Stream(filepath string, w io.Writer) error
}

// ListService allows for objects to be listed.
type ListService interface {
	// Objects should return an iterator for all objects in a bucket.
	Objects() (ObjectIterator, func())
}

// Object is metadata relating to an object in the data service.
type Object struct {
	// Filepath for the object.
	Filepath string

	// ContentType is the MIME type of the object's content.
	ContentType string

	// Size is the length of the object's content.
	Size int64
}

// ObjectIterator iterates through objects in a list.
type ObjectIterator interface {
	// Next returns the next result. Its second return value is iterator.Done if
	// there are no more results. Once Next returns iterator.Done, all subsequent
	// calls will return iterator.Done.
	Next() (*Object, error)
}
