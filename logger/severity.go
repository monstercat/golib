package logger

type Severity string

const (
	// SeverityDebug is for local development purposes only. If Debug is set to false in the CLI, the loggers should
	// ignore this.
	SeverityDebug Severity= "Debug"

	// SeverityInfo
	// These log lines contain useful information about the state or result of a current action.  Most logs with these
	// lines will typically contain the request and response status of a served request.
	SeverityInfo Severity= "Info"

	// SeverityNotice
	// These log lines typically contain information that you want specifically to stand out amongst the other lines.
	// These should be helpful messages. No redundant lines.
	SeverityNotice Severity= "Notice"

	// SeverityWarning
	// These log lines should contain information about an error that was encountered and recovered. There should also
	// be a corresponding error message that this resolved. These lines should also contain information that can allow
	//us to take preventative efforts for possible future errors.
	SeverityWarning Severity= "Warning"

	// SeverityError
	// All errors encountered should be logged with this level.
	SeverityError Severity= "Error"

	// SeverityCritical
	// Any system level error encountered either hardware related (file system errors, out of memory errors, database
	//errors) or in the application bootstrapping process that prevents the server from running
	SeverityCritical Severity= "Critical"
)

func InList(haystack []Severity, needle Severity) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}