package storage

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/monstercat/golib/data"
)

// Client wraps google's *storage.Client and implements the data.Service interface
type Client struct {
	creds      *credentialsFile
	BucketName string

	Client *storage.Client
	Bucket *storage.BucketHandle
}

// NewClient creates a new GCP storage Client.
func NewClient(creds []byte, bucket string) (*Client, error) {
	c := &Client{
		BucketName: bucket,
	}
	if err := c.decodeCreds(creds); err != nil {
		return nil, err
	}

	client, err := storage.NewClient(context.Background(), option.WithCredentialsJSON(creds))
	if err != nil {
		return nil, err
	}

	c.Client = client
	c.Bucket = client.Bucket(bucket)

	// Check if the bucket exists. If not, return error.
	_, err = c.Bucket.Attrs(context.Background())
	if err != nil {
		return nil, err
	}

	return c, nil
}

// credentialsFile is the unmarshalled representation of a credentials file.
// This is taken from golang.org/x/oauth2/google/default.go
type credentialsFile struct {
	Type string `json:"type"` // serviceAccountKey or userCredentialsKey

	// Service Account fields
	ClientEmail  string `json:"client_email"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	TokenURL     string `json:"token_uri"`
	ProjectID    string `json:"project_id"`

	// User Credential fields
	// (These typically come from gcloud auth.)
	ClientSecret string `json:"client_secret"`
	ClientID     string `json:"client_id"`
	RefreshToken string `json:"refresh_token"`
}

func (c *Client) decodeCreds(creds []byte) error {
	var x credentialsFile
	if err := json.Unmarshal(creds, &x); err != nil {
		return err
	}
	c.creds = &x
	return nil
}

// Close closes the Client connection
func (c *Client) Close() error {
	if c.Client == nil {
		return nil
	}
	return c.Client.Close()
}

// Exists returns whether the filepath exists in the bucket.
func (c *Client) Exists(filepath string) (bool, error) {
	_, err := c.Bucket.Object(filepath).Attrs(context.Background())
	return err != storage.ErrObjectNotExist, err
}

// Head returns information regarding the requested file
func (c *Client) Head(filepath string) (*data.HeadInfo, error) {
	attrs, err := c.Bucket.Object(filepath).Attrs(context.Background())
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
	return c.Bucket.Object(filepath).NewReader(context.Background())
}

func (c *Client) Delete(filepath string) error {
	return c.Bucket.Object(filepath).Delete(context.Background())
}

func (c *Client) Put(filepath string, r io.Reader) error {
	w := c.Bucket.Object(filepath).NewWriter(context.Background())
	defer w.Close()

	_, err := io.Copy(w, r)
	return err
}

func (c *Client) SignedUrl(filepath string, tm time.Duration, cfg *data.SignedUrlConfig) (string, error) {
	str, err := storage.SignedURL(c.BucketName, filepath, &storage.SignedURLOptions{
		GoogleAccessID: c.creds.ClientEmail,
		PrivateKey:     []byte(c.creds.PrivateKey),
		Method:         http.MethodGet,
		Expires:        time.Now().Add(tm),
	})
	if err != nil {
		return "", err
	}

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
