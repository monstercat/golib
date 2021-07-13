package twitch

import (
	"net/http"
	"net/url"
	"time"
)

const (
	ExchangeOAuthTokenUrl = "https://id.twitch.tv/oauth2/token"
)

// Token details the access and refresh tokens as well as allowed scopes.
type Token struct {
	AccessToken   string   `json:"access_token"`
	ExpiresInSecs int      `json:"expires_in"`
	RefreshToken  string   `json:"refresh_token"`
	Scope         []string `json:"scope"`
	TokenType     string   `json:"token_type"`
}

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
