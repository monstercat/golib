package twitch

import "errors"

const (
	ErrTypeBody    = "could not read body"
	ErrTypeJSON    = "json error"
	ErrTypeUnknown = "unknown"
)

var (
	ErrInvalidScope       = errors.New("invalid scope")
	ErrMissingClientId    = errors.New("missing client id")
	ErrMissingRedirectUri = errors.New("missing redirect uri")
	ErrMissingScopes      = errors.New("missing scopes")
	ErrNotFound           = errors.New("not found")
)

type Error struct {
	Type    string `json:"error"`
	Status  int    `json:"status"`
	Message string `json:"message"`
	Body    string `json:"-"`
}

func (e Error) Error() string {
	return e.Type + ": " + e.Message
}
