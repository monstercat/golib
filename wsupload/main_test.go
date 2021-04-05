package wsupload

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/monstercat/websocket"
	"github.com/monstercat/websocket/wsjson"
)

var (
	service = &TestDataService{
		Uploads: make(map[string]*TestUpload),
	}
)

const (
	ChunkLimit = 32000
)

type TestUpload struct {
	Id     string
	Length int
	Data   []byte
}

type TestDataService struct {
	Uploads map[string]*TestUpload
}

func (s *TestDataService) NewUpload(id string, length int, r io.Reader) error {
	byt, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	u := &TestUpload{
		Id:     id,
		Length: length,
		Data:   byt,
	}
	s.Uploads[id] = u

	if len(byt) > length {
		return errors.New("data too long")
	}

	return nil
}

func (s *TestDataService) ResumeUpload(id string, r io.Reader) error {
	u, ok := s.Uploads[id]
	if !ok {
		return errors.New("cannot find upload")
	}

	byt, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	u.Data = append(u.Data, byt...)

	if len(byt) > u.Length {
		return errors.New("data too long")
	}

	return nil
}

func (s *TestDataService) GetUploadedBytes(id string) (int, error) {
	u, ok := s.Uploads[id]
	if !ok {
		return 0, errors.New("cannot find upload")
	}
	return len(u.Data), nil
}

func startWebsocketServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := &Handler{
			W:       w,
			R:       r,
			FS:      service,
			Timeout: 15 * time.Second,
		}
		if err := handler.Handle(); err != nil {
			log.Print(err)
		}
	}))
}

func createData(size int) []byte {
	byt := make([]byte, 0, size)
	for i := 0; i < size; i++ {
		byt = append(byt, byte(i))
	}
	return byt
}

func TestWebsocketUploading(t *testing.T) {
	ts := startWebsocketServer()
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, strings.Replace(ts.URL, "http", "ws", 1), nil)
	if err != nil {
		t.Fatal(err)
	}

	// We will send a single file to the server as a test.
	// Since we are sending a 1MB file, we will need to send it in multiple stints.
	// The protocol does not require a resume.
	contentLength := 1024 * 1024
	byt := createData(contentLength)
	id := "blob/file1"

	writeBin := func(byt []byte) error {
		w, err := conn.Writer(ctx, websocket.MessageBinary)
		if err != nil {
			return err
		}
		defer w.Close()
		_, err = io.Copy(w, bytes.NewReader(byt))
		return err
	}

	if err := wsjson.Write(ctx, conn, StartMessage{
		Message:       Message{Type: MsgTypeInit},
		Identifier:    id,
		ContentLength: contentLength,
	}); err != nil {
		t.Fatal(err)
	}
	var msg Message
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Type != MsgTypeOk {
		t.Fatal("Expecting return msg ok")
	}
	if err := writeBin(byt[0:ChunkLimit]); err != nil {
		t.Fatal(err)
	}
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Type != MsgTypeOk {
		t.Fatal("Expecting return msg ok")
	}
	for curr := ChunkLimit; curr < contentLength; curr+=ChunkLimit {
		end := curr+ChunkLimit
		if end > contentLength {
			end = contentLength
		}
		if err := writeBin(byt[curr : end]); err != nil {
			t.Fatal(err)
		}
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			t.Fatal(err)
		}
		if msg.Type != MsgTypeOk {
			t.Fatal("Expecting return msg ok")
		}
	}

	// Wait for the process to completely finish.
	time.Sleep(10 * time.Millisecond)

	// Check the upload file to make sure it's the same.
	up, ok := service.Uploads[id]
	if !ok {
		t.Fatal("Expected file to exist")
	}
	if len(up.Data) != contentLength {
		t.Fatal("Length not the same!")
	}

	for i, b := range up.Data {
		if b != byt[i] {
			t.Fatal("Data not the same!")
		}
	}
}

func TestWebsocketUploadingOkLater(t *testing.T) {
	ts := startWebsocketServer()
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, strings.Replace(ts.URL, "http", "ws", 1), nil)
	if err != nil {
		t.Fatal(err)
	}

	// We will send a single file to the server as a test.
	// Since we are sending a 1MB file, we will need to send it in multiple stints.
	// The protocol does not require a resume.
	contentLength := 1024 * 1024
	byt := createData(contentLength)
	id := "blob/file1"

	writeBin := func(byt []byte) error {
		w, err := conn.Writer(ctx, websocket.MessageBinary)
		if err != nil {
			return err
		}
		defer w.Close()
		_, err = io.Copy(w, bytes.NewReader(byt))
		return err
	}

	if err := wsjson.Write(ctx, conn, StartMessage{
		Message:       Message{Type: MsgTypeInit},
		Identifier:    id,
		ContentLength: contentLength,
	}); err != nil {
		t.Fatal(err)
	}
	var msg Message
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Type != MsgTypeOk {
		t.Fatal("Expecting return msg ok")
	}
	if err := writeBin(byt[0:ChunkLimit]); err != nil {
		t.Fatal(err)
	}
	for curr := ChunkLimit; curr < contentLength; curr+=ChunkLimit {
		end := curr+ChunkLimit
		if end > contentLength {
			end = contentLength
		}
		if err := writeBin(byt[curr : end]); err != nil {
			t.Fatal(err)
		}
	}

	// Read ok messages
	for curr := 0; curr < contentLength; curr+=ChunkLimit {
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			t.Fatal(err)
		}
		if msg.Type != MsgTypeOk {
			t.Fatal("Expecting return msg ok")
		}
	}

	// Wait for the process to completely finish.
	time.Sleep(10 * time.Millisecond)

	// Check the upload file to make sure it's the same.
	up, ok := service.Uploads[id]
	if !ok {
		t.Fatal("Expected file to exist")
	}
	if len(up.Data) != contentLength {
		t.Fatal("Length not the same!")
	}

	for i, b := range up.Data {
		if b != byt[i] {
			t.Fatal("Data not the same!")
		}
	}
}

func TestWebsocketUploadingWithResume(t *testing.T) {
	ts := startWebsocketServer()
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, strings.Replace(ts.URL, "http", "ws", 1), nil)
	if err != nil {
		t.Fatal(err)
	}

	// We will send a single file to the server as a test.
	// Since we are sending a 1MB file, we will need to send it in multiple stints.
	// Protocol doesn't require a resume, but we are using one anyways.
	contentLength := 1024 * 1024
	byt := createData(contentLength)
	id := "blob/file1"

	writeBin := func(byt []byte) error {
		w, err := conn.Writer(ctx, websocket.MessageBinary)
		if err != nil {
			return err
		}
		defer w.Close()
		_, err = io.Copy(w, bytes.NewReader(byt))
		return err
	}

	if err := wsjson.Write(ctx, conn, StartMessage{
		Message:       Message{Type: MsgTypeInit},
		Identifier:    id,
		ContentLength: contentLength,
	}); err != nil {
		t.Fatal(err)
	}
	var msg Message
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Type != MsgTypeOk {
		t.Fatal("Expecting return msg ok")
	}
	if err := writeBin(byt[0:ChunkLimit]); err != nil {
		t.Fatal(err)
	}
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Type != MsgTypeOk {
		t.Fatal("Expecting return msg ok")
	}
	for curr := ChunkLimit; curr < contentLength; curr+=ChunkLimit {
		if err := wsjson.Write(ctx, conn, StartMessage{
			Message:       Message{Type: MsgTypeResume},
			Identifier: id,
		}); err != nil {
			t.Fatal(err)
		}
		var statusMsg StatusMessage
		if err := wsjson.Read(ctx, conn, &statusMsg); err != nil {
			t.Fatal(err)
		}
		if statusMsg.Uploaded != curr {
			t.Fatal("Expected # of uploaded bytes do not match")
		}
		end := curr+ChunkLimit
		if end > contentLength {
			end = contentLength
		}
		if err := writeBin(byt[curr : end]); err != nil {
			t.Fatal(err)
		}
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			t.Fatal(err)
		}
		if msg.Type != MsgTypeOk {
			t.Fatal("Expecting return msg ok")
		}
	}

	// Wait for the process to completely finish.
	time.Sleep(10 * time.Millisecond)

	// Check the upload file to make sure it's the same.
	up, ok := service.Uploads[id]
	if !ok {
		t.Fatal("Expected file to exist")
	}
	if len(up.Data) != contentLength {
		t.Fatal("Length not the same!")
	}

	for i, b := range up.Data {
		if b != byt[i] {
			t.Fatal("Data not the same!")
		}
	}
}
