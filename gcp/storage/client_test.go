package storage

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/fsouza/fake-gcs-server/fakestorage"

	"github.com/monstercat/golib/data"
)

const (
	fakePrivateKey = `
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQCqGw3pOPKDx6WYg4BGxu2g3E404x7lP3U2gOWRfXs5ELOV16+O
KLOKUiMVLrnPP0uDj0w8IDgfDOrWdTZGExyB26yM+QNoRxMoVMeGijQyXVGZifBP
BEKq4f8fMF5vrwrjQj59T4ryoM8uPtBq1cwDDfrcl0DlVO4y5thP0gmaIQIDAQAB
AoGBAKXXVoqogJfFz0aP/kICs63+2yhovbhXU9lddXOQ2M/b3poZ/Agm2lPinF2M
fo71cJPE41hDOTPcjh+jitRq0YCVpfzUPVqeYibAhQN6bAh3z6ElvHSPKvexXbNL
OKVTFInaLrXBa26z0I4ibRh5C7WV7Wig1UnAgGMj1FIF72YhAkEA8um0HYPH8aO7
3HBOlR6/J0uTS+L3D9D5wPWnExMvrt5Fqf4KSG7UJ6aQJyBKw4jlryV3I1lcmYZu
Ii6apLjx1QJBALNFLlmgb+02KASdLKkxKhUg7+dn/Q/TzY8Nusf6FNWThrztHpuA
u7JaNDF2NbLW4ovlb8foa9ZIA3J5E2Vi4R0CQAvLSRF9yoFy/7YOReJ7obBYvQgc
Nv6vmNDDnJ8SeWg2Jo/AY+NsbiSWs70Slk60IOLGIOi4eASEQGisdpm02RkCQA13
ZervpVjBV7I5CFDRU6LwrXTJl/XnaCqV0nERNR1yDo4ElecCfZcBNah9g70ibTQr
EQGIUQlwsWmY9L8J9XUCQQDVm21njykzYhoflqyfrdDSpx1eu9n3c9g6SJU7qo7+
BBg4NK/4TTs3xZSb+YqGvP7W324PxJkmKlc1cGOTOOGu
-----END RSA PRIVATE KEY-----
`
	fakePublicKey = `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCqGw3pOPKDx6WYg4BGxu2g3E40
4x7lP3U2gOWRfXs5ELOV16+OKLOKUiMVLrnPP0uDj0w8IDgfDOrWdTZGExyB26yM
+QNoRxMoVMeGijQyXVGZifBPBEKq4f8fMF5vrwrjQj59T4ryoM8uPtBq1cwDDfrc
l0DlVO4y5thP0gmaIQIDAQAB
-----END PUBLIC KEY-----
`
)

func createCreds() []byte {
	creds := &credentialsFile{
		ClientEmail: "test@test.com",
		PrivateKey:  fakePrivateKey,
	}

	b, _ := json.Marshal(creds)
	return b
}

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
		Host: "127.0.0.1",
		Port: 1337,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()

	client := &Client{
		BucketName: "Test-Bucket",
		Client:     server.Client(),
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
}

func TestClient_SignedUrl(t *testing.T) {
	creds := createCreds()
	client := &Client{
		BucketName: "bucket",
	}

	if err := client.decodeCreds(creds); err != nil {
		t.Fatal(err)
	}
	str, err := client.SignedUrl("test-file-path", time.Hour, &data.SignedUrlConfig{
		ContentType: "application/json",
		Download:    true,
		Filename:    "test-file-name",
	})
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse(str)
	if err != nil {
		t.Fatal(err)
	}

	qry := u.Query()
	ct := qry.Get("response-content-type")
	cd := qry.Get("response-content-disposition")

	if ct != "application/json" {
		t.Errorf("Content Type should be provided through response-content-type. Got %s", ct)
	}
	if cd != "attachment; filename=\"test-file-name\"" {
		t.Errorf("Content disposition invalid. Got %s", cd)
	}

	str, err = client.SignedUrl("test-file-path", time.Hour, &data.SignedUrlConfig{
		ContentType: "application/json",
		Download:    false,
	})
	if err != nil {
		t.Fatal(err)
	}

	u, err = url.Parse(str)
	if err != nil {
		t.Fatal(err)
	}

	qry = u.Query()
	ct = qry.Get("response-content-type")
	cd = qry.Get("response-content-disposition")

	if ct != "application/json" {
		t.Errorf("Content Type should be provided through response-content-type. Got %s", ct)
	}
	if cd != "inline" {
		t.Errorf("Content disposition invalid. Got %s", cd)
	}
}
