package twitch

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Request struct {
	Method string
	Url    string
	Params url.Values
	Body   io.Reader

	ClientId    string
	AccessToken string

	/// Timeout for the request
	Timeout time.Duration
}

func Run(r *Request) (*http.Response, error) {
	req, err := http.NewRequest(r.Method, r.Url+"?"+r.Params.Encode(), r.Body)
	if err != nil {
		return nil, err
	}

	// AccessToken and ClientId are required, except for the ExchangeOAuthToken route. This is just to make sure
	// that the token will go through.
	if r.ClientId != "" {
		req.Header.Add("client-id", r.ClientId)
	}
	if r.AccessToken != "" {
		req.Header.Add("Authorization", "Bearer "+r.AccessToken)
	}

	client := http.Client{
		Timeout: r.Timeout,
	}
	return client.Do(req)
}

func DecodeResponse(res *http.Response, out interface{}) error {
	// Twitch errors are in the range of 400 - 503
	if res.StatusCode >= 400 {
		var e Error
		if err := json.NewDecoder(res.Body).Decode(&e); err == nil {
			return e
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		e.Status = res.StatusCode
		e.Type = ErrTypeUnknown
		e.Message = string(body)
		return e
	}

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		body, bodyErr := ioutil.ReadAll(res.Body)
		if bodyErr != nil {
			return Error{
				Type:    ErrTypeBody,
				Message: bodyErr.Error(),
			}
		}
		return Error{
			Type:    ErrTypeJSON,
			Message: err.Error(),
			Body:    string(body),
		}
	}
	return nil
}
