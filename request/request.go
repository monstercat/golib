package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	TypeJSONUTF8 = "application/json; charset=UTF-8"
	TypeJSON     = "application/json"
)

var (
	ErrNotFound = errors.New("not found")
)

var ErrStatusBounds = errors.New("out of bounds http status code")

type Params struct {
	ResponseBody string
	Headers      map[string]string
	Cookies      []*http.Cookie
	Method       string
	Password     string
	Request      *http.Request
	Response     *http.Response
	Timeout      time.Duration
	Url          string
	Username     string
}

type Error struct {
	Status  int
	Message interface{}
}

func (e *Error) Error() string {
	str, err := json.Marshal(e.Message)
	if err != nil {
		return ""
	}
	return string(str)
}

func RequestErr(p *Params, payload interface{}, body interface{}, e error) (error, bool) {
	err := Request(p, payload, body)
	if err == nil {
		return nil, false
	}
	if err := json.Unmarshal([]byte(p.ResponseBody), e); err != nil {
		return err, false
	}
	return e, true
}

func Get(uri string, params map[string][]string, response interface{}) (*http.Response, error) {
	if uri[len(uri)-1] == '?' {
		uri = uri[0:len(uri)-1]
	}
	if params != nil {
		uri += "?" + url.Values(params).Encode()
	}

	p := &Params{
		Timeout: 100000,
		Method: "GET",
		Url: uri,
	}
	if err := Request(p, nil, nil); err != nil {
		return nil, err
	}
	return p.Response, DecodeResponseBody(p, response)
}

func Post(uri string, payload, response interface{}) (*http.Response, error) {
	p := &Params{
		Timeout: 100000,
		Method:  "POST",
		Url:     uri,
	}
	if err := Request(p, payload, nil); err != nil {
		return nil, err
	}

	return p.Response, DecodeResponseBody(p, response)
}

func DecodeResponseBody(p *Params, response interface{}) error {
	resp := p.Response
	if resp.StatusCode == 404 {
		return &Error{Status: 404, Message: p.Url + " not found"}
	}
	if resp.StatusCode >= 400 {
		return &Error{Status: resp.StatusCode, Message: p.ResponseBody}
	}
	return json.Unmarshal([]byte(p.ResponseBody), response)
}

// Request makes an http request with params and optional payload
// to specified Url in params.Url.
// It will save respond body value into body.
func Request(params *Params, payload interface{}, body interface{}) error {
	// Assign payload support io.Reader, JSON, & FormData
	var _payload io.Reader
	var contentType string
	var contentLength int
	if payload != nil {
		if v, ok := payload.(io.Reader); ok {
			_payload = v
		} else if v, ok := payload.(url.Values); ok {
			_payload = strings.NewReader(v.Encode())
			contentType = "application/x-www-form-urlencoded"
			contentLength = len(v.Encode())
		} else {
			data, err := json.Marshal(payload)
			if err != nil {
				return err
			}
			_payload = bytes.NewReader(data)
			contentType = TypeJSONUTF8
			contentLength = len(string(data))
		}
	}

	// Apply defaults...
	if params.Timeout == 0 {
		params.Timeout = 10
	}

	//This line below can be uncommented to fix an x509 error that happens on mac when making
	//requests some times
	tr := &http.Transport{
		//TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Execute
	client := &http.Client{
		Timeout:   params.Timeout * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(params.Method, params.Url, _payload)
	if err != nil {
		return err
	}
	for header, value := range params.Headers {
		req.Header.Add(header, value)
	}
	if len(params.Cookies) > 0 {
		for _, c := range params.Cookies {
			req.AddCookie(c)
		}
	}

	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
		req.Header.Add("Content-Length", strconv.Itoa(contentLength))
	}
	if params.Username != "" || params.Password != "" {
		req.SetBasicAuth(params.Username, params.Password)
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	params.Request = req
	params.Response = res
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	params.ResponseBody = string(data)

	if body != nil && strings.Contains(res.Header.Get("Content-Type"), TypeJSON) {
		if err = json.Unmarshal(data, body); err != nil {
			return err
		}
	}

	if !StatusCodeInBounds(res.StatusCode) {
		return errors.Wrap(ErrStatusBounds, fmt.Sprintf("Got code %d and body %s", res.StatusCode, params.ResponseBody))
	}
	return nil
}

func StatusCodeInBounds(code int) bool {
	return code >= 200 && code < 400
}
