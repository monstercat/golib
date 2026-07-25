package s3util

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/monstercat/golib/data"
)

type sequentialWriterAt struct {
	w io.Writer
}

func (sw sequentialWriterAt) WriteAt(p []byte, offset int64) (int, error) {
	return sw.w.Write(p)
}

func (s *Service) download(filepath string, w io.WriterAt, concurrency int) error {
	dlr := s3manager.NewDownloader(s.Session)
	dlr.Concurrency = concurrency
	_, err := dlr.Download(w, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(filepath),
	})
	return err
}

func (s *Service) Stream(filepath string, w io.Writer) error {
	sw := &sequentialWriterAt{w: w}

	// Concurrency = 1 means it will be sequential
	return s.download(filepath, sw, 1)
}

// Get gets a file reader from Service
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

func (s *Service) Download(filepath string, w io.WriterAt, p *data.DownloadParams) error {
	concurrency := s.Concurrency
	if p != nil {
		if p.Concurrency != 0 {
			concurrency = p.Concurrency
		}
	}
	return s.download(filepath, w, concurrency)
}

// RangeReader allows for a part of the file to be read. Dictate the
// [start, finish) of the range. For example, if start=0 and finish=5, the
// returned reader provides bytes 0...4.
func (s *Service) RangeReader(filepath string, start, finish int) (io.ReadCloser, error) {
	// An empty range has nothing to request. S3 would reject the resulting
	// header outright.
	if finish <= start {
		return io.NopCloser(bytes.NewReader(nil)), nil
	}

	// The end of the HTTP range header is inclusive, whereas finish is
	// exclusive.
	out, err := s.Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(filepath),
		Range:  aws.String(fmt.Sprintf("bytes=%d-%d", start, finish-1)),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
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
