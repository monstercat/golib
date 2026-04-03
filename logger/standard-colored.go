package logger

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
)

const (
	StdOutColorReset  = "\033[0m"
	StdOutColorRed    = "\033[31m"
	StdOutColorGreen  = "\033[32m"
	StdOutColorYellow = "\033[33m"
	StdOutColorBlue   = "\033[34m"
	StdOutColorPurple = "\033[35m"
	StdOutColorCyan   = "\033[36m"
	StdOutColorGray   = "\033[37m"
	StdOutColorWhite  = "\033[97m"
)

// StandardColoredLogger extends Standard with ANSI color-coded output.
// Regular log messages are written to stdout; error messages are written to stderr
// with the short filename included for easier debugging.
type StandardColoredLogger struct {
	Standard

	log *log.Logger
	err *log.Logger
}

// NewColored returns a StandardColoredLogger with default loggers writing to stdout and stderr.
func NewColored() *StandardColoredLogger {
	return &StandardColoredLogger{
		log: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.LUTC),
		err: log.New(os.Stderr, "", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile),
	}
}

// SetLogger replaces the standard output logger.
func (std *StandardColoredLogger) SetLogger(lgr *log.Logger) *StandardColoredLogger {
	std.log = lgr
	return std
}

// SetErrorLogger replaces the error output logger.
func (std *StandardColoredLogger) SetErrorLogger(lgr *log.Logger) *StandardColoredLogger {
	std.err = lgr
	return std
}

// ColorForSeverity returns the ANSI color code corresponding to the given severity level.
func (std *StandardColoredLogger) ColorForSeverity(sev Severity) string {
	switch sev {
	case SeverityNotice:
		return StdOutColorBlue
	case SeverityWarning:
		return StdOutColorYellow
	case SeverityError:
		return StdOutColorRed
	case SeverityCritical:
		return StdOutColorPurple
	case SeverityInfo:
		return StdOutColorGray
	}
	return StdOutColorReset
}

// Log writes a color-coded message to the appropriate output. Critical and error severities
// go to stderr; all others go to stdout. For ContextualEntry payloads, the log and context
// are printed on separate lines. If PrintStack is set, a stack trace is printed after each entry.
func (std *StandardColoredLogger) Log(severity Severity, payload interface{}) {
	if InList(std.BlockList, severity) {
		return
	}

	if p, ok := payload.(Default); ok {
		p.SetDefault()
	}

	color := std.ColorForSeverity(severity)
	sev := fmt.Sprintf("%s[%s]%s", color, severity, StdOutColorReset)

	lgr := std.log
	switch severity {
	case SeverityCritical, SeverityError:
		lgr = std.err
	}

	switch v := payload.(type) {
	case string:
		lgr.Printf("%s %s", sev, v)
	case ContextualEntry:
		// We want to separate the "Log" and the context info in the case that this is a contextual type.
		lgr.Printf("%s %s%s%s", sev, color, std.prepareInterface(v.GetLog()), StdOutColorReset)
		lgr.Printf("     %s%s%s", StdOutColorGray, std.prepareInterface(v.GetContext()), StdOutColorReset)
	default:
		lgr.Printf("     %s%s%s", StdOutColorGray, std.prepareInterface(v), StdOutColorReset)
	}

	if std.PrintStack {
		debug.PrintStack()
	}
}
