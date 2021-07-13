package youtube

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	BaseUrlV3 = "https://www.googleapis.com/youtube/v3"
)

type Request struct {
	Method string
	Url    string
	Params url.Values
	Body   io.Reader

	// Each request must either specify an APIKey or provide an OAUTH token.
	// This struct will default to AccessToken if provided
	// @see https://developers.google.com/youtube/v3/docs
	ApiKey      string
	AccessToken string

	/// Timeout for the request
	Timeout time.Duration
}

func Run(r *Request) (*http.Response, error) {
	if r.AccessToken == "" && r.ApiKey != "" {
		r.Params.Add("key", r.ApiKey)
	}
	req, err := http.NewRequest(r.Method, r.Url+"?"+r.Params.Encode(), r.Body)
	if err != nil {
		return nil, err
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
	if res.StatusCode >= 400 {
		var e Error
		if err := json.NewDecoder(res.Body).Decode(&e); err == nil {
			e.StatusCode = res.StatusCode
			return e
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		e.StatusCode = res.StatusCode
		e.ErrorType = ErrTypeUnknown
		e.Description = string(body)
		return e
	}

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		body, bodyErr := ioutil.ReadAll(res.Body)
		if bodyErr != nil {
			return Error{
				ErrorType:   ErrTypeBody,
				Description: bodyErr.Error(),
			}
		}
		return Error{
			ErrorType:   ErrTypeJSON,
			Description: err.Error(),
			Body:        string(body),
		}
	}

	return nil
}
