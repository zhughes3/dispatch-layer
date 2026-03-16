package log_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/zhughes3/dispatch-layer/pkg/log"
)

// capture creates a logger that writes to a buffer and returns both.
func capture(format log.Format, level slog.Level) (*log.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	l := log.New(log.Config{
		Format: format,
		Level:  level,
		Output: buf,
	})
	return l, buf
}

// ---------------------------------------------------------------------------
// Constructor / Config
// ---------------------------------------------------------------------------

func TestNew_Defaults(t *testing.T) {
	// Zero-value config should not panic.
	l := log.New(log.Config{})
	if l == nil {
		t.Fatal("New returned nil")
	}
}

func TestNew_TextFormat(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelDebug)
	l.Info("hello text")
	if !strings.Contains(buf.String(), "hello text") {
		t.Fatalf("expected 'hello text' in output, got: %s", buf.String())
	}
}

func TestNew_JSONFormat(t *testing.T) {
	l, buf := capture(log.FormatJSON, slog.LevelDebug)
	l.Info("hello json")
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if m["msg"] != "hello json" {
		t.Fatalf("unexpected msg field: %v", m["msg"])
	}
}

func TestNew_LevelFiltering(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelWarn)
	l.Debug("should be suppressed")
	l.Info("also suppressed")
	l.Warn("should appear")
	out := buf.String()
	if strings.Contains(out, "should be suppressed") {
		t.Error("DEBUG message leaked through WARN filter")
	}
	if strings.Contains(out, "also suppressed") {
		t.Error("INFO message leaked through WARN filter")
	}
	if !strings.Contains(out, "should appear") {
		t.Error("WARN message was incorrectly filtered out")
	}
}

func TestSetLevel_Dynamic(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelWarn)
	l.Debug("before change")
	l.SetLevel(slog.LevelDebug)
	l.Debug("after change")
	out := buf.String()
	if strings.Contains(out, "before change") {
		t.Error("debug log appeared before level change")
	}
	if !strings.Contains(out, "after change") {
		t.Error("debug log missing after level change")
	}
}

// ---------------------------------------------------------------------------
// Level methods
// ---------------------------------------------------------------------------

func TestDebug(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelDebug)
	l.Debug("dbg msg", "k", "v")
	out := buf.String()
	if !strings.Contains(out, "dbg msg") || !strings.Contains(out, "k=v") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestDebugf(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelDebug)
	l.Debugf("value is %d", 42)
	if !strings.Contains(buf.String(), "value is 42") {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}

func TestDebugContext(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelDebug)
	ctx := log.ContextWith(context.Background(), "trace_id", "abc123")
	l.DebugContext(ctx, "ctx debug")
	out := buf.String()
	if !strings.Contains(out, "trace_id=abc123") {
		t.Fatalf("context attribute missing: %s", out)
	}
}

func TestDebugContextf(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelDebug)
	ctx := log.ContextWith(context.Background(), "req", "xyz")
	l.DebugContextf(ctx, "req %s", "processed")
	out := buf.String()
	if !strings.Contains(out, "req=xyz") || !strings.Contains(out, "req processed") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestInfo(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelInfo)
	l.Info("info msg")
	if !strings.Contains(buf.String(), "info msg") {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}

func TestWarn(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelWarn)
	l.Warn("warn msg")
	if !strings.Contains(buf.String(), "warn msg") {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}

func TestError(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelError)
	l.Error("err msg")
	if !strings.Contains(buf.String(), "err msg") {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}

// ---------------------------------------------------------------------------
// Fatal (panics)
// ---------------------------------------------------------------------------

func TestFatal_Panics(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelDebug)
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Fatal did not panic")
		}
		if !strings.Contains(buf.String(), "fatal message") {
			t.Errorf("fatal log line missing: %s", buf.String())
		}
	}()
	l.Fatal("fatal message")
}

func TestFatalf_Panics(t *testing.T) {
	l, _ := capture(log.FormatText, slog.LevelDebug)
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Fatalf did not panic")
		}
	}()
	l.Fatalf("fatal %s", "formatted")
}

func TestFatalContext_Panics(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelDebug)
	ctx := log.ContextWith(context.Background(), "session", "s1")
	defer func() {
		r := recover()
		if r == nil {
			t.Error("FatalContext did not panic")
		}
		if !strings.Contains(buf.String(), "session=s1") {
			t.Errorf("context attribute missing in fatal: %s", buf.String())
		}
	}()
	l.FatalContext(ctx, "ctx fatal")
}

func TestFatalContextf_Panics(t *testing.T) {
	l, _ := capture(log.FormatText, slog.LevelDebug)
	defer func() {
		if recover() == nil {
			t.Error("FatalContextf did not panic")
		}
	}()
	l.FatalContextf(context.Background(), "fatal %d", 1)
}

// ---------------------------------------------------------------------------
// Context helpers
// ---------------------------------------------------------------------------

func TestContextWith_MultipleKeys(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelInfo)
	ctx := log.ContextWith(context.Background(), "a", 1, "b", "two")
	l.InfoContext(ctx, "multi")
	out := buf.String()
	if !strings.Contains(out, "a=1") || !strings.Contains(out, "b=two") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestContextWith_Accumulation(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelInfo)
	ctx := log.ContextWith(context.Background(), "first", "1")
	ctx = log.ContextWith(ctx, "second", "2")
	l.InfoContext(ctx, "acc")
	out := buf.String()
	if !strings.Contains(out, "first=1") || !strings.Contains(out, "second=2") {
		t.Fatalf("context attributes not accumulated: %s", out)
	}
}

func TestContextWith_NilContext(t *testing.T) {
	// Should not panic; nil context fields are skipped.
	l, buf := capture(log.FormatText, slog.LevelInfo)
	l.InfoContext(context.Background(), "no ctx attrs")
	if !strings.Contains(buf.String(), "no ctx attrs") {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}

// ---------------------------------------------------------------------------
// Entry builder
// ---------------------------------------------------------------------------

func TestEntry_With(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelInfo)
	l.With("foo", "bar").With("num", 99).Info("chained")
	out := buf.String()
	if !strings.Contains(out, "foo=bar") || !strings.Contains(out, "num=99") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestEntry_WithImmutability(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelInfo)
	base := l.With("base", "val")
	base.With("extra", "x").Info("with extra")
	base.Info("without extra")
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d:\n%s", len(lines), buf.String())
	}
	if !strings.Contains(lines[0], "extra=x") {
		t.Errorf("first line missing extra: %s", lines[0])
	}
	if strings.Contains(lines[1], "extra=x") {
		t.Errorf("second line should not have extra: %s", lines[1])
	}
}

func TestEntry_WithError(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelError)
	l.WithError(errors.New("something broke")).Error("op failed")
	if !strings.Contains(buf.String(), `error="something broke"`) {
		t.Fatalf("error attribute missing: %s", buf.String())
	}
}

func TestEntry_WithError_Nil(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelInfo)
	l.WithError(nil).Info("no error")
	if !strings.Contains(buf.String(), "error=<nil>") {
		t.Fatalf("nil error attribute missing: %s", buf.String())
	}
}

func TestEntry_WithDur(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelInfo)
	l.WithDur(250 * time.Millisecond).Info("timed op")
	if !strings.Contains(buf.String(), "duration") {
		t.Fatalf("duration attribute missing: %s", buf.String())
	}
}

func TestEntry_WithStructs(t *testing.T) {
	type User struct {
		ID    int    `slog:"user_id"`
		Email string `slog:"email"`
		pass  string // unexported — should be skipped
	}
	l, buf := capture(log.FormatText, slog.LevelInfo)
	u := User{ID: 7, Email: "test@example.com", pass: "secret"}
	l.WithStructs(u).Info("user event")
	out := buf.String()
	if !strings.Contains(out, "user_id=7") || !strings.Contains(out, "email=test@example.com") {
		t.Fatalf("struct fields missing: %s", out)
	}
	if strings.Contains(out, "secret") {
		t.Errorf("unexported field leaked: %s", out)
	}
}

func TestEntry_WithStructs_Pointer(t *testing.T) {
	type Req struct {
		Method string
	}
	l, buf := capture(log.FormatText, slog.LevelInfo)
	l.WithStructs(&Req{Method: "GET"}).Info("ptr struct")
	if !strings.Contains(buf.String(), "Method=GET") {
		t.Fatalf("pointer struct field missing: %s", buf.String())
	}
}

func TestEntry_WithStructs_NilPointer(t *testing.T) {
	type Req struct{ Method string }
	l, _ := capture(log.FormatText, slog.LevelInfo)
	var r *Req
	// Should not panic.
	l.WithStructs(r).Info("nil ptr")
}

func TestEntry_WithStructs_MultipleStructs(t *testing.T) {
	type A struct{ X int }
	type B struct{ Y string }
	l, buf := capture(log.FormatText, slog.LevelInfo)
	l.WithStructs(A{X: 1}, B{Y: "hello"}).Info("multi struct")
	out := buf.String()
	if !strings.Contains(out, "X=1") || !strings.Contains(out, "Y=hello") {
		t.Fatalf("multi-struct fields missing: %s", out)
	}
}

func TestEntry_ContextAttributes(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelInfo)
	ctx := log.ContextWith(context.Background(), "span", "s99")
	l.With("entry_key", "eVal").InfoContext(ctx, "entry+ctx")
	out := buf.String()
	if !strings.Contains(out, "entry_key=eVal") || !strings.Contains(out, "span=s99") {
		t.Fatalf("entry or context attr missing: %s", out)
	}
}

func TestEntry_Fatal_Panics(t *testing.T) {
	l, buf := capture(log.FormatText, slog.LevelDebug)
	defer func() {
		if recover() == nil {
			t.Error("entry Fatal did not panic")
		}
		if !strings.Contains(buf.String(), "critical") {
			t.Errorf("fatal line missing: %s", buf.String())
		}
	}()
	l.With("k", "v").Fatal("critical")
}

// ---------------------------------------------------------------------------
// JSON output completeness
// ---------------------------------------------------------------------------

func TestJSON_EntryFields(t *testing.T) {
	l, buf := capture(log.FormatJSON, slog.LevelInfo)
	l.With("service", "api").WithError(errors.New("oops")).Info("json entry")
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if m["service"] != "api" {
		t.Errorf("service field wrong: %v", m["service"])
	}
	if m["error"] != "oops" {
		t.Errorf("error field wrong: %v", m["error"])
	}
}
