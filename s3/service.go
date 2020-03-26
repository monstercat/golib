package s3util

import (
	"fmt"
	"io"
	"math"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const DefaultChunkSizeLimit = 1024 * 1024 * 50
const DefaultIncompleteExpiry = time.Hour * 24

type Service struct {
	Bucket      string
	Region      string
	Timeout     time.Duration
	Concurrency int // Concurrent downloading

	ChunkSizeLimit         int
	MinUploadSizeChunked   int
	IncompleteUploadExpiry time.Duration

	incomplete map[string]*Upload
	lock       sync.RWMutex

	Session *session.Session
	Client  *s3.S3

	uploads  chan *Upload
	shutdown chan bool
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
	if s.MinUploadSizeChunked == 0 {
		s.MinUploadSizeChunked = DefaultChunkSizeLimit
	}
	if s.IncompleteUploadExpiry == 0 {
		s.IncompleteUploadExpiry = DefaultIncompleteExpiry
	}

	s.Client = s3.New(sess)
	s.Session = sess
	s.uploads = make(chan *Upload, 10)
	s.shutdown = make(chan bool)
	s.incomplete = make(map[string]*Upload)
	return nil
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

// Gets a file reader from Service
func (s *Service) Get(filepath string) (io.ReadCloser, error) {
	out, err := s.Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(filepath),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *Service) Download(filepath string, w io.WriterAt) error {
	dlr := s3manager.NewDownloader(s.Session)
	dlr.Concurrency = s.Concurrency
	_, err := dlr.Download(w, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(filepath),
	})
	return err
}

func (s *Service) DownloadRange(filepath string, w io.WriterAt, start, finish int) error {
	dlr := s3manager.NewDownloader(s.Session)
	dlr.Concurrency = s.Concurrency
	_, err := dlr.Download(w, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(filepath),
		Range:  aws.String(fmt.Sprintf("bytes=%d-%d", start, finish)),
	})
	return err
}

// Put here is generally only used if you do not care about progress or status updates as it hides
// this information from the user. You only get the final error or non-error code.
func (s *Service) Put(filepath string, filesize int, r io.Reader) error {
	ch := s.PutWithStatus(filepath, filesize, r)
	for {
		select {
		case <-time.After(s.Timeout):
			return ErrTimeout
		case status := <-ch:
			switch status.Code {
			case UploadStatusCodeOk:
				return nil
			case UploadStatusCodeError:
				return status.Error
			}
		}
	}
}

// Tries to put the file in Service. Use the returned channel to retrieve messages from the
// upload. With small uploads, the only statuses returned should be Ok and Error.
func (s *Service) PutWithStatus(filepath string, filesize int, r io.Reader) chan UploadStatus {
	if filesize <= s.MinUploadSizeChunked {
		return s.putSimply(filepath, r)
	}

	// Here, we need to put the file in a more complicated manner!
	// in this case, there might be more details to send back to the Client.
	chunks := int(math.Ceil(float64(filesize) / float64(s.ChunkSizeLimit)))

	upload := &Upload{
		Filepath:       filepath,
		FileSize:       filesize,
		R:              r,
		s:              s,
		Expiry:         time.Now().Add(s.IncompleteUploadExpiry),
		notifier:       make(chan UploadStatus),
		parts:          make(chan []byte, chunks),
		completedParts: make([]*s3.CompletedPart, 0, chunks),
	}
	s.uploads <- upload
	go upload.Send()

	return upload.notifier
}

// Resumes an incomplete upload.
func (s *Service) ResumePutWithStatus(filepath string, offset int, r io.Reader) (chan UploadStatus, error) {
	upload := s.GetIncompleteUpload(filepath)
	if upload == nil {
		return nil, ErrUploadNotFound
	}
	if upload.Uploaded != offset {
		return nil, ErrResumedUploadOffsetInvalid
	}

	// Take this upload and put it back on the channel, but with a new reader.
	s.removeFromIncomplete(filepath)
	upload.R = r
	upload.Expiry = time.Now().Add(s.IncompleteUploadExpiry)
	s.uploads <- upload
	go upload.Send()

	return upload.notifier, nil
}

// Deletes the file from S3
func (s *Service) Delete(filepath string) error {
	_, err := s.Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(filepath),
	})
	return err
}

// Retrieves the incomplete upload, it if exists.
func (s *Service) GetIncompleteUpload(filepath string) *Upload {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.incomplete[filepath]
}

func (s *Service) removeFromIncomplete(filepath string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.incomplete, filepath)
}

func (s *Service) registerIncompleteUpload(upload *Upload) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.incomplete[ upload.Filepath ] = upload
}

func (s *Service) putSimply(filepath string, r io.Reader) chan UploadStatus {
	upl := s3manager.NewUploader(s.Session)
	_, err := upl.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(filepath),
		Body:   r,
	})

	ch := make(chan UploadStatus, 1)
	if err != nil {
		ch <- errStatus(err)
	}
	ch <- okStatus()
	return ch
}

// This runner uploads larger files to S3. This should be run as its own service
// on the main function.
func (s *Service) RunUploader() {
	for {
		select {
		case <-s.shutdown:
			return
		case <- time.After(time.Hour):
			s.lock.Lock()
			for id, u := range s.incomplete {
				if u.Expiry.After(time.Now()) {
					delete(s.incomplete, id)
				}
			}
			s.lock.Unlock()
		case upload := <-s.uploads:
			go func() {
				if err := upload.Run(); err != nil {
					s.registerIncompleteUpload(upload)
					upload.notifier <- errStatus(err)
				}
			}()
		}
	}
}

func (s *Service) Shutdown() {
	// Only the main process needs to shutdown. The other goroutines should
	// shutdown on their own as they are fitted with exit on timeout.
	s.shutdown <- true
}
