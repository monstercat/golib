package storage

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/monstercat/golib/data"
)

// TestClient tests GCS using the fakestorage client for GCS.
// - Client.Exists
// - Client.Delete
// - Client.Head
// - Client.Get
// - Client.Put
//
// It also ensures that SignedUrl works without error.
func TestClient(t *testing.T) {
	content := []byte("test file content")
	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		InitialObjects: []fakestorage.Object{
			{
				ObjectAttrs: fakestorage.ObjectAttrs{
					BucketName: "Test-Bucket",
					Name:       "test-file.txt",
				},
				Content: content,
			},
		},
		NoListener: true,
		Host:       "127.0.0.1",
		Port:       1337,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()

	// this is required in order to retrieve proper credentials. Otherwise,
	// server.Client() would be sufficient.
	gcsClient, err := storage.NewClient(
		context.Background(),
		option.WithHTTPClient(server.HTTPClient()),
	)
	client := &Client{
		Client: gcsClient,
	}
	client.Bucket = client.Client.Bucket("Test-Bucket")

	exists, err := client.Exists("test-file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("test-file.txt should exist")
	}

	// Test a file that does not exist.
	head, err := client.Head("test-file.txts")
	if err != nil {
		t.Fatal(err)
	}
	if head == nil {
		t.Fatal("Head should not be nil")
	}
	if head.Exists {
		t.Fatal("test-file.txts should not exist")
	}

	head, err = client.Head("test-file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if head == nil {
		t.Fatal("Head should not be nil")
	}
	if !head.Exists {
		t.Fatal("test-file.txt should exist")
	}
	if int(head.ContentLength) != len(content) {
		t.Fatal("Content length invalid")
	}

	r, err := client.Get("test-file.txt")
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != string(content) {
		t.Fatalf("Content is not the same. Got %s", b)
	}

	if err := client.Delete("test-file.txts"); err != storage.ErrObjectNotExist {
		t.Fatalf("Expecting an 'object doesn't exist' error. Got %s", err)
	}
	if err := client.Delete("test-file.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Get("test-file.txt"); err != storage.ErrObjectNotExist {
		t.Fatalf("Expect object doesn't exists error. Got %s", err)
	}

	if err := client.Put("test-file.txt", bytes.NewReader(content)); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Get("test-file.txt"); err != nil {
		t.Fatal(err)
	}

	s, err := client.SignedUrl("test-file.txt", time.Hour, &data.SignedUrlConfig{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
}

// TestChunkedUpload tests the chunked upload functionality. It ensures that Put
// Resume work properly.
func TestChunkedUpload(t *testing.T) {
	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		Host:       "127.0.0.1",
		Port:       1337,
		NoListener: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()

	client := &Client{
		Client:     server.Client(),
		incomplete: make(map[string]*ChunkUpload),
	}
	client.Bucket = client.Client.Bucket("Test-Bucket")

	// Create the bucket
	if err := client.Bucket.Create(context.Background(), "", nil); err != nil {
		t.Fatal(err)
	}

	// Constants for this test
	const filepath = "test_file"
	const chunkSize = googleapi.DefaultUploadChunkSize
	const numChunks = 5
	const filesize = chunkSize * numChunks

	// Contents.
	contents := make([]rune, 0, filesize)
	for i := 0; i < filesize; i++ {
		contents = append(contents, rune((i%26)+65))
	}
	getChunk := func(i int) io.Reader {
		if i >= numChunks {
			return nil
		}
		start := i * chunkSize
		finish := start + chunkSize
		return strings.NewReader(string(contents[start:finish]))
	}

	// 1. Put a file with status.
	notifier := client.PutWithStatus(filepath, filesize, getChunk(0))
	go func() {
		// We are just going to ignore the notifier
		for n := range notifier {
			t.Logf("%#v", n)
		}
	}()
	defer close(notifier)

	for i := 1; i < filesize/chunkSize; i++ {
		_, err := client.ResumePutWithStatus(filepath, getChunk(i))
		if err != nil {
			t.Fatal(err)
		}
	}

	if client.GetIncompleteUpload(filepath) != nil {
		t.Errorf("Expecting upload to be complete. But we were still able to retrieve an incomplete upload.")
	}
}
