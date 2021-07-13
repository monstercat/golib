package youtube

import "errors"

var (
	ErrInvalidAccessType  = errors.New("invalid access type")
	ErrInvalidPart        = errors.New("invalid part")
	ErrInvalidPrompt      = errors.New("invalid prompt")
	ErrInvalidScope       = errors.New("invalid scope")
	ErrMissingClientId    = errors.New("missing client id")
	ErrMissingParts       = errors.New("missing parts")
	ErrMissingRedirectUri = errors.New("missing redirect uri")
	ErrMissingScopes      = errors.New("missing scopes")
	ErrNotFound           = errors.New("not found")
)

const (
	ErrAccessDenied        OAuthError = "access_denied"
	ErrAdminPolicyEnforced OAuthError = "admin_policy_enforced"
	ErrDisallowedUserAgent OAuthError = "disallowed_useragent"
	ErrOrgInternal         OAuthError = "org_internal"
	ErrRedirectUriMismatch OAuthError = "redirect_uri_mismatch"

	ErrTypeBody         ErrorType = "could not read body"
	ErrTypeInvalidGrant ErrorType = "invalid_grant"
	ErrTypeJSON         ErrorType = "json_error"
	ErrTypeUnknown      ErrorType = "unknown"
)

type OAuthError string
type ErrorType string

func (e OAuthError) Error() string {
	return string(e)
}

type Error struct {
	StatusCode  int       `json:"-"`
	ErrorType   ErrorType `json:"error"`
	Description string    `json:"error_description"`
	Body        string    `json:"body"`
}

func (e Error) Error() string {
	return string(e.ErrorType) + ": " + e.Description
}
