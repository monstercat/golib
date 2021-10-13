package logger

// Multi will log to multiple outputs.
type Multi struct {
	Loggers []Logger
}

// Log will log the payload into all provided loggers
func (l *Multi) Log(severity Severity, payload interface{}) {
	for _, l := range l.Loggers {
		l.Log(severity, payload)
	}
}

