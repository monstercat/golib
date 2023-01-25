package s3util

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/monstercat/golib/data"
	strutil "github.com/monstercat/golib/string"
)

const DefaultChunkSizeLimit = 1024 * 1024 * 50
const DefaultIncompleteExpiry = time.Hour * 24

// Adheres to the data.Service interface
type Service struct {
	data.ChunkManager

	Bucket      string
	Region      string
	Timeout     time.Duration
	Concurrency int // Concurrent downloading

	//incomplete map[string]*Upload
	lock sync.RWMutex

	Session *session.Session
	Client  *s3.S3
}

func (s *Service) SetTimeout(tm time.Duration) {
	s.Timeout = tm
}

func (s *Service) SetConcurrency(concurrency int) {
	s.Concurrency = concurrency
}

func (s *Service) Init(id, secret string) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(id, secret, ""),
		Region:      aws.String(s.Region),
	})
	if err != nil {
		return err
	}

	if s.ChunkSizeLimit == 0 {
		s.ChunkSizeLimit = DefaultChunkSizeLimit
	}
	if s.IncompleteUploadExpiry == 0 {
		s.IncompleteUploadExpiry = DefaultIncompleteExpiry
	}

	s.Client = s3.New(sess)
	s.Session = sess
	return nil
}

func (s *Service) Head(filepath string) (*data.HeadInfo, error) {
	obj, err := s.Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(filepath),
	})
	if err != nil {
		aerr, ok := err.(awserr.Error)
		if !ok {
			return nil, err
		}
		if aerr.Code() == "NotFound" {
			return &data.HeadInfo{
				Exists: false,
			}, nil
		}
	}

	info := &data.HeadInfo{
		Exists: true,
	}
	if obj.LastModified != nil {
		info.LastModified = *obj.LastModified
	}
	if obj.ContentLength != nil {
		info.ContentLength = *obj.ContentLength
	}
	return info, nil
}

func (s *Service) SignedUrl(filename string, tm time.Duration, config *data.SignedUrlConfig) (string, error) {
	// Before signing, the filename needs to be in ISO 8859 1 format.
	var err error
	config.Filename, err = strutil.ToISO_8859_1(config.Filename)
	if err != nil {
		return "", err
	}

	req, _ := s.Client.GetObjectRequest(GetObjectInputFromConfig(
		config,
		s.Bucket,
		filename,
	))
	return req.Presign(tm)
}

// Checks if the file exists
func (s *Service) Exists(filepath string) (bool, error) {
	_, err := s.Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(filepath),
	})
	if err == nil {
		return true, nil
	}

	aerr, ok := err.(awserr.Error)
	if !ok || aerr.Code() != "NotFound" {
		return false, err
	}
	return false, nil
}

func (s *Service) Put(filepath string, r io.Reader) error {
	ch := s.putSimply(filepath, r)
	select {
	case <-time.After(s.Timeout):
		return ErrTimeout
	case status := <-ch:
		switch status.Code {
		case data.UploadStatusCodeError:
			return status.Error
		default:
			return nil
		}
	}
}

// Upload here is generally only used if you do not care about progress or status updates as it hides
// this information from the user. You only get the final error or non-error code.
func (s *Service) Upload(filepath string, filesize int, r io.Reader) error {
	ch := s.PutWithStatus(filepath, filesize, r)
	for {
		select {
		case <-time.After(s.Timeout):
			return ErrTimeout
		case status := <-ch:
			switch status.Code {
			case data.UploadStatusCodeOk:
				return nil
			case data.UploadStatusCodeError:
				return status.Error
			}
		}
	}
}

// Delete the file from FS
func (s *Service) Delete(filepath string) error {
	_, err := s.Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(filepath),
	})
	return err
}

func (s *Service) putSimply(filepath string, r io.Reader) chan data.UploadStatus {
	upl := s3manager.NewUploader(s.Session)
	_, err := upl.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(filepath),
		Body:   r,
	})

	ch := make(chan data.UploadStatus, 1)
	if err != nil {
		ch <- data.ErrUploadStatus(err)
	} else {
		ch <- data.OkUploadStatus()
	}
	return ch
}

// ChunkFileService is a data.ChunkFileService specific for FS.
type ChunkFileService struct {
	Bucket string
	Client *s3.S3

	// Filepath stored from initialization
	Filepath string

	// ID provided by AWS
	uploadId *string

	// Completed Parts to be sent back when operation is finished.
	completedParts []*s3.CompletedPart
}

// IsInitialized should return true if the ChunkFileService has already been initialized.
func (s *ChunkFileService) IsInitialized() bool {
	return s.uploadId != nil
}

// InitializeChunk initializes the chunk upload operation.
func (s *ChunkFileService) InitializeChunk(u *data.ChunkUpload) error {
	resp, err := s.Client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(u.Filepath),
	})
	if err != nil {
		return err
	}

	s.uploadId = resp.UploadId
	s.Filepath = u.Filepath
	return nil
}

// UploadPart should upload the provided bytes to the cloud. Uploaded parts should be completely sequential.
func (s *ChunkFileService) UploadPart(part []byte) error {
	currPartNum := len(s.completedParts) + 1

	res, err := s.Client.UploadPart(&s3.UploadPartInput{
		Body:          bytes.NewReader(part),
		Bucket:        aws.String(s.Bucket),
		Key:           aws.String(s.Filepath),
		PartNumber:    aws.Int64(int64(currPartNum)),
		UploadId:      s.uploadId,
		ContentLength: aws.Int64(int64(len(part))),
	})
	if err != nil {
		return err
	}

	s.completedParts = append(s.completedParts, &s3.CompletedPart{
		PartNumber: aws.Int64(int64(currPartNum)),
		ETag:       res.ETag,
	})
	return nil
}

// CompleteUpload should perform any necessary steps to finalize the multipart upload.
func (s *ChunkFileService) CompleteUpload() error {
	_, err := s.Client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s.Bucket),
		Key:      aws.String(s.Filepath),
		UploadId: s.uploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: s.completedParts,
		},
	})
	return err
}

// Cleanup should delete the ChunkFileService and do any other cleanup necessary. It is only called when an
// incomplete upload has timed out.
func (s *ChunkFileService) Cleanup() {
	// do nothing for now.
}
