package youtube

import (
	"net/url"
	"strings"
)

const (
	OAuthUrl = "https://accounts.google.com/o/oauth2/v2/auth"
)

type Scope string
type AccessType string
type Prompt string

const (
	ScopeAccount            Scope = "https://www.googleapis.com/auth/youtube"
	ScopeChannelMemberships Scope = "https://www.googleapis.com/auth/youtube.channel-memberships.creator"
	ScopeForceSSL           Scope = "https://www.googleapis.com/auth/youtube.force-ssl"
	ScopeReadOnly           Scope = "https://www.googleapis.com/auth/youtube.readonly"
	ScopeUpload             Scope = "https://www.googleapis.com/auth/youtube.upload"
	ScopePartner            Scope = "https://www.googleapis.com/auth/youtubepartner"
	ScopeAudit              Scope = "https://www.googleapis.com/auth/youtubepartner-channel-audit"

	AccessTypeOnline  AccessType = "online"
	AccessTypeOffline AccessType = "offline"

	PromptNone    Prompt = "none"
	PromptConsent Prompt = "consent"
	PromptSelect  Prompt = "select_account"
)

type OAuthOptions struct {
	ClientId             string
	RedirectUri          string
	Scopes               []Scope
	AccessType           AccessType
	State                string
	IncludeGrantedScopes bool
	LoginHint            bool
	Prompts              []Prompt
}

func (p Prompt) IsValid() bool {
	switch p {
	case PromptNone, PromptConsent, PromptSelect:
		return true
	}
	return false
}

func (t AccessType) IsValid() bool {
	if t == "" {
		return true
	}
	switch t {
	case AccessTypeOnline, AccessTypeOffline:
		return true
	}
	return false
}

func (s Scope) IsValid() bool {
	switch s {
	case ScopeAudit, ScopeAccount, ScopeChannelMemberships, ScopeForceSSL, ScopeReadOnly, ScopeUpload, ScopePartner:
		return true
	}
	return false
}

func (o OAuthOptions) Validate() error {
	if o.ClientId == "" {
		return ErrMissingClientId
	}
	if o.RedirectUri == "" {
		return ErrMissingRedirectUri
	}
	if !o.AccessType.IsValid() {
		return ErrInvalidAccessType
	}
	if len(o.Scopes) == 0 {
		return ErrMissingScopes
	}
	for _, s := range o.Scopes {
		if !s.IsValid() {
			return ErrInvalidScope
		}
	}
	for _, p := range o.Prompts {
		if !p.IsValid() {
			return ErrInvalidPrompt
		}
	}
	return nil
}

func (o OAuthOptions) convertScopes() []string {
	if len(o.Scopes) == 0 {
		return []string{}
	}
	s := make([]string, 0, len(o.Scopes))
	for _, scope := range o.Scopes {
		s = append(s, string(scope))
	}
	return s
}

func (o OAuthOptions) convertPrompts() []string {
	if len(o.Prompts) == 0 {
		return []string{}
	}
	p := make([]string, 0, len(o.Prompts))
	for _, prompt := range o.Prompts {
		p = append(p, string(prompt))
	}
	return p
}

func (o OAuthOptions) Values() url.Values {
	v := url.Values{}
	v.Add("client_id", o.ClientId)
	v.Add("redirect_uri", o.RedirectUri)
	v.Add("response_type", "code")
	v.Add("scope", strings.Join(o.convertScopes(), " "))

	if o.AccessType != "" {
		v.Add("access_type", string(o.AccessType))
	}
	if o.State != "" {
		v.Add("state", o.State)
	}
	if o.IncludeGrantedScopes {
		v.Add("include_granted_scopes", "true")
	}
	if o.LoginHint {
		v.Add("login_hint", "true")
	}
	if len(o.Prompts) > 0 {
		v.Add("prompt", strings.Join(o.convertPrompts(), " "))
	}
	return v
}

func GenerateOAuthUrl(options OAuthOptions) (*url.URL, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}
	vals := options.Values()
	return url.Parse(OAuthUrl + "?" + vals.Encode())
}
