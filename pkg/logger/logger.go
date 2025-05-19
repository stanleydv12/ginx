// Package logger provides a structured, leveled logging implementation
// built on top of the standard library's log/slog package.
// It is safe for concurrent use by multiple goroutines.
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	// defaultLogger is the singleton logger instance
	defaultLogger *Logger
	once          sync.Once
)

// Level is an alias for slog.Level for better documentation and usability
type Level = slog.Level

// Log levels
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Logger provides structured, leveled logging
// It is safe for concurrent use by multiple goroutines.
type Logger struct {
	mu      sync.RWMutex
	handler slog.Handler
}

// New creates a new logger with the given output and log level.
// The returned logger is safe for concurrent use by multiple goroutines.
func New(w io.Writer, level Level) *Logger {
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				// Get the relative path from the module root
				relPath, err := filepath.Rel(getModuleRoot(), source.File)
				if err != nil {
					relPath = source.File
				}
				return slog.String("source", fmt.Sprintf("%s:%d", relPath, source.Line))
			}
			return a
		},
		AddSource: true,
	}

	handler := slog.NewJSONHandler(w, opts)
	return &Logger{handler: handler}
}

// getModuleRoot finds the module root directory
func getModuleRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "."
}

// Default returns the singleton logger instance, initializing it if necessary
func Default() *Logger {
	once.Do(func() {
		defaultLogger = New(os.Stderr, LevelInfo)
	})
	return defaultLogger
}

// SetDefault sets up the global logger with default settings
func SetDefault() {
	slog.SetDefault(slog.New(Default().handler))
}

// Debug logs a debug message with optional key-value pairs
func Debug(msg string, args ...interface{}) {
	Default().logCaller(1, context.Background(), LevelDebug, msg, args...)
}

// DebugContext logs a debug message with context and optional key-value pairs
func DebugContext(ctx context.Context, msg string, args ...interface{}) {
	Default().logCaller(1, ctx, LevelDebug, msg, args...)
}

// Info logs an info message with optional key-value pairs
func Info(msg string, args ...interface{}) {
	Default().logCaller(1, context.Background(), LevelInfo, msg, args...)
}

// InfoContext logs an info message with context and optional key-value pairs
func InfoContext(ctx context.Context, msg string, args ...interface{}) {
	Default().logCaller(1, ctx, LevelInfo, msg, args...)
}

// Warn logs a warning message with optional key-value pairs
func Warn(msg string, args ...interface{}) {
	Default().logCaller(1, context.Background(), LevelWarn, msg, args...)
}

// WarnContext logs a warning message with context and optional key-value pairs
func WarnContext(ctx context.Context, msg string, args ...interface{}) {
	Default().logCaller(1, ctx, LevelWarn, msg, args...)
}

// Error logs an error message with optional key-value pairs
func Error(msg string, args ...interface{}) {
	Default().logCaller(1, context.Background(), LevelError, msg, args...)
}

// ErrorContext logs an error message with context and optional key-value pairs
func ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	Default().logCaller(1, ctx, LevelError, msg, args...)
}

// With adds attributes to the logger and returns a new logger instance.
// The new logger's handler is a copy of the original with the specified attributes.
func (l *Logger) With(args ...interface{}) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return &Logger{handler: l.handler.WithAttrs(l.argsToAttrs(args...))}
}

// Debug logs a debug message with optional key-value pairs
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.logCaller(1, context.Background(), LevelDebug, msg, args...)
}

// DebugContext logs a debug message with context and optional key-value pairs
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	l.logCaller(1, ctx, LevelDebug, msg, args...)
}

// Info logs an info message with optional key-value pairs
func (l *Logger) Info(msg string, args ...interface{}) {
	l.logCaller(1, context.Background(), LevelInfo, msg, args...)
}

// InfoContext logs an info message with context and optional key-value pairs
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	l.logCaller(1, ctx, LevelInfo, msg, args...)
}

// Warn logs a warning message with optional key-value pairs
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.logCaller(1, context.Background(), LevelWarn, msg, args...)
}

// WarnContext logs a warning message with context and optional key-value pairs
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	l.logCaller(1, ctx, LevelWarn, msg, args...)
}

// Error logs an error message with optional key-value pairs
func (l *Logger) Error(msg string, args ...interface{}) {
	l.logCaller(1, context.Background(), LevelError, msg, args...)
}

// ErrorContext logs an error message with context and optional key-value pairs
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	l.logCaller(1, ctx, LevelError, msg, args...)
}

// logCaller is the internal logging function that handles all log levels with caller info
func (l *Logger) logCaller(skip int, ctx context.Context, level Level, msg string, args ...interface{}) {
	l.mu.RLock()
	handler := l.handler
	l.mu.RUnlock()

	if !handler.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	n := runtime.Callers(skip+2, pcs[:]) // skip runtime.Callers and logCaller
	if n == 0 {
		// If we can't get the caller, log without source info
		r := slog.NewRecord(time.Now(), level, msg, 0)
		if len(args) > 0 {
			r.AddAttrs(l.argsToAttrs(args...)...)
		}
		_ = handler.Handle(ctx, r)
		return
	}

	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	if len(args) > 0 {
		r.AddAttrs(l.argsToAttrs(args...)...)
	}

	if err := handler.Handle(ctx, r); err != nil {
		// Fallback to stderr if logging fails
		_, _ = fmt.Fprintf(os.Stderr, "log: failed to handle log record: %v\n", err)
	}
}

// argsToAttrs converts key-value pairs to slog.Attr
func (l *Logger) argsToAttrs(args ...interface{}) []slog.Attr {
	if len(args) == 0 {
		return nil
	}

	attrs := make([]slog.Attr, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		if i+1 >= len(args) {
			break
		}

		key, ok := args[i].(string)
		if !ok {
			continue
		}

		attrs = append(attrs, slog.Any(key, args[i+1]))
	}

	return attrs
}
