package log

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

// sourceAttrs resolves the caller's file:line and function name by walking
// `skip` frames up the call stack, then returns them as two slog.Attr values
// ready to append to any log record.
//
// skip=2 is correct for direct Logger methods  (user → Logger.Method → here).
// skip=3 is correct for Entry terminal methods (user → Entry.Method → emit → here).
func sourceAttrs(skip int) []slog.Attr {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return []slog.Attr{
			slog.String("source", "N/A"),
			slog.String("function", "N/A"),
		}
	}
	frames := runtime.CallersFrames([]uintptr{pc})
	frame, _ := frames.Next()
	return []slog.Attr{
		slog.String("source", formatSource(frame)),
		slog.String("function", formatFunction(frame)),
	}
}

func formatSource(frame runtime.Frame) string {
	fileIndex := strings.LastIndex(frame.File, "/")
	filename := frame.File[fileIndex+1:]
	return fmt.Sprintf("%s:%d", filename, frame.Line)
}

func formatFunction(frame runtime.Frame) string {
	fnIndex := strings.LastIndex(frame.Function, "/")
	return frame.Function[fnIndex+1:]
}
