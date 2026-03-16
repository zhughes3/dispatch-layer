package log

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"
)

// Entry is a fluent log-entry builder. Attributes are accumulated via With*
// methods and the entry is emitted by calling a terminal method such as
// Debug, Info, Warn, Error, or Fatal.
//
// Entry values should not be stored or reused across goroutines after any
// terminal method has been called.
type Entry struct {
	logger *Logger
	attrs  []slog.Attr
}

// newEntry allocates a fresh Entry bound to the given logger.
func newEntry(l *Logger) *Entry {
	return &Entry{logger: l}
}

// clone returns a shallow copy of the entry so that chained With calls do
// not mutate the original.
func (e *Entry) clone() *Entry {
	c := &Entry{
		logger: e.logger,
		attrs:  make([]slog.Attr, len(e.attrs)),
	}
	copy(c.attrs, e.attrs)
	return c
}

// ---------------------------------------------------------------------------
// Builder methods
// ---------------------------------------------------------------------------

// With adds a single key/value attribute to the entry.
//
// Example:
//
//	logger.With("order_id", 42).With("status", "pending").Info("order updated")
func (e *Entry) With(key string, value any) *Entry {
	c := e.clone()
	c.attrs = append(c.attrs, slog.Any(key, value))
	return c
}

// WithStructs extracts all exported fields from each provided struct (or
// pointer to struct) and adds them as log attributes. Unexported fields are
// silently skipped. Non-struct arguments are also skipped.
//
// Text format: fields are added flat at the top level (key=value …).
//
// JSON format: each struct is nested under its type name as a JSON object:
//
//	{"order": {"order_id": 7, "amount": 49.99}}
func (e *Entry) WithStructs(args ...any) *Entry {
	c := e.clone()
	for _, arg := range args {
		if arg == nil {
			continue
		}
		if e.logger.format == FormatJSON {
			if a, ok := structToGroup(arg); ok {
				c.attrs = append(c.attrs, a)
			}
		} else {
			c.attrs = append(c.attrs, structToAttrs(arg)...)
		}
	}
	return c
}

// WithError adds an "error" attribute to the entry. If err is nil the
// attribute value is set to "<nil>".
//
// Example:
//
//	logger.WithError(err).Error("database query failed")
func (e *Entry) WithError(err error) *Entry {
	c := e.clone()
	if err != nil {
		c.attrs = append(c.attrs, slog.String("error", err.Error()))
	} else {
		c.attrs = append(c.attrs, slog.String("error", "<nil>"))
	}
	return c
}

// WithDur adds a "duration" attribute to the entry.
//
// Example:
//
//	logger.WithDur(elapsed).Info("operation completed")
func (e *Entry) WithDur(dur time.Duration) *Entry {
	c := e.clone()
	c.attrs = append(c.attrs, slog.Duration("duration", dur))
	return c
}

// ---------------------------------------------------------------------------
// Terminal methods
// ---------------------------------------------------------------------------

// Debug emits the entry at DEBUG level.
func (e *Entry) Debug(msg string) {
	e.logger.slog.LogAttrs(context.Background(), slog.LevelDebug, msg, e.withSource(2)...)
}

// Debugf emits the entry at DEBUG level with a formatted message.
func (e *Entry) Debugf(format string, args ...any) {
	e.logger.slog.LogAttrs(context.Background(), slog.LevelDebug, fmt.Sprintf(format, args...), e.withSource(2)...)
}

// DebugContext emits the entry at DEBUG level, also including context attributes.
func (e *Entry) DebugContext(ctx context.Context, msg string) {
	e.logger.slog.LogAttrs(ctx, slog.LevelDebug, msg, e.withSourceAndCtx(2, ctx)...)
}

// DebugContextf emits a formatted entry at DEBUG level with context attributes.
func (e *Entry) DebugContextf(ctx context.Context, format string, args ...any) {
	e.logger.slog.LogAttrs(ctx, slog.LevelDebug, fmt.Sprintf(format, args...), e.withSourceAndCtx(2, ctx)...)
}

// Info emits the entry at INFO level.
func (e *Entry) Info(msg string) {
	e.logger.slog.LogAttrs(context.Background(), slog.LevelInfo, msg, e.withSource(2)...)
}

// Infof emits the entry at INFO level with a formatted message.
func (e *Entry) Infof(format string, args ...any) {
	e.logger.slog.LogAttrs(context.Background(), slog.LevelInfo, fmt.Sprintf(format, args...), e.withSource(2)...)
}

// InfoContext emits the entry at INFO level with context attributes.
func (e *Entry) InfoContext(ctx context.Context, msg string) {
	e.logger.slog.LogAttrs(ctx, slog.LevelInfo, msg, e.withSourceAndCtx(2, ctx)...)
}

// InfoContextf emits a formatted entry at INFO level with context attributes.
func (e *Entry) InfoContextf(ctx context.Context, format string, args ...any) {
	e.logger.slog.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf(format, args...), e.withSourceAndCtx(2, ctx)...)
}

// Warn emits the entry at WARN level.
func (e *Entry) Warn(msg string) {
	e.logger.slog.LogAttrs(context.Background(), slog.LevelWarn, msg, e.withSource(2)...)
}

// Warnf emits the entry at WARN level with a formatted message.
func (e *Entry) Warnf(format string, args ...any) {
	e.logger.slog.LogAttrs(context.Background(), slog.LevelWarn, fmt.Sprintf(format, args...), e.withSource(2)...)
}

// WarnContext emits the entry at WARN level with context attributes.
func (e *Entry) WarnContext(ctx context.Context, msg string) {
	e.logger.slog.LogAttrs(ctx, slog.LevelWarn, msg, e.withSourceAndCtx(2, ctx)...)
}

// WarnContextf emits a formatted entry at WARN level with context attributes.
func (e *Entry) WarnContextf(ctx context.Context, format string, args ...any) {
	e.logger.slog.LogAttrs(ctx, slog.LevelWarn, fmt.Sprintf(format, args...), e.withSourceAndCtx(2, ctx)...)
}

// Error emits the entry at ERROR level.
func (e *Entry) Error(msg string) {
	e.logger.slog.LogAttrs(context.Background(), slog.LevelError, msg, e.withSource(2)...)
}

// Errorf emits the entry at ERROR level with a formatted message.
func (e *Entry) Errorf(format string, args ...any) {
	e.logger.slog.LogAttrs(context.Background(), slog.LevelError, fmt.Sprintf(format, args...), e.withSource(2)...)
}

// ErrorContext emits the entry at ERROR level with context attributes.
func (e *Entry) ErrorContext(ctx context.Context, msg string) {
	e.logger.slog.LogAttrs(ctx, slog.LevelError, msg, e.withSourceAndCtx(2, ctx)...)
}

// ErrorContextf emits a formatted entry at ERROR level with context attributes.
func (e *Entry) ErrorContextf(ctx context.Context, format string, args ...any) {
	e.logger.slog.LogAttrs(ctx, slog.LevelError, fmt.Sprintf(format, args...), e.withSourceAndCtx(2, ctx)...)
}

// Fatal emits the entry at FATAL level and then panics.
func (e *Entry) Fatal(msg string) {
	e.logger.slog.LogAttrs(context.Background(), fatalLevel, msg, e.withSource(2)...)
	panic(msg)
}

// Fatalf emits a formatted entry at FATAL level and then panics.
func (e *Entry) Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	e.logger.slog.LogAttrs(context.Background(), fatalLevel, msg, e.withSource(2)...)
	panic(msg)
}

// FatalContext emits the entry at FATAL level with context attributes, then panics.
func (e *Entry) FatalContext(ctx context.Context, msg string) {
	e.logger.slog.LogAttrs(ctx, fatalLevel, msg, e.withSourceAndCtx(2, ctx)...)
	panic(msg)
}

// FatalContextf emits a formatted entry at FATAL level with context attributes,
// then panics.
func (e *Entry) FatalContextf(ctx context.Context, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	e.logger.slog.LogAttrs(ctx, fatalLevel, msg, e.withSourceAndCtx(2, ctx)...)
	panic(msg)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// withSource returns the entry's attrs prepended with source/function attrs
// resolved at the given call stack depth. skip=2 from a public terminal
// method lands on the user's call site.
func (e *Entry) withSource(skip int) []slog.Attr {
	src := sourceAttrs(skip + 1) // +1 for withSource itself
	out := make([]slog.Attr, 0, len(src)+len(e.attrs))
	out = append(out, src...)
	out = append(out, e.attrs...)
	return out
}

// withSourceAndCtx is like withSource but also appends context attributes.
func (e *Entry) withSourceAndCtx(skip int, ctx context.Context) []slog.Attr {
	src := sourceAttrs(skip + 1) // +1 for withSourceAndCtx itself
	ctxAttrs := attrsFromContext(ctx)
	out := make([]slog.Attr, 0, len(src)+len(e.attrs)+len(ctxAttrs))
	out = append(out, src...)
	out = append(out, e.attrs...)
	out = append(out, ctxAttrs...)
	return out
}

// structToGroup wraps a struct's exported fields in a slog.Group keyed by
// the struct's type name (lowercase). Used for JSON format so each struct
// appears as a nested object rather than flat top-level fields.
func structToGroup(v any) (slog.Attr, bool) {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return slog.Attr{}, false
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return slog.Attr{}, false
	}

	rt := rv.Type()
	key := strings.ToLower(rt.Name())
	if key == "" {
		key = "struct" // anonymous struct fallback
	}

	fields := structToAttrs(v)
	any := make([]any, len(fields))
	for i, f := range fields {
		any[i] = f
	}
	return slog.Group(key, any...), true
}

// slog.Attr values. If a field has a `slog:"name"` tag that name is used as
// the key; otherwise the field name is used verbatim.
func structToAttrs(v any) []slog.Attr {
	rv := reflect.ValueOf(v)
	// Dereference pointer(s).
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}

	rt := rv.Type()
	attrs := make([]slog.Attr, 0, rt.NumField())

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		key := field.Name
		if tag, ok := field.Tag.Lookup("slog"); ok && tag != "" && tag != "-" {
			key = tag
		}

		attrs = append(attrs, slog.Any(key, rv.Field(i).Interface()))
	}
	return attrs
}
