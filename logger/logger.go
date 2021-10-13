package logger

// Logger logs with severity.
type Logger interface {
	Log(severity Severity, payload interface{})
}

// Default defines a log that is able to set defaults
type Default interface {
	SetDefault()
}