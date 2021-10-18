package logger

import (
	"encoding/json"

	"cloud.google.com/go/logging"
	"github.com/golang/protobuf/ptypes"
	logtypepb "google.golang.org/genproto/googleapis/logging/type"
)

// Google logs to the Google cloud console logging platform.
type Google struct {
	// BaseEntry is used to define parameters other than Severity and Payload on the Google Logger.
	BaseEntry logging.Entry
	BlockList []Severity

	// Logger here is what is actually doing the logging work. This interface is primarily for testing purposes.
	// For example, we need to make sure that the logger discards all debug logs.
	Logger interface {
		Log(entry logging.Entry)
	}
}

// NewGoogle returns a new Google logger using the provided client and name. Note that this method creates a new
// *logging.Logger which is expensive (as each *logging.Logger can store up to 1000 entries and 1 MB of data before
// sending the logs and clearing its own cache).
func NewGoogle(client *logging.Client, name string) *Google {
	return &Google{
		Logger: client.Logger(name),
	}
}

// WithBaseEntry populates the base entry for the google log, which will be used for all google logs.
func (l *Google) WithBaseEntry(b logging.Entry) *Google {
	return &Google{
		BaseEntry: b,
		Logger:    l.Logger,
		BlockList: l.BlockList,
	}
}

func (l *Google) WithBlockList(s ...Severity) *Google {
	return &Google{
		BaseEntry: l.BaseEntry,
		Logger:    l.Logger,
		BlockList: s,
	}
}

// Log an entry to the Google cloud logging console.
func (l *Google) Log(severity Severity, payload interface{}) {
	// As per spec, we need to discard all debug logs. These are for local development only.
	if InList(l.BlockList, severity) {
		return
	}

	if p, ok := payload.(Default); ok {
		p.SetDefault()
	}

	l.Logger.Log(logging.Entry{
		Timestamp:   l.BaseEntry.Timestamp,
		Severity:    severity.ToGoogleSeverity(),
		Payload:     payload,
		Labels:      l.BaseEntry.Labels,
		InsertID:    l.BaseEntry.InsertID,
		HTTPRequest: l.BaseEntry.HTTPRequest,
		Operation:   l.BaseEntry.Operation,
		LogName:     l.BaseEntry.LogName,
		Resource:    l.BaseEntry.Resource,
		Trace:       l.BaseEntry.Trace,
	})
}

// ToGoogleSeverity parses the LogSeverity into someting usable for Google.
func (s Severity) ToGoogleSeverity() logging.Severity {
	return logging.ParseSeverity(string(s))
}

// HTTPRequestContext helps with marshalling the *logging.HTTPRequest from Google. This can be used for the
// logger.Contextual so that the context is a proper JSON string.
type HTTPRequestContext struct {
	*logging.HTTPRequest
}

func NewHTTPRequestContext(req *logging.HTTPRequest) *HTTPRequestContext {
	return &HTTPRequestContext{
		HTTPRequest: req,
	}
}

// Copied from logging.fromHTTPRequest
func (c *HTTPRequestContext) toJsonable() *logtypepb.HttpRequest {
	if c == nil {
		return nil
	}
	if c.Request == nil {
		panic("HTTPRequest must have a non-nil Request")
	}
	u := *c.Request.URL
	u.Fragment = ""
	pb := &logtypepb.HttpRequest{
		RequestMethod:                  c.Request.Method,
		RequestUrl:                     u.String(),
		RequestSize:                    c.RequestSize,
		Status:                         int32(c.Status),
		ResponseSize:                   c.ResponseSize,
		UserAgent:                      c.Request.UserAgent(),
		ServerIp:                       c.LocalIP,
		RemoteIp:                       c.RemoteIP, // TODO(jba): attempt to parse http.Request.RemoteAddr?
		Referer:                        c.Request.Referer(),
		CacheHit:                       c.CacheHit,
		CacheValidatedWithOriginServer: c.CacheValidatedWithOriginServer,
	}
	if c.Latency != 0 {
		pb.Latency = ptypes.DurationProto(c.Latency)
	}
	return pb
}

// MarshalJSON allows for custom JSON marshalling.
func (c *HTTPRequestContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.toJsonable())
}
