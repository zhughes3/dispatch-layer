package log

import (
	"context"
	"fmt"
	"log/slog"
)

// ---------------------------------------------------------------------------
// Debug
// ---------------------------------------------------------------------------

// Debug logs a message at DEBUG level.
func (l *Logger) Debug(msg string, args ...any) {
	l.slog.LogAttrs(context.Background(), slog.LevelDebug, msg, withSource(2, args)...)
}

// Debugf logs a formatted message at DEBUG level.
func (l *Logger) Debugf(format string, args ...any) {
	l.slog.LogAttrs(context.Background(), slog.LevelDebug, fmt.Sprintf(format, args...), sourceAttrs(2)...)
}

// DebugContext logs a message at DEBUG level, automatically including any
// key/value attributes stored in the context via ContextWith.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	ctxArgs := argsFromContext(ctx)
	l.slog.LogAttrs(ctx, slog.LevelDebug, msg, withSource(2, mergeArgs(ctxArgs, args))...)
}

// DebugContextf logs a formatted message at DEBUG level with context attributes.
func (l *Logger) DebugContextf(ctx context.Context, format string, args ...any) {
	ctxArgs := argsFromContext(ctx)
	l.slog.LogAttrs(ctx, slog.LevelDebug, fmt.Sprintf(format, args...), append(sourceAttrs(2), toAttrs(ctxArgs)...)...)
}

// ---------------------------------------------------------------------------
// Info
// ---------------------------------------------------------------------------

// Info logs a message at INFO level.
func (l *Logger) Info(msg string, args ...any) {
	l.slog.LogAttrs(context.Background(), slog.LevelInfo, msg, withSource(2, args)...)
}

// Infof logs a formatted message at INFO level.
func (l *Logger) Infof(format string, args ...any) {
	l.slog.LogAttrs(context.Background(), slog.LevelInfo, fmt.Sprintf(format, args...), sourceAttrs(2)...)
}

// InfoContext logs a message at INFO level with context attributes.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	ctxArgs := argsFromContext(ctx)
	l.slog.LogAttrs(ctx, slog.LevelInfo, msg, withSource(2, mergeArgs(ctxArgs, args))...)
}

// InfoContextf logs a formatted message at INFO level with context attributes.
func (l *Logger) InfoContextf(ctx context.Context, format string, args ...any) {
	ctxArgs := argsFromContext(ctx)
	l.slog.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf(format, args...), append(sourceAttrs(2), toAttrs(ctxArgs)...)...)
}

// ---------------------------------------------------------------------------
// Warn
// ---------------------------------------------------------------------------

// Warn logs a message at WARN level.
func (l *Logger) Warn(msg string, args ...any) {
	l.slog.LogAttrs(context.Background(), slog.LevelWarn, msg, withSource(2, args)...)
}

// Warnf logs a formatted message at WARN level.
func (l *Logger) Warnf(format string, args ...any) {
	l.slog.LogAttrs(context.Background(), slog.LevelWarn, fmt.Sprintf(format, args...), sourceAttrs(2)...)
}

// WarnContext logs a message at WARN level with context attributes.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	ctxArgs := argsFromContext(ctx)
	l.slog.LogAttrs(ctx, slog.LevelWarn, msg, withSource(2, mergeArgs(ctxArgs, args))...)
}

// WarnContextf logs a formatted message at WARN level with context attributes.
func (l *Logger) WarnContextf(ctx context.Context, format string, args ...any) {
	ctxArgs := argsFromContext(ctx)
	l.slog.LogAttrs(ctx, slog.LevelWarn, fmt.Sprintf(format, args...), append(sourceAttrs(2), toAttrs(ctxArgs)...)...)
}

// ---------------------------------------------------------------------------
// Error
// ---------------------------------------------------------------------------

// Error logs a message at ERROR level.
func (l *Logger) Error(msg string, args ...any) {
	l.slog.LogAttrs(context.Background(), slog.LevelError, msg, withSource(2, args)...)
}

// Errorf logs a formatted message at ERROR level.
func (l *Logger) Errorf(format string, args ...any) {
	l.slog.LogAttrs(context.Background(), slog.LevelError, fmt.Sprintf(format, args...), sourceAttrs(2)...)
}

// ErrorContext logs a message at ERROR level with context attributes.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	ctxArgs := argsFromContext(ctx)
	l.slog.LogAttrs(ctx, slog.LevelError, msg, withSource(2, mergeArgs(ctxArgs, args))...)
}

// ErrorContextf logs a formatted message at ERROR level with context attributes.
func (l *Logger) ErrorContextf(ctx context.Context, format string, args ...any) {
	ctxArgs := argsFromContext(ctx)
	l.slog.LogAttrs(ctx, slog.LevelError, fmt.Sprintf(format, args...), append(sourceAttrs(2), toAttrs(ctxArgs)...)...)
}

// ---------------------------------------------------------------------------
// Fatal  (logs then panics)
// ---------------------------------------------------------------------------

// fatalLevel is a custom level above ERROR used to mark fatal entries in the
// log output before the program panics.
const fatalLevel = slog.Level(12) // above slog.LevelError (8)

// Fatal logs a message at FATAL level and then panics with the message.
func (l *Logger) Fatal(msg string, args ...any) {
	l.slog.LogAttrs(context.Background(), fatalLevel, msg, withSource(2, args)...)
	panic(msg)
}

// Fatalf logs a formatted message at FATAL level and then panics.
func (l *Logger) Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.slog.LogAttrs(context.Background(), fatalLevel, msg, sourceAttrs(2)...)
	panic(msg)
}

// FatalContext logs a message at FATAL level with context attributes, then panics.
func (l *Logger) FatalContext(ctx context.Context, msg string, args ...any) {
	ctxArgs := argsFromContext(ctx)
	l.slog.LogAttrs(ctx, fatalLevel, msg, withSource(2, mergeArgs(ctxArgs, args))...)
	panic(msg)
}

// FatalContextf logs a formatted message at FATAL level with context attributes,
// then panics.
func (l *Logger) FatalContextf(ctx context.Context, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	ctxArgs := argsFromContext(ctx)
	l.slog.LogAttrs(ctx, fatalLevel, msg, append(sourceAttrs(2), toAttrs(ctxArgs)...)...)
	panic(msg)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// withSource prepends sourceAttrs(skip) to a []any args slice, converting
// them inline so LogAttrs receives a uniform []slog.Attr.
func withSource(skip int, args []any) []slog.Attr {
	src := sourceAttrs(skip + 1) // +1 because withSource itself is a frame
	out := make([]slog.Attr, 0, len(src)+len(args))
	out = append(out, src...)
	out = append(out, toAttrs(args)...)
	return out
}

// toAttrs converts a []any produced by argsFromContext (which stores
// slog.Attr values) into []slog.Attr. Non-Attr elements are wrapped with
// slog.Any using a synthetic key.
func toAttrs(args []any) []slog.Attr {
	if len(args) == 0 {
		return nil
	}
	out := make([]slog.Attr, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch v := args[i].(type) {
		case slog.Attr:
			out = append(out, v)
		default:
			// key/value pair
			if i+1 < len(args) {
				if key, ok := args[i].(string); ok {
					out = append(out, slog.Any(key, args[i+1]))
					i++
					continue
				}
			}
			out = append(out, slog.Any(fmt.Sprintf("arg%d", i), v))
		}
	}
	return out
}

// mergeArgs merges two []any slices, returning a single flat slice.
// Either argument may be nil.
func mergeArgs(a, b []any) []any {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	merged := make([]any, 0, len(a)+len(b))
	merged = append(merged, a...)
	merged = append(merged, b...)
	return merged
}
