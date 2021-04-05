package wsupload

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/monstercat/websocket"
)

var (
	ErrClosedByClient = errors.New("closed by client")
)

// This package implements a websocket handler with a protocol to handle (sequential) resumable data uploads. It
//contains the following messages:
// - INIT - initialize a new file upload with a identifier and a contentLength.
// - RESUME - resume a file upload. Request should come with a identifier. It will return a # indicating the number
//    of uploaded bytes. Upload should proceed with missing bytes.
// See Message for message signature (JSON).
//
// Other message can be sent using the handler's Send function.

// DataServiceInterface is an interface the websocket handler is prepared to receive.
type DataServiceInterface interface {
	NewUpload(identifier string, contentLength int, r io.Reader) error
	ResumeUpload(identifier string, r io.Reader) error
	GetUploadedBytes(identifier string) (int, error)
}

// Websocket handler.
type Handler struct {
	W       http.ResponseWriter
	R       *http.Request
	FS      DataServiceInterface
	Timeout time.Duration
	ErrCH   chan error

	// Delay between reading of successive messages.
	// This gives time for the server to handle messages
	// that are already in the system. This will typically be
	// something like 10ms.
	MessageDelay time.Duration

	// Id of the currently being uploaded data and
	// whether or not the upload itself has started.
	startMsg   *StartMessage
	hasStarted bool

	conn  *websocket.Conn
	close chan bool
}

// Creates a context with timeout, if a timeout is provided as
// part of initializing the handler.
func (h *Handler) ctx() (context.Context, context.CancelFunc) {
	if h.Timeout == 0 {
		return h.R.Context(), func() {}
	}
	return context.WithTimeout(h.R.Context(), h.Timeout)
}

// Handle will implement the upload protocol. If an uploadId is provided
// it will assume that the upload is a continuing upload. Otherwise, it
// is a new upload.
//
// Calling functions should assume that errors are already handled and
// that the connection will be closed on error. HOWEVER, in the case
// of success, the connection will *not* be closed. This way,
// the calling function can continue to send messages through this
// same connection.
//
// Input error channel is to handle errors from the file service.
// The calling function can continue to send messages
func (h *Handler) Handle() error {
	// Initialize parameters
	h.close = make(chan bool)

	// Start accepting the websocket.
	wsc, err := websocket.Accept(h.W, h.R, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	h.conn = wsc

	for {
		select {
		case <-h.close:
			return err
		case <-time.After(h.MessageDelay):
		}

		err := h.decodeNextMessage()
		switch {
		case err == ErrClosedByClient:
			// closed by client. Nothing to handle anymore.
			return err
		case isDataError(err) || err == ErrInvalidMessageType:
			h.Send(NewErrorMessage(err))
		case err != nil:
			h.ErrCH <- err
		}
	}

}

func isDataError(err error) bool {
	_, ok := err.(DataError)
	return ok
}

func (h *Handler) decodeNextMessage() error {
	ctx, cancel := h.ctx()
	defer cancel()

	t, r, err := h.conn.Reader(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "WebSocket closed") {
			return ErrClosedByClient
		}
		return DataError{err}
	}

	switch t {
	// If binary, then we have some type of data to upload
	case websocket.MessageBinary:
		// If no message was previously received first, we have no context
		// of what the upload should be.
		if h.startMsg == nil {
			return ErrInvalidMessageType
		}
		// We have received the message properly. Let's send back an OK
		// while it is uploading.
		if err := h.Send(OkMessage); err != nil {
			return err
		}
		if h.hasStarted {
			return h.FS.ResumeUpload(h.startMsg.Identifier, r)
		}
		h.hasStarted = true
		return h.FS.NewUpload(h.startMsg.Identifier, h.startMsg.ContentLength, r)
	// If it is text, it should be a StartMessage, which is the
	// only message that the user can send. In this case, we can only
	// parse this message and return the error.
	case websocket.MessageText:
		var msg StartMessage
		byt, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(byt, &msg); err != nil {
			return err
		}
		return h.handleStartMessage(&msg)
	}

	h.startMsg = nil
	return ErrInvalidMessageType
}

func (h *Handler) handleStartMessage(msg *StartMessage) error {
	switch msg.Type {
	case MsgTypeInit:
		h.startMsg = msg
		h.hasStarted = false

		// Send OK message back to client.
		return h.Send(OkMessage)
	case MsgTypeResume:
		h.startMsg = msg
		h.hasStarted = true

		// Respond with a status message message so that we know to expect
		// a binary message next and the calling function knows where to
		// start sending data from.
		uploadedBytes, err := h.FS.GetUploadedBytes(msg.Identifier)
		if err != nil {
			return err
		}
		return h.Send(NewStatusMessage(uploadedBytes))
	}

	h.startMsg = nil
	return ErrInvalidMessageType
}

func (h *Handler) Send(msg interface{}) error {
	ctx, cancel := h.ctx()
	defer cancel()

	byt, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return h.conn.Write(ctx, websocket.MessageText, byt)
}

func (h *Handler) Close(status websocket.StatusCode, reason string) error {
	h.close <- true
	return h.conn.Close(status, reason)
}
