package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func newTestColoredLogger() (*StandardColoredLogger, *bytes.Buffer, *bytes.Buffer) {
	var out, errOut bytes.Buffer
	lgr := &StandardColoredLogger{
		log: log.New(&out, "", 0),
		err: log.New(&errOut, "", 0),
	}
	return lgr, &out, &errOut
}

func TestColoredLog_String(t *testing.T) {
	lgr, out, _ := newTestColoredLogger()
	lgr.Log(SeverityInfo, "hello world")
	if !strings.Contains(out.String(), "hello world") {
		t.Errorf("expected output to contain %q, got %q", "hello world", out.String())
	}
}

func TestColoredLog_ErrorGoesToStderr(t *testing.T) {
	lgr, out, errOut := newTestColoredLogger()
	lgr.Log(SeverityError, "something broke")
	if out.Len() != 0 {
		t.Errorf("expected stdout to be empty for error severity, got %q", out.String())
	}
	if !strings.Contains(errOut.String(), "something broke") {
		t.Errorf("expected stderr to contain %q, got %q", "something broke", errOut.String())
	}
}

func TestColoredLog_ContextualEntry(t *testing.T) {
	lgr, out, _ := newTestColoredLogger()
	payload := NewContextualPayload("contextual message").Add(map[string]interface{}{
		"key": "value",
	})
	lgr.Log(SeverityInfo, payload)
	if !strings.Contains(out.String(), "contextual message") {
		t.Errorf("expected output to contain %q, got %q", "contextual message", out.String())
	}
}

func TestColoredLog_BlockList(t *testing.T) {
	lgr, out, _ := newTestColoredLogger()
	lgr.BlockList = []Severity{SeverityInfo}
	lgr.Log(SeverityInfo, "should be blocked")
	if out.Len() != 0 {
		t.Errorf("expected blocked severity to produce no output, got %q", out.String())
	}
}

func TestColorForSeverity(t *testing.T) {
	lgr := &StandardColoredLogger{}
	cases := []struct {
		sev      Severity
		expected string
	}{
		{SeverityNotice, StdOutColorBlue},
		{SeverityWarning, StdOutColorYellow},
		{SeverityError, StdOutColorRed},
		{SeverityCritical, StdOutColorPurple},
		{SeverityInfo, StdOutColorGray},
		{SeverityDebug, StdOutColorReset},
	}
	for _, c := range cases {
		got := lgr.ColorForSeverity(c.sev)
		if got != c.expected {
			t.Errorf("ColorForSeverity(%s) = %q, want %q", c.sev, got, c.expected)
		}
	}
}
