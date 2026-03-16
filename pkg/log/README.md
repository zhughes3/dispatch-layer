# log

A structured logging package built on top of Go's standard [`log/slog`](https://pkg.go.dev/log/slog) library.  
It adds **context propagation**, **fatal logging with panic**, and a fluent **builder pattern** on top of the standard handler ecosystem.


---

## Quick Start

```go
log := log.New(log.Config{
    Format: log.FormatJSON,
    Level:  slog.LevelDebug,
})

log.Info("server started", "port", 8080)
log.With("component", "db").WithError(err).Error("query failed")
```

---

## Configuration

`log.New` accepts a `Config` struct:

| Field       | Type         | Default          | Description                                      |
|-------------|--------------|------------------|--------------------------------------------------|
| `Format`    | `Format`     | `FormatText`     | `"text"` or `"json"`                             |
| `Level`     | `slog.Level` | `slog.LevelInfo` | Minimum level emitted                            |
| `Output`    | `io.Writer`  | `os.Stderr`      | Destination for log output                       |

The log level can be changed at runtime via `logger.SetLevel(slog.LevelDebug)`.

---

## Log Levels

Each level exposes four variants:

| Variant           | Signature                                              |
|-------------------|--------------------------------------------------------|
| Plain             | `Debug(msg string, args ...any)`                       |
| Formatted         | `Debugf(format string, args ...any)`                   |
| Context-aware     | `DebugContext(ctx context.Context, msg string, args ...any)` |
| Context-formatted | `DebugContextf(ctx context.Context, format string, args ...any)` |

Available levels: **Debug · Info · Warn · Error · Fatal**

### Fatal

`Fatal*` functions write the log entry **then call `panic(msg)`**.  
The log line is guaranteed to appear even if the caller uses `recover()`.

```go
log.FatalContext(ctx, "startup check failed")
// → logs the message with all context attrs, then panics
```

---

## Context Propagation

Attach key/value pairs to a `context.Context` once (e.g. in middleware) and
they are automatically included in every context-aware log call downstream.

```go
// In HTTP middleware:
ctx = log.ContextWith(ctx, "trace_id", traceID, "user_id", userID)

// Deep in a handler or service:
log.InfoContext(ctx, "processed order")
// → time=... level=INFO msg="processed order" trace_id=abc user_id=42

// Multiple ContextWith calls accumulate (they do not overwrite):
ctx = log.ContextWith(ctx, "span_id", spanID)
```

---

## Builder Pattern

All `With*` methods return a new `*Entry` — the original logger is never mutated.

```go
log.With("key", value)                 // single key/value
log.WithStructs(myStruct, otherStruct) // all exported struct fields
log.WithError(err)                     // adds "error" key
log.WithDur(elapsed)                   // adds "duration" key
```

### Chaining

```go
log.
    With("component", "payments").
    WithError(err).
    WithDur(elapsed).
    ErrorContext(ctx, "charge failed")
```

### WithStructs & slog tags

`WithStructs` uses reflection to extract exported fields. Add a `slog:"key_name"` struct tag to control the output key name. Use `slog:"-"` to skip a field.

```go
type Order struct {
    ID       int     `slog:"order_id"`
    Amount   float64 `slog:"amount"`
    internal string  // unexported — always skipped
}

log.WithStructs(order).Info("order placed")
// → ... order_id=7 amount=49.99
```

---

## Package-Level Default Logger

A default `*Logger` is available through package-level functions so you can
use `log` like the standard library:

```go
log.Info("ready")
log.With("env", "prod").Warn("high memory")

// Replace the default with your configured logger:
log.SetDefault(myLogger)
```

---

## Output Examples

**Text**
```
time=2024-01-15T10:30:00.000Z level=INFO msg="request complete" method=POST path=/orders duration=134ms status_code=201
```

**JSON**
```json
{"time":"2024-01-15T10:30:00.000Z","level":"INFO","msg":"request complete","method":"POST","path":"/orders","duration":"134ms","status_code":201}
```

---

## Running Tests

```bash
go test ./...
```