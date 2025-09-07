// Package logger provides a structured logging interface for Moley.
package logger

import (
	"github.com/rs/zerolog"
)

func addFields(event *zerolog.Event, fields map[string]any) {
	for k, v := range fields {
		event.Interface(k, v)
	}
}

func Info(msg string) {
	logger.Info().Msg(msg)
}

func Infof(msg string, fields map[string]any) {
	event := logger.Info()
	addFields(event, fields)
	event.Msg(msg)
}

func Debug(msg string) {
	logger.Debug().Msg(msg)
}

func Debugf(msg string, fields map[string]any) {
	event := logger.Debug()
	addFields(event, fields)
	event.Msg(msg)
}

func Warn(msg string) {
	logger.Warn().Msg(msg)
}

func Warnf(msg string, fields map[string]any) {
	event := logger.Warn()
	addFields(event, fields)
	event.Msg(msg)
}

func Error(msg string) {
	logger.Error().Msg(msg)
}

func Errorf(msg string, fields map[string]any) {
	event := logger.Error()
	addFields(event, fields)
	event.Msg(msg)
}

func LogError(err error, msg string) {
	if err == nil {
		return
	}
	logger.Error().Err(err).Msg(msg)
}

func LogErrorf(err error, msg string, fields map[string]any) {
	if err == nil {
		return
	}
	event := logger.Error().Err(err)
	addFields(event, fields)
	event.Msg(msg)
}

func Fatal(msg string) {
	logger.Fatal().Msg(msg)
}

func Fatalf(msg string, fields map[string]any) {
	event := logger.Fatal()
	addFields(event, fields)
	event.Msg(msg)
}
