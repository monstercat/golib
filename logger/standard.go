package logger

import (
	"encoding/json"
	"fmt"
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

func (l *Standard) prepareInterface(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%#v", v)
		}
		return string(b)
	}
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
	case ContextualEntry:
		// We want to separate the "Log" and the context info in the case that this is a contextual type.
		log.Printf("[%s] %s", severity, l.prepareInterface(v.GetLog()))
		log.Printf("     %s", l.prepareInterface(v.GetContext()))
	default:
		log.Printf("[%s] %s", severity, l.prepareInterface(v))
	}

	if l.PrintStack {
		debug.PrintStack()
	}
}

