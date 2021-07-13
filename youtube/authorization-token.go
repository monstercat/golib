package youtube

import (
	"net/http"
	"net/url"
	"time"
)

const (
	ExchangeOAuthTokenUrl = "https://oauth2.googleapis.com/token"
)

// Token details the access and refresh tokens as well as allowed scopes.
type Token struct {
	AccessToken   string `json:"access_token"`
	ExpiresInSecs int    `json:"expires_in"`
	RefreshToken  string `json:"refresh_token"`
	Scope         string `json:"scope"`
	TokenType     string `json:"token_type"`
}

// ExchangeAuthToken exchanges an authorization token retrieved through YouTube's OAUTH flow for a Token which contains
// the access token, refresh token, and expiry. It will also contain the TokenType, but that field is always set to
// "Bearer".
// @see https://developers.google.com/youtube/v3/guides/auth/server-side-web-apps#exchange-authorization-code
func ExchangeAuthToken(clientId, clientSecret, code, redirect string, timeout time.Duration) (*Token, error) {
	vals := url.Values{}
	vals.Add("client_id", clientId)
	vals.Add("client_secret", clientSecret)
	vals.Add("grant_type", "authorization_code")
	vals.Add("code", code)
	vals.Add("redirect_uri", redirect)

	res, err := Run(&Request{
		Method:  http.MethodPost,
		Url:     ExchangeOAuthTokenUrl,
		Params:  vals,
		Timeout: timeout,
	})
	if err != nil {
		return nil, err
	}

	var t Token
	if err := DecodeResponse(res, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
