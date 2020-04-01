package s3util

import (
	"bytes"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/monstercat/golib/data"
)

// This is a file upload!
type Upload struct {
	Filepath string
	FileSize int
	Uploaded int
	R        io.Reader

	// Expiry refers to the time when the upload will cease to exist from
	// the incomplete uploads list. Any expired partial file will be removed from
	// S3 as well as from memory.
	Expiry time.Time

	s        *Service
	notifier chan data.UploadStatus
	parts    chan []byte

	// Information required for S3 uploading
	uploadId       *string
	completedParts []*s3.CompletedPart

	// Whether or not it is currently in processing
	isProcessing bool
	lock         sync.RWMutex
}

func (u *Upload) GetFilepath() string {
	return u.Filepath
}
func (u *Upload) GetUploaded() int {
	return u.Uploaded
}

// This run function will load the parts from the channel and send them to S3, after initializing the transfer.
// This is called by the Service's eternally running goroutine after receiving the upload request from a channel.
func (u *Upload) Run() error {
	if u.uploadId == nil {
		if err := u.initS3Upload(); err != nil {
			// Errors are returned to the calling functions through the notifier channel.
			// returning error here allows for the upload to be sent back to incomplete status.
			//
			// In this case, it shouldn't be sent back to incomplete, as the upload never even started.
			u.notifier <- data.ErrUploadStatus(InitError{err})
			return nil
		}
	}

	for {
		select {
		case <-time.After(u.s.Timeout):
			return ErrTimeout
		case part := <-u.parts:
			//TODO: max retries
			if err := u.uploadPart(part); err != nil {
				return UploadError{err}
			}
			u.notifier <- data.UploadStatus{
				Code:    data.UploadStatusCodeProgress,
				Message: strconv.Itoa(u.Uploaded),
			}
			if len(u.completedParts) != cap(u.completedParts) {
				continue
			}
			if err := u.completeUpload(); err != nil {
				return UploadError{err}
			}
			u.notifier <- data.UploadStatus{
				Code:    data.UploadStatusCodeOk,
				Message: "Upload completed",
			}
			return nil
		}
	}
}

func (u *Upload) initS3Upload() error {
	resp, err := u.s.Client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(u.s.Bucket),
		Key:    aws.String(u.Filepath),
	})
	if err != nil {
		return err
	}
	u.uploadId = resp.UploadId
	return nil
}

func (u *Upload) uploadPart(part []byte) error {
	currPartNum := len(u.completedParts) + 1

	res, err := u.s.Client.UploadPart(&s3.UploadPartInput{
		Body:          bytes.NewReader(part),
		Bucket:        aws.String(u.s.Bucket),
		Key:           aws.String(u.Filepath),
		PartNumber:    aws.Int64(int64(currPartNum)),
		UploadId:      u.uploadId,
		ContentLength: aws.Int64(int64(len(part))),
	})
	if err != nil {
		return err
	}

	u.Uploaded += len(part)
	u.completedParts = append(u.completedParts, &s3.CompletedPart{
		PartNumber: aws.Int64(int64(currPartNum)),
		ETag:       res.ETag,
	})
	return nil
}

func (u *Upload) completeUpload() error {
	_, err := u.s.Client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(u.s.Bucket),
		Key:      aws.String(u.Filepath),
		UploadId: u.uploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: u.completedParts,
		},
	})
	return err
}

// This function sends to the parts channel from the reader.
func (u *Upload) Send() {
	for {
		b := make([]byte, u.s.ChunkSizeLimit * 2)
		n, err := u.R.Read(b)
		if err == io.EOF {
			// do nothing!
			return
		}
		if err != nil {
			u.notifier <- data.ErrUploadStatus(err)
			return
		}
		if n == 0 {
			continue
		}

		u.parts <- b[:n]
	}
}

func (u *Upload) setIsProcessing() {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.isProcessing = true
}

func (u *Upload) setDoneProcessing() {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.isProcessing = false
}

func (u *Upload) getIsProcessing() bool {
	u.lock.Lock()
	defer u.lock.Unlock()
	return u.isProcessing
}