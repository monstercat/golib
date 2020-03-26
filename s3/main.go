package s3util

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

// DEPRECATED: this file is deprecated. We should migrate to the newer S3 interface.

const (
	TempDirectoryEnv = "TEMP_DIR" // Would just use default temp directory if not specified
)

type S3Info interface {
	GetSession() *session.Session
	// The default operational bucket
	DefaultBucket() string
	// Allows prefixing, suffixing around keys based on hash sums
	MkHashKey(string) string
	// Creates an Service URL that utilizes the DefaultBucket
	MkURL(string) string
}

// Wrapper around Service Downloader for stubbing
type Downloader interface {
	Download(w io.WriterAt, i *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (int64, error)
}

// Wrapper around Service Uploader for stubbing
type Uploader interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

func NewDownloader(info S3Info) Downloader {
	return s3manager.NewDownloader(info.GetSession())
}

func NewConcurrentDownloader(info S3Info, count int) Downloader {
	x := s3manager.NewDownloader(info.GetSession())
	x.Concurrency = count
	return x
}

func NewOrderedDownloader(info S3Info) Downloader {
	x := s3manager.NewDownloader(info.GetSession())
	x.Concurrency = 1
	return x
}

func NewUploader(info S3Info) Uploader {
	return s3manager.NewUploader(info.GetSession())
}

type S3Baked struct {
	sess       *session.Session
	Bucket     string
	HashPrefix string
	Region     string
}

func (s *S3Baked) Init(id, secret string) error {
	var r string
	if v := os.Getenv("AWS_REGION"); v != "" {
		r = v
	} else {
		r = s.Region
	}
	var c *credentials.Credentials
	if v := os.Getenv("AWS_ACCESS_KEY_ID"); v != "" {
		c = credentials.NewEnvCredentials()
	} else {
		c = credentials.NewStaticCredentials(id, secret, "")
	}
	sess, err := session.NewSession(&aws.Config{
		Credentials: c,
		Region:      aws.String(r),
	})
	if err != nil {
		return err
	}
	s.sess = sess
	return nil
}

func (s *S3Baked) GetSession() *session.Session {
	return s.sess
}

func (s *S3Baked) DefaultBucket() string {
	if v := os.Getenv("S3_BUCKET"); v != "" {
		return v
	}
	return s.Bucket
}

func (s *S3Baked) MkHashKey(hash string) string {
	if v := os.Getenv("S3_KEY_PREFIX"); v != "" {
		return path.Join(v, hash)
	}
	return path.Join(s.HashPrefix, hash)
}

// MkURL creates a full s3 url from a file hashKey which is usually a prefix/hash combination
func (s *S3Baked) MkURL(key string) string {
	return "s3://" + path.Join(s.DefaultBucket(), key)
}

type SignedUrlConfig struct {
	Download bool
	Filename string
}

func (c SignedUrlConfig) GetDisposition() string {
	if !c.Download {
		return "inline"
	}
	if c.Filename == "" {
		return "attachment"
	}
	return fmt.Sprintf("attachment; filename=\"%s\"", c.Filename)
}

func (c SignedUrlConfig) GetObjectInput(bucket, key string) *s3.GetObjectInput {
	return &s3.GetObjectInput{
		Bucket:                     aws.String(bucket),
		Key:                        aws.String(key),
		ResponseContentDisposition: aws.String(c.GetDisposition()),
	}
}

func ExistsS3(info S3Info, key string) (bool, error) {
	_, exists, err := ObjectExistsS3(info, key)
	return exists, err
}

func ObjectExistsS3(info S3Info, key string) (*s3.HeadObjectOutput, bool, error) {
	sess := s3.New(info.GetSession())
	obj, err := sess.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(info.DefaultBucket()),
		Key:    aws.String(info.MkHashKey(key)),
	})

	if err == nil {
		return obj, true, nil
	}

	aerr, ok := err.(awserr.Error)
	if !ok {
		return nil, false, err
	}

	if aerr.Code() == "NotFound" {
		return nil, false, nil
	}
	return nil, false, err
}

func DeleteS3(info S3Info, key string) error {
	sess := s3.New(info.GetSession())
	_, err := sess.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(info.DefaultBucket()),
		Key:    aws.String(key),
	})
	return err
}

func SignedUrl(info S3Info, key string, duration time.Duration, cfg *SignedUrlConfig) (string, error) {
	sess := s3.New(info.GetSession())
	req, _ := sess.GetObjectRequest(cfg.GetObjectInput(
		info.DefaultBucket(),
		info.MkHashKey(key),
	))
	return req.Presign(duration)
}

func StreamS3(_s3 S3Info, hash string) (io.ReadCloser, error) {
	s3util := s3.New(_s3.GetSession())
	obj, err := s3util.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(_s3.DefaultBucket()),
		Key:    aws.String(_s3.MkHashKey(hash)),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Issue creating stream for key %s", hash)
	}
	return obj.Body, nil
}

func DownloadS3(downloader Downloader, bucket, key string) (*os.File, error) {
	file, err := ioutil.TempFile(os.Getenv(TempDirectoryEnv), "tagus-jobs-s3-file")
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create temp file while downloading s3 file %v:%v", bucket, key)
	}

	// Write the contents of Service Object to the file
	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to download s3 file %v:%v", bucket, key)
	}

	// Not checking the error probably doesn't matter if you can already write to the tmp file
	file.Seek(0, io.SeekStart)
	return file, nil
}

func DownloadRangeS3(dlr Downloader, w io.WriterAt, bucket, key string, start, finish int) error {
	_, err := dlr.Download(w, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Range:  aws.String(fmt.Sprintf("bytes=%d-%d", start, finish)),
	})
	return err
}

func UploadS3(
	info S3Info,
	r io.Reader,
	ct string,
	perms string,
	key string,
) error {

	sess := NewUploader(info)
	_, err := sess.Upload(&s3manager.UploadInput{
		ACL:         aws.String(perms),
		Body:        r,
		Bucket:      aws.String(info.DefaultBucket()),
		ContentType: aws.String(ct),
		Key:         aws.String(info.MkHashKey(key)),
	})
	return err
}

func splitS3(s3URL string) (bucket string, key string, err error) {
	// Todo handle virtual stlyle hosts?
	u, err := url.Parse(s3URL)
	if err != nil {
		return "", "", err
	}
	path := u.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	return u.Hostname(), path, nil
}
