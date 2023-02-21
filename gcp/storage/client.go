package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/monstercat/golib/data"
)

var (
	ErrTimeout = errors.New("timeout")
)

// Client wraps google's *storage.Client to implement:
// - data.Service
// - data.HeadService
// - data.ChunkService
// - data.SignedUrlService
//
// Note that it doesn't use data.ChunkManager because the default GCS client
// is already chunked. Additional chunking would only cause an extra useless
// goroutine to be run.
//
// Ensure that proper authentication is provided. Note that, if SignedUrl is
// to be used, that `google auth application-default login` is *not* used as
// the credential source as it is not adequate.
//
// Instead, use the GOOGLE_APPLICATION_CREDENTIALS environment variable or
// attach credentials (e.g., via passing in options.WithCredentialsJSON into
// the New function).
type Client struct {
	// The actual storage client for GCS
	*storage.Client

	// A quick way to access the bucket directly.
	Bucket *storage.BucketHandle

	// Timeout for each item.
	Timeout time.Duration

	// Mutex for thread safety.
	lock sync.RWMutex

	// List of 'incomplete' uploads.
	incomplete map[string]*ChunkUpload
}

func (c *Client) SetTimeout(tm time.Duration) {
	c.Timeout = tm
}

// NewClient creates a new GCP storage Client.
//
// DEPRECATED. Use New instead.
func NewClient(creds []byte, bucket string) (*Client, error) {
	return New(bucket, option.WithCredentialsJSON(creds))
}

// New generates a new client. option.ClientOption is not necessary unless
// special credentials are required in order to authenticate against GCS. For
// example, if the target machine is running on GCP or has the gcloud CLI setup
// with default authentication credentials, opts is not required.
//
// If nil is provided for context, context.Background is used by default.
func New(bucket string, opts ...option.ClientOption) (*Client, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	// Ensure that the bucket exists and can be read by the authenticated user.
	b := client.Bucket(bucket)
	_, err = b.Attrs(ctx)
	if err != nil {
		return nil, err
	}

	c := &Client{
		Bucket:     b,
		Client:     client,
		incomplete: make(map[string]*ChunkUpload),
	}
	return c, nil
}

// Close closes the Client connection
func (c *Client) Close() error {
	if c.Client == nil {
		return nil
	}
	return c.Client.Close()
}

func (c *Client) createContext() (context.Context, func()) {
	if c.Timeout > 0 {
		return context.WithTimeout(context.Background(), c.Timeout)
	}
	return context.Background(), func() {}
}

// Exists returns whether the filepath exists in the bucket.
func (c *Client) Exists(filepath string) (bool, error) {
	ctx, cancel := c.createContext()
	defer cancel()

	_, err := c.Bucket.Object(filepath).Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Head returns information regarding the requested file
func (c *Client) Head(filepath string) (*data.HeadInfo, error) {
	ctx, cancel := c.createContext()
	defer cancel()

	attrs, err := c.Bucket.Object(filepath).Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return &data.HeadInfo{
			Exists: false,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return &data.HeadInfo{
		Exists:        true,
		LastModified:  attrs.Updated,
		ContentLength: attrs.Size,
	}, nil
}

// Get returns a reader that can be used to retrieve a file.
func (c *Client) Get(filepath string) (io.ReadCloser, error) {
	ctx, cancel := c.createContext()
	return cancelWrapReadCloser(cancel)(c.Bucket.Object(filepath).NewReader(ctx))
}

func cancelWrapReadCloser(cancel func()) func(closer io.ReadCloser, err error) (io.ReadCloser, error) {
	return func(r io.ReadCloser, err error) (io.ReadCloser, error) {
		if err != nil {
			return nil, err
		}
		return &contextCancelReadCloser{
			cancel:     cancel,
			ReadCloser: r,
		}, nil
	}
}

// contextCancelReadCloser is a special read closer that calls cancel when
// close is cancelled. This is to allow timeout to work for the storage client.
type contextCancelReadCloser struct {
	cancel func()
	io.ReadCloser
}

// Close implements the close interface.
func (c *contextCancelReadCloser) Close() error {
	c.cancel()
	return c.ReadCloser.Close()
}

func (c *Client) Delete(filepath string) error {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.Bucket.Object(filepath).Delete(ctx)
}

func (c *Client) Put(filepath string, r io.Reader) error {
	ctx, cancel := c.createContext()
	defer cancel()

	w := c.Bucket.Object(filepath).NewWriter(ctx)
	defer w.Close()

	_, err := io.Copy(w, r)
	return err
}

// SignedUrl is for the Client to implement the SignedUrlService.
// SignedUrlService allows a service to return a signed URL for a filepath.
//
// For the implementation below, the GoogleAccessID is required. However, there
// is a chance that the environment does *not* provide this GoogleAccessID.
// For example, a system using `google auth application-default login`
// while looking similar, is *not* a valid credential to use as it does
// not contain the required `client_email` parameter.
//
// Ensure (in testing and on the server) that proper credentials are provided
// when creating the client.
// https://cloud.google.com/docs/authentication/application-default-credentials
func (c *Client) SignedUrl(filepath string, tm time.Duration, cfg *data.SignedUrlConfig) (string, error) {
	// As of https://github.com/googleapis/google-cloud-go/pull/4604 we do not
	// need to send credentials. Existing credentials (e.g., as eet up via
	// option.ClientOption can be used.
	//
	// It checks for existence of GoogleAccessID through the metadata service,
	// and signs the bytes properly unless a credentials JSON is provided, in
	// which case it uses the private key from the credentials JSON.
	//
	// In the first case, an HTTP call is used to authenticate. However,
	// for the credentials JSON, the filepath is signed automatically.
	str, err := c.Bucket.SignedURL(filepath, &storage.SignedURLOptions{
		// TODO: test SigningSchemeV4.
		//Scheme: storage.SigningSchemeV4,
		Method:  http.MethodGet,
		Expires: time.Now().Add(tm),
	})
	if err != nil {
		return "", err
	}

	// The signed URL does not contain any of the required content type and
	// disposition parameters to make the URL usable. We can add it directly
	// to the URL through additional query string parameters.
	//
	// https://cloud.google.com/storage/docs/access-control/signed-urls-v2
	// https://cloud.google.com/storage/docs/access-control/signed-urls
	// In both V2 and V4, the signed URLs are an XML-API endpoint and can
	// therefore use the query parameters specified for XML-API, below.
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	qry := u.Query()

	// https://cloud.google.com/storage/docs/xml-api/reference-headers#responsecontentdisposition
	qry.Set("response-content-disposition", cfg.GetDisposition())

	// https://cloud.google.com/storage/docs/xml-api/reference-headers#responsecontenttype
	qry.Set("response-content-type", cfg.GetContentType())
	u.RawQuery = qry.Encode()

	return u.String(), nil
}

// writeAtWrap wraps an io.WriterAt so that it functions as a Write. It does
// this by keeping an index of the current write location, which starts at 0.
type writeAtWrap struct {
	io.WriterAt
	curr int64
	mu   sync.RWMutex
}

func (w *writeAtWrap) Write(p []byte) (int, error) {
	// Each write needs to be sequential
	w.mu.Lock()
	defer w.mu.Unlock()

	// Write it and increment
	n, err := w.WriteAt(p, w.curr)
	if err != nil {
		return 0, err
	}
	w.curr += int64(n)
	return n, nil
}

// Download is supposed to download a file in parallel (by splitting it into
// different parts to download). However, GCS doesn't seem to have this
// functionality built in. Thus, we ignore it.
func (c *Client) Download(filepath string, w io.WriterAt, p *data.DownloadParams) error {
	rdr, err := c.Get(filepath)
	if err != nil {
		return err
	}
	defer rdr.Close()

	wr := &writeAtWrap{WriterAt: w}
	if _, err := io.Copy(wr, rdr); err != nil {
		return err
	}
	return nil
}

// Upload sends each reader to the chunk uploader for automatic chunked upload.
func (c *Client) Upload(filepath string, filesize int, r io.Reader) error {
	ch := c.PutWithStatus(filepath, filesize, r)
	for {
		select {
		case <-time.After(c.Timeout):
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

// ChunkUpload satisfies the data.Upload interface for GCS chunked uploads.
type ChunkUpload struct {
	writer *storage.Writer

	// Number of total upload bytes. This is to provide a progress message giving them number of uploaded bytes.
	UploadedBytes int

	// Notifier to pass information back to the user.
	Notifier chan data.UploadStatus

	// Total file size
	FileSize int

	// Lock for thread safety
	lock sync.RWMutex
}

// GetFilepath returns the filepath. Required to implement Upload interface.
func (u *ChunkUpload) GetFilepath() string {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.writer.Name
}

// GetUploaded returns the number of bytes already uplaoded. Required to implement Upload interface.
func (u *ChunkUpload) GetUploaded() int {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.UploadedBytes
}

func (u *ChunkUpload) UpdateUploadedBytes(newPartBytes int) {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.UploadedBytes += newPartBytes
}

func (u *ChunkUpload) IsComplete() bool {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.UploadedBytes >= u.FileSize
}

// Copy writes the contents fo the reader into the chunk upload. It also
// handles all error and success notifications.
func (u *ChunkUpload) Copy(r io.Reader) (bool, error) {
	n, err := io.Copy(u.writer, r)
	if err != nil {
		u.Notifier <- data.UploadStatus{
			Code:  data.UploadStatusCodeError,
			Error: err,
		}
		return false, err
	}
	u.UpdateUploadedBytes(int(n))

	if !u.IsComplete() {
		return false, nil
	}

	if err := u.writer.Close(); err != nil {
		u.Notifier <- data.UploadStatus{
			Code:  data.UploadStatusCodeError,
			Error: err,
		}
		return false, err
	}

	u.Notifier <- data.UploadStatus{
		Code:    data.UploadStatusCodeOk,
		Message: "Upload completed",
	}

	return true, nil
}

// PutWithStatus tells GCS to start the upload. It adds the upload to the
// upload registry and proceeds to send data from the provided io.Reader.
func (c *Client) PutWithStatus(filepath string, filesize int, r io.Reader) chan data.UploadStatus {
	u := &ChunkUpload{
		writer:   c.Bucket.Object(filepath).NewWriter(context.Background()),
		Notifier: make(chan data.UploadStatus, 3),
		FileSize: filesize,
	}
	// Write to the notifier based on progress.
	u.writer.ProgressFunc = func(i int64) {
		go func() {
			u.Notifier <- data.UploadStatus{
				Code:    data.UploadStatusCodeProgress,
				Message: fmt.Sprintf("%d", i),
			}
		}()
	}

	// We need to return notifier *right away*. Progress reporting might cause
	// it to get stuck! We ignore the error because it is already sent on the
	// notifier. 
	go func() {
		finished, _ := u.Copy(r)
		if !finished {
			c.lock.Lock()
			c.incomplete[filepath] = u
			c.lock.Unlock()
		}
	}()
	return u.Notifier
}

// ResumePutWithStatus tells CSV to add more data to the upload. If the upload
// cannot be found on the upload registry, it returns data.ErrUploadNotFound.
func (c *Client) ResumePutWithStatus(filepath string, r io.Reader) (chan data.UploadStatus, error) {
	upload := c.getIncompleteUpload(filepath)
	if upload == nil {
		return nil, data.ErrUploadNotFound
	}
	finished, err := upload.Copy(r)
	if err != nil {
		return upload.Notifier, err
	}
	if finished {
		c.lock.Lock()
		delete(c.incomplete, filepath)
		c.lock.Unlock()
	}

	return upload.Notifier, nil
}

func (c *Client) getIncompleteUpload(filepath string) *ChunkUpload {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.incomplete[filepath]
}

// GetIncompleteUpload returns an incomplete upload based on the filepath.
func (c *Client) GetIncompleteUpload(filepath string) data.Upload {
	return c.getIncompleteUpload(filepath)
}

// GetChunkSize returns the ChunkSize so that operators know how much data to
// provide.
func (c *Client) GetChunkSize() int {
	return googleapi.DefaultUploadChunkSize
}

// Stream allows data to be streamed into a writer.
func (c *Client) Stream(filepath string, w io.Writer) error {
	ctx, cancel := c.createContext()
	defer cancel()

	r, err := c.Bucket.Object(filepath).NewReader(ctx)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, r); err != nil {
		return err
	}
	return nil
}

// DownloadRange allows for a part of the file to be downloaded. Dictate the
// [start, finish) of the download, and the result will be written into
// io.WriterAt. For example, if start=0 and end=5, bytes 0...4 should be
// writen.
func (c *Client) DownloadRange(filepath string, w io.WriterAt, start, finish int) error {
	ww := &writeAtWrap{WriterAt: w}

	ctx, cancel := c.createContext()
	defer cancel()

	offset := int64(start)
	length := int64(finish) - offset
	r, err := c.Bucket.Object(filepath).NewRangeReader(ctx, offset, length)
	if err != nil {
		return err
	}
	if _, err := io.Copy(ww, r); err != nil {
		return err
	}
	return nil
}

// Objects should return an iterator for all objects in a bucket.
func (c *Client) Objects() (data.ObjectIterator, func()) {
	ctx, cancel := c.createContext()
	it := c.Bucket.Objects(ctx, nil)
	return &gcsObjectIterator{ObjectIterator: it}, cancel
}

type gcsObjectIterator struct {
	*storage.ObjectIterator
}

func (i *gcsObjectIterator) Next() (*data.Object, error) {
	obj, err := i.ObjectIterator.Next()
	if err != nil {
		return nil, err
	}
	return &data.Object{
		Filepath:    obj.Name,
		ContentType: obj.ContentType,
		Size:        obj.Size,
	}, nil
}
