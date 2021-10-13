package logger

import (
	"encoding/json"
	"log"
	"runtime/debug"
)

// Standard prints to standard output.
//
// Severity will be printed within square brackets (e.g., [Info]).
// If the payload is a string, it will print it inline.
//
// Otherwise, it will attempt to JSON encode it (with indents) and will print the message on a separate line.
// If JSON marshalling fails, it will print the go representation of the payload.
type Standard struct {
	PrintStack bool
	BlockList []Severity
}

// Log logs the payload to the standard console. Debug logs will only be shown if debug is true.
func (l *Standard) Log(severity Severity, payload interface{}) {
	if InList(l.BlockList, severity) {
		return
	}

	if p, ok := payload.(Default); ok {
		p.SetDefault()
	}

	switch v := payload.(type) {
	case string:
		log.Printf("[%s] %s", severity, v)
	default:
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			log.Printf("[%s]\n%#v", severity, payload)
			return
		}
		log.Printf("[%s]\n%s", severity, b)
	}

	if l.PrintStack {
		debug.PrintStack()
	}
}

