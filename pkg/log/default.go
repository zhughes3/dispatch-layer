package log

import (
	"context"
	"log/slog"
	"sync/atomic"
	"unsafe"
)

// defaultLogger is the package-level logger used by the top-level functions
// (Debug, Info, etc.). It is initialised to sensible defaults and can be
// replaced atomically via SetDefault.
var defaultLoggerPtr unsafe.Pointer

func init() {
	l := New(Config{
		Format: FormatText,
		Level:  slog.LevelInfo,
	})
	atomic.StorePointer(&defaultLoggerPtr, unsafe.Pointer(l))
}

// Default returns the package-level default logger.
func Default() *Logger {
	return (*Logger)(atomic.LoadPointer(&defaultLoggerPtr))
}

// SetDefault replaces the package-level default logger.
func SetDefault(l *Logger) {
	atomic.StorePointer(&defaultLoggerPtr, unsafe.Pointer(l))
}

// ---------------------------------------------------------------------------
// Package-level convenience functions (delegate to Default())
// ---------------------------------------------------------------------------

func Debug(msg string, args ...any)     { Default().Debug(msg, args...) }
func Debugf(format string, args ...any) { Default().Debugf(format, args...) }
func DebugContext(ctx context.Context, msg string, args ...any) {
	Default().DebugContext(ctx, msg, args...)
}
func DebugContextf(ctx context.Context, format string, args ...any) {
	Default().DebugContextf(ctx, format, args...)
}

func Info(msg string, args ...any)     { Default().Info(msg, args...) }
func Infof(format string, args ...any) { Default().Infof(format, args...) }
func InfoContext(ctx context.Context, msg string, args ...any) {
	Default().InfoContext(ctx, msg, args...)
}
func InfoContextf(ctx context.Context, format string, args ...any) {
	Default().InfoContextf(ctx, format, args...)
}

func Warn(msg string, args ...any)     { Default().Warn(msg, args...) }
func Warnf(format string, args ...any) { Default().Warnf(format, args...) }
func WarnContext(ctx context.Context, msg string, args ...any) {
	Default().WarnContext(ctx, msg, args...)
}
func WarnContextf(ctx context.Context, format string, args ...any) {
	Default().WarnContextf(ctx, format, args...)
}

func Error(msg string, args ...any)     { Default().Error(msg, args...) }
func Errorf(format string, args ...any) { Default().Errorf(format, args...) }
func ErrorContext(ctx context.Context, msg string, args ...any) {
	Default().ErrorContext(ctx, msg, args...)
}
func ErrorContextf(ctx context.Context, format string, args ...any) {
	Default().ErrorContextf(ctx, format, args...)
}

func Fatal(msg string, args ...any)     { Default().Fatal(msg, args...) }
func Fatalf(format string, args ...any) { Default().Fatalf(format, args...) }
func FatalContext(ctx context.Context, msg string, args ...any) {
	Default().FatalContext(ctx, msg, args...)
}
func FatalContextf(ctx context.Context, format string, args ...any) {
	Default().FatalContextf(ctx, format, args...)
}

func With(key string, value any) *Entry { return Default().With(key, value) }
func WithStructs(args ...any) *Entry    { return Default().WithStructs(args...) }
func WithError(err error) *Entry        { return Default().WithError(err) }
