package logger

import (
	"testing"

)

type TestLogItem struct {
	Severity Severity
	Payload  interface{}
}

func (i *TestLogItem) TestSeverity(t *testing.T, sev Severity) *TestLogItem {
	if i == nil {
		t.Errorf("missing log item for severity test %s", sev)
		return  i
	}
	if i.Severity != sev {
		t.Errorf("log item severity is %s. Expecting %s.", i.Severity, sev)
	}
	return i
}


func (i *TestLogItem) TestString(t *testing.T, str string) *TestLogItem {
	if i == nil {
		t.Error("missing log item for string '" + str + "'")
		return i
	}
	s, ok := i.Payload.(string)
	if !ok {
		t.Error("log item is not a string. Testing string '" + str + "'")
		return i
	}
	if s != str {
		t.Errorf("Log Item does not match. Got '%s', expect '%s'", s, str)
	}
	return i
}

type TestLogger struct {
	Logs []TestLogItem
}

func (l *TestLogger) Clear() {
	l.Logs = make([]TestLogItem, 0)
}

func (l *TestLogger) Log(severity Severity, payload interface{}) {
	l.Logs = append(l.Logs, TestLogItem{
		Severity: severity,
		Payload:  payload,
	})
}

func (l *TestLogger) Unshift() *TestLogItem {
	if len(l.Logs) == 0 {
		return nil
	}
	log := l.Logs[0]
	l.Logs = l.Logs[1:]
	return &log
}
