package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

// logger is the global logger instance
var logger = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: "2006-01-02 15:04:05",
	FormatCaller: func(i interface{}) string {
		path := fmt.Sprintf("%v", i)

		dir := filepath.Base(filepath.Dir(path)) // e.g. cmd
		file := filepath.Base(path)              // e.g. config.go:19

		short := fmt.Sprintf("%s/%s", dir, file) // e.g. cmd/config.go:19

		// Truncate to 20 characters, left-aligned
		// This ensures the output is consistent and fits within a fixed width
		return fmt.Sprintf("%-30s", short[:min(len(short), 25)])
	},
}).Level(zerolog.InfoLevel).With().Timestamp().CallerWithSkipFrameCount(3).Logger()

// Helper function to add fields to a log event
func addFields(event *zerolog.Event, fields map[string]interface{}) {
	for k, v := range fields {
		event = event.Interface(k, v)
	}
}

// Info logs an info message
func Info(msg string) {
	logger.Info().Msg(msg)
}

// Infof logs an info message with fields
func Infof(msg string, fields map[string]interface{}) {
	event := logger.Info()
	addFields(event, fields)
	event.Msg(msg)
}

// Debug logs a debug message
func Debug(msg string) {
	logger.Debug().Msg(msg)
}

// Debugf logs a debug message with fields
func Debugf(msg string, fields map[string]interface{}) {
	event := logger.Debug()
	addFields(event, fields)
	event.Msg(msg)
}

// Warn logs a warning message
func Warn(msg string) {
	logger.Warn().Msg(msg)
}

// Warnf logs a warning message with fields
func Warnf(msg string, fields map[string]interface{}) {
	event := logger.Warn()
	addFields(event, fields)
	event.Msg(msg)
}

// Error logs an error message
func Error(msg string) {
	logger.Error().Msg(msg)
}

// Errorf logs an error message with fields
func Errorf(msg string, fields map[string]interface{}) {
	event := logger.Error()
	addFields(event, fields)
	event.Msg(msg)
}

// LogError logs an error with details
// This should be the preferred method for logging errors in the application
func LogError(err error, msg string) {
	if err == nil {
		return
	}
	logger.Error().Err(err).Msg(msg)
}

// LogErrorf logs an error with details and fields
func LogErrorf(err error, msg string, fields map[string]interface{}) {
	if err == nil {
		return
	}
	event := logger.Error().Err(err)
	addFields(event, fields)
	event.Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string) {
	logger.Fatal().Msg(msg)
}

// Fatalf logs a fatal message with fields and exits
func Fatalf(msg string, fields map[string]interface{}) {
	event := logger.Fatal()
	addFields(event, fields)
	event.Msg(msg)
}

// LogFatal logs a fatal error with details and exits
func LogFatal(err error, msg string) {
	if err == nil {
		return
	}
	logger.Fatal().Err(err).Msg(msg)
}

// SetLevel sets the log level
func SetLevel(level zerolog.Level) {
	logger = logger.Level(level)
}

// WithContext returns a logger with additional context
// This is useful for adding persistent fields to log messages
// in a specific component or function chain
func WithContext(fields map[string]interface{}) zerolog.Logger {
	logger := logger
	for k, v := range fields {
		logger = logger.With().Interface(k, v).Logger()
	}
	return logger
}
