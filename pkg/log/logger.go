// Package golog provides a structured logging package built on top of log/slog.
// It supports text and JSON output formats, configurable log levels, context-aware
// logging, and a fluent builder pattern for constructing log entries.
package log

import (
	"io"
	"log/slog"
	"os"
	"time"
)

// Format specifies the output format for log entries.
type Format string

const (
	// FormatText outputs logs as human-readable key=value pairs.
	FormatText Format = "text"
	// FormatJSON outputs logs as JSON objects.
	FormatJSON Format = "json"
)

// Config holds the configuration for creating a new Logger.
type Config struct {
	// Format specifies the log output format. Defaults to FormatText.
	Format Format

	// Level sets the minimum log level. Defaults to slog.LevelInfo.
	Level slog.Level

	// Output is the writer to send log output to. Defaults to os.Stderr.
	Output io.Writer
}

// Logger is the core logging type. It wraps slog.Logger and provides
// additional methods for context extraction, fatal logging, and
// builder-pattern log entries.
type Logger struct {
	slog   *slog.Logger
	level  *slog.LevelVar
	format Format
}

// New creates and returns a new Logger using the provided Config.
// Any zero-value fields in Config are filled with sensible defaults:
//   - Format defaults to FormatText
//   - Level defaults to slog.LevelInfo
//   - Output defaults to os.Stderr
func New(cfg Config) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}
	if cfg.Format == "" {
		cfg.Format = FormatText
	}

	level := &slog.LevelVar{}
	level.Set(cfg.Level)

	opts := &slog.HandlerOptions{
		AddSource: false, // source is injected manually via sourceAttrs()
		Level:     level,
	}

	var handler slog.Handler
	switch cfg.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(cfg.Output, opts)
	default:
		handler = slog.NewTextHandler(cfg.Output, opts)
	}

	return &Logger{
		slog:   slog.New(handler),
		level:  level,
		format: cfg.Format,
	}
}

// SetLevel dynamically updates the minimum log level.
func (l *Logger) SetLevel(level slog.Level) {
	l.level.Set(level)
}

// slogLogger returns the underlying *slog.Logger, used internally and
// for cases where callers need direct slog access.
func (l *Logger) slogLogger() *slog.Logger {
	return l.slog
}

// With starts a new builder entry with a single key/value attribute.
// The returned Entry can be chained with additional With* calls before
// calling a terminal log method (Debug, Info, Warn, Error, Fatal, etc.).
//
// Example:
//
//	logger.With("request_id", reqID).With("user", userID).Info(ctx, "request received")
func (l *Logger) With(key string, value any) *Entry {
	return newEntry(l).With(key, value)
}

// WithStructs starts a new builder entry by extracting all exported
// fields from the provided structs as key/value log attributes.
// Unexported fields are silently skipped.
//
// Example:
//
//	logger.WithStructs(req, user).Info(ctx, "processed")
func (l *Logger) WithStructs(args ...any) *Entry {
	return newEntry(l).WithStructs(args...)
}

// WithError starts a new builder entry with an "error" attribute.
func (l *Logger) WithError(err error) *Entry {
	return newEntry(l).WithError(err)
}

// WithDur starts a new builder entry with a "duration" attribute.
func (l *Logger) WithDur(dur time.Duration) *Entry {
	return newEntry(l).WithDur(dur)
}

// child returns a logger with additional persistent attributes attached.
// These attributes appear on every subsequent log entry from the returned logger.
func (l *Logger) child(attrs ...slog.Attr) *Logger {
	args := make([]any, len(attrs))
	for i, a := range attrs {
		args[i] = a
	}
	return &Logger{
		slog:   l.slog.With(args...),
		level:  l.level,
		format: l.format,
	}
}
