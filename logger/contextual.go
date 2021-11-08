package logger

// StandardContextEntry is what will be passed into the logger by default.
type StandardContextEntry struct {
	Context interface{}
	Log     interface{}
}

func (e StandardContextEntry) SetContext(ctx interface{}) {
	e.Context = ctx
}

func (e StandardContextEntry) GetContext() interface{} {
	return e.Context
}

func (e StandardContextEntry) GetLog() interface{} {
	return e.Log
}

// ContextualEntry is a log that can have a context. It will be used by the Context to insert a context. This
// is if the log should have a specific structure regardless of whether a context is provided or not.
type ContextualEntry interface {
	SetContext(interface{})
	GetContext() interface{}
	GetLog() interface{}
}

// Contextual allows the user to input extra context information into their logs.
// It takes in a logger that actually does the logging work, but amends the payload
// to include some context information.
//
// To use the Contextual, wrap another .
//
// For example:
// logger := &Contextual{
//    Logger: baseLogger,
//    Context: map[string]interface{} {...}
// }
//
// ...
// logger.Log(LogSeverityInfo, "Some log") should present a log that has:
// {
//    Context: ...
//    Log: "Some log"
// }
//
// In the case that the provided log is a struct that implements the ContextualLogEntry interface, it will instead use
// the SetContext method.
//
// logger.Log(LogSeverityInfo, LogMetricEntry{...})
// {
//    ... parts of LogMetricEntry
//    Context: {...}  // This variable is defined in LogMetricEntry and is used to store the Context for that entry.
// }
type Contextual struct {
	Logger
	Context interface{}
}

func (l *Contextual) Log(severity Severity, payload interface{}) {
	if e, ok := payload.(ContextualEntry); ok {
		e.SetContext(l.Context)
		l.Logger.Log(severity, e)
		return
	}

	l.Logger.Log(severity, StandardContextEntry{
		Context: l.Context,
		Log:     payload,
	})
}
