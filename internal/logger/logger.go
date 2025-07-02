package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// Logger is the global logger instance
var Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger().Level(zerolog.TraceLevel)

// Info logs an info message
func Info(msg string) {
	Logger.Info().Msg(msg)
}

// Infof logs an info message with fields
func Infof(msg string, fields map[string]interface{}) {
	event := Logger.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Warn logs a warning message
func Warn(msg string) {
	Logger.Warn().Msg(msg)
}

// Warnf logs a warning message with fields
func Warnf(msg string, fields map[string]interface{}) {
	event := Logger.Warn()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Error logs an error message
func Error(msg string) {
	Logger.Error().Msg(msg)
}

// Errorf logs an error message with fields
func Errorf(msg string, fields map[string]interface{}) {
	event := Logger.Error()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Debug logs a debug message
func Debug(msg string) {
	Logger.Debug().Msg(msg)
}

// Debugf logs a debug message with fields
func Debugf(msg string, fields map[string]interface{}) {
	event := Logger.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string) {
	Logger.Fatal().Msg(msg)
}

// Fatalf logs a fatal message with fields and exits
func Fatalf(msg string, fields map[string]interface{}) {
	event := Logger.Fatal()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// SetLevel sets the log level
func SetLevel(level zerolog.Level) {
	Logger = Logger.Level(level)
}

// WithContext returns a logger with additional context
func WithContext(fields map[string]interface{}) zerolog.Logger {
	logger := Logger
	for k, v := range fields {
		logger = logger.With().Interface(k, v).Logger()
	}
	return logger
}
