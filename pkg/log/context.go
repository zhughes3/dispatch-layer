package log

import (
	"context"
	"log/slog"
)

// contextKey is the unexported type used to store log attributes in a context,
// preventing collisions with keys from other packages.
type contextKey struct{}

// contextAttrs is the value type stored under contextKey in a context.
type contextAttrs struct {
	attrs []slog.Attr
}

// ContextWith returns a new context with the given key/value pairs attached
// as log attributes. These will be automatically extracted and included in
// any log call that accepts a context (e.g. DebugContext, InfoContext).
//
// Multiple calls accumulate attributes; later calls append to earlier ones.
//
// Example:
//
//	ctx = golog.ContextWith(ctx, "request_id", reqID, "user_id", userID)
//	logger.InfoContext(ctx, "handling request")
func ContextWith(ctx context.Context, args ...any) context.Context {
	existing := attrsFromContext(ctx)

	var newAttrs []slog.Attr
	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		newAttrs = append(newAttrs, slog.Any(key, args[i+1]))
	}

	merged := make([]slog.Attr, 0, len(existing)+len(newAttrs))
	merged = append(merged, existing...)
	merged = append(merged, newAttrs...)

	return context.WithValue(ctx, contextKey{}, &contextAttrs{attrs: merged})
}

// ContextWithAttrs is like ContextWith but accepts already-formed slog.Attr values.
func ContextWithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	existing := attrsFromContext(ctx)
	merged := make([]slog.Attr, 0, len(existing)+len(attrs))
	merged = append(merged, existing...)
	merged = append(merged, attrs...)
	return context.WithValue(ctx, contextKey{}, &contextAttrs{attrs: merged})
}

// attrsFromContext extracts log attributes stored in the context by ContextWith.
// Returns nil if no attributes are present.
func attrsFromContext(ctx context.Context) []slog.Attr {
	if ctx == nil {
		return nil
	}
	val, ok := ctx.Value(contextKey{}).(*contextAttrs)
	if !ok || val == nil {
		return nil
	}
	return val.attrs
}

// argsFromContext converts context-stored attributes into a flat []any slice
// suitable for passing to slog's Log/LogAttrs methods.
func argsFromContext(ctx context.Context) []any {
	attrs := attrsFromContext(ctx)
	if len(attrs) == 0 {
		return nil
	}
	args := make([]any, len(attrs))
	for i, a := range attrs {
		args[i] = a
	}
	return args
}
