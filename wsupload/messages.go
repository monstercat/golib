package wsupload

type MsgType string

var (
	MsgTypeInit   MsgType = "INIT"
	MsgTypeResume MsgType = "RESUME"
	MsgTypeStatus MsgType = "STATUS"
	MsgTypeOk     MsgType = "OK"
	MsgTypeError  MsgType = "ERROR"
)

type Message struct {
	Type MsgType
}

type StartMessage struct {
	Message
	Identifier    string
	ContentLength int
}

var OkMessage = Message{
	Type: MsgTypeOk,
}

type StatusMessage struct {
	Message
	Uploaded int
}

type ErrorMessage struct {
	Message
	Error string
}

func NewStatusMessage(uploaded int) StatusMessage {
	return StatusMessage{
		Message: Message{
			Type: MsgTypeStatus,
		},
		Uploaded: uploaded,
	}
}

func NewErrorMessage(err error) ErrorMessage {
	return ErrorMessage{
		Message: Message{
			Type: MsgTypeError,
		},
		Error: err.Error(),
	}
}
