package logutil

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

var (
	logger *slog.Logger
)

func init() {
	// Initialize structured logger with JSON handler
	opts := &slog.HandlerOptions{
		Level: getLogLevel(),
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger = slog.New(handler)

	// Set as default logger for standard log package integration
	slog.SetDefault(logger)
}

func getLogLevel() slog.Level {
	lvl := os.Getenv("LOG_LEVEL")
	switch lvl {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // Default to Info if unset or invalid
	}
}

// Debug logs at Debug level
func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

// Debugf logs at Debug level (formatted message)
func Debugf(format string, args ...any) {
	logger.Debug(fmt.Sprintf(format, args...))
}

// Info logs at Info level
func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

// Infof logs at Info level (formatted message)
func Infof(format string, args ...any) {
	logger.Info(fmt.Sprintf(format, args...))
}

// Warn logs at Warn level
func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

// Warnf logs at Warn level (formatted message)
func Warnf(format string, args ...any) {
	logger.Warn(fmt.Sprintf(format, args...))
}

// Error logs at Error level
func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

// Errorf logs at Error level (formatted message)
func Errorf(format string, args ...any) {
	logger.Error(fmt.Sprintf(format, args...))
}

// Fatal logs at Error level and exits
func Fatal(msg string, args ...any) {
	logger.Error(msg, args...)
	os.Exit(1)
}

// Fatalf logs at Error level (formatted message) and exits
func Fatalf(format string, args ...any) {
	logger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// With creates a child logger with the given args
func With(args ...any) *slog.Logger {
	return logger.With(args...)
}

// FromContext returns a logger from context (if set) or default
// Placeholder for future context-based logging
func FromContext(ctx context.Context) *slog.Logger {
	return logger
}
