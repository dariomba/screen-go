package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type contextKey string

const loggerKey contextKey = "logger"

type Config struct {
	Level  string // "debug", "info", "warn", "error"
	Pretty bool
}

// Init initializes the global logger with the given config.
func Init(cfg Config) {
	// Parse log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339

	var output io.Writer = os.Stdout
	if cfg.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			NoColor:    false, // Enable colors in development
		}
	}

	// Set global logger with timestamp and caller info
	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()
}

// Ctx returns logger from context, or global logger if not found.
// Usage:
//
//	logger.Ctx(ctx).Info().Msg("Processing request")
func Ctx(ctx context.Context) *zerolog.Logger {
	if ctx != nil {
		if logger, ok := ctx.Value(loggerKey).(*zerolog.Logger); ok {
			return logger
		}
	}
	return &log.Logger
}

// WithRequestID adds request ID to context logger.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	logger := Ctx(ctx).With().
		Str("request_id", requestID).
		Logger()
	return context.WithValue(ctx, loggerKey, &logger)
}

// WithJobID adds job ID to context logger.
func WithJobID(ctx context.Context, jobID string) context.Context {
	logger := Ctx(ctx).With().
		Str("job_id", jobID).
		Logger()
	return context.WithValue(ctx, loggerKey, &logger)
}

// Info returns a new info-level event from the global logger
func Info() *zerolog.Event {
	return log.Info()
}

// Debug returns a new debug-level event from the global logger
func Debug() *zerolog.Event {
	return log.Debug()
}

// Warn returns a new warn-level event from the global logger
func Warn() *zerolog.Event {
	return log.Warn()
}

// Error returns a new error-level event from the global logger
func Error() *zerolog.Event {
	return log.Error()
}

// Fatal returns a new fatal-level event from the global logger
// Fatal logs call os.Exit(1) after writing the message
func Fatal() *zerolog.Event {
	return log.Fatal()
}

// Panic returns a new panic-level event from the global logger
// Panic logs call panic() after writing the message
func Panic() *zerolog.Event {
	return log.Panic()
}
