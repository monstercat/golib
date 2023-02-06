package logger

import "fmt"

// AppendableContextualEntry describes a context to which information
// can be appended. It implements ContextualEntry. It is also the default
// context for the logger.
//
// Usage:
// ```
//
//	 l.Log(
//		logger.SeverityInfo,
//		NewContextualPayload("my log"),
//	 )
//
// ```
//
// For more details, see SetContext.
type AppendableContextualEntry struct {
	Context interface{}   // The context to return
	Parent  []interface{} // Any parent data.
	Log     interface{}   // The log to return
}

// AppendableContext is used to provide better naming to the parent context.
type AppendableContext struct {
	Name    string      // Name for the context
	Context interface{} // Actual context information.
}

// NewContext creates a new AppendableContext. This can be added as a context to
// the Contextual logger to provide a better key/name for the log.
func NewContext(name string, value interface{}) *AppendableContext {
	return &AppendableContext{
		Name:    name,
		Context: value,
	}
}

// NewContextualPayload creates a new log payload for a Contextual logger.
func NewContextualPayload(log interface{}) *AppendableContextualEntry {
	return &AppendableContextualEntry{
		Log:     log,
		Context: make(map[string]interface{}),
	}
}

// Add sets the current context
func (e *AppendableContextualEntry) Add(context interface{}) *AppendableContextualEntry {
	e.Context = context
	return e
}

// SetContext sets the context for the current entry. Note that this entry may
// already have a context. Therefore, appending is always necessary.
//
// If the context provided is a map, then we add the map keys to the currently
// provided context.
//
// If the context provided is *not* a map, it adds the context into a "Parent"
// key.
func (e *AppendableContextualEntry) SetContext(ctx interface{}) {
	e.Parent = append(e.Parent, ctx)
}

func (e *AppendableContextualEntry) GetContext() interface{} {
	m := make(map[string]interface{})
	if v, ok := e.Context.(map[string]interface{}); ok {
		m = v
	} else {
		m["Current"] = e.Context
	}
	if len(e.Parent) > 0 {
		for i, p := range e.Parent {
			if v, ok := p.(*AppendableContext); ok {
				m[v.Name] = v.Context
			} else {
				m[fmt.Sprintf("Parent:%d", -i)] = p
			}
		}
	}
	return m
}

func (e *AppendableContextualEntry) GetLog() interface{} {
	return e.Log
}
