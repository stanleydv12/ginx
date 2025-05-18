package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
)

type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

type Logger struct {
	handler slog.Handler
}

// New creates a new logger with the given output and log level
func New(w io.Writer, level Level) *Logger {
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				shortFile := source.File[strings.LastIndex(source.File, "/")+1:]
				return slog.String("source", fmt.Sprintf("%s:%d", shortFile, source.Line))
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(w, opts)
	return &Logger{handler: handler}
}

// Default creates a new logger with default settings (stderr, info level)
func Default() *Logger {
	return New(os.Stderr, LevelInfo)
}

// With adds attributes to the logger
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{handler: l.handler.WithAttrs(l.argsToAttrs(args...))}
}

// Debug logs a debug message with optional key-value pairs
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(context.Background(), LevelDebug, msg, args...)
}

// Info logs an info message with optional key-value pairs
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(context.Background(), LevelInfo, msg, args...)
}

// Warn logs a warning message with optional key-value pairs
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(context.Background(), LevelWarn, msg, args...)
}

// Error logs an error message with optional key-value pairs
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(context.Background(), LevelError, msg, args...)
}

// log is the internal logging function that handles all log levels
func (l *Logger) log(ctx context.Context, level Level, msg string, args ...interface{}) {
	if !l.handler.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])

	if len(args) > 0 {
		r.AddAttrs(l.argsToAttrs(args...)...)
	}

	_ = l.handler.Handle(ctx, r)
}

// argsToAttrs converts key-value pairs to slog.Attr
func (l *Logger) argsToAttrs(args ...interface{}) []slog.Attr {
	if len(args) == 0 {
		return nil
	}

	// Handle key-value pairs
	if len(args)%2 != 0 {
		args = append(args, "(MISSING)")
	}

	attrs := make([]slog.Attr, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			key = fmt.Sprintf("(key %d)", i/2)
		}
		attrs = append(attrs, slog.Any(key, args[i+1]))
	}

	return attrs
}

// SetDefault sets the default logger to use this logger
func (l *Logger) SetDefault() {
	slog.SetDefault(slog.New(l.handler))
}