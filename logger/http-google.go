package logger

import (
	"time"

	"cloud.google.com/go/logging"
)

// GoogleHTTPRequest is a logger that stores information into the LogRequest variable. It should be used in conjunction
// with the Google logger.
//
// e.g.,
//
// req := &logging.HTTPRequest{Request: req}
// logger := &logger.GoogleHTTPRequest{
//    Logger: &MultiLogger{ // Can be any logger as long as Google is somewhere in here.
//       Loggers: []Logger{
//          &logger.Standard{}, // or ContextLogger,
//          logger.NewGoogle(...).WithBaseEntry(req)
//       },
//       LogRequest: req,
//    }
// }
type GoogleHTTPRequest struct {
	Logger
	LogRequest *logging.HTTPRequest

	startTime time.Time
}

// StartTimer - see HTTPRequest.StartTimer
func (l *GoogleHTTPRequest) StartTimer() {
	l.startTime = time.Now()
}


// SetStatus - see HTTPRequest.SetStatus
func (l *GoogleHTTPRequest) SetStatus(status int) {
	if l.LogRequest == nil {
		return
	}
	l.LogRequest.Status = status
}

// SetCached - see HTTPRequest.SetCached
func (l *GoogleHTTPRequest) SetCached(cached bool) {
	if l.LogRequest == nil {
		return
	}
	l.LogRequest.CacheHit = cached
}

// SetLatency - see HTTPRequest.SetLatency
func (l *GoogleHTTPRequest) SetLatency() {
	if l.LogRequest == nil {
		return
	}
	l.LogRequest.Latency = time.Now().Sub(l.startTime)
}
