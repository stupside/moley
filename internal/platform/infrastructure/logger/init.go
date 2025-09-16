// Package logger provides structured logging for Moley.
package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func InitLogger(level zerolog.Level) {
	logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
		FormatCaller: func(i any) string {
			path := fmt.Sprintf("%v", i)
			dir := filepath.Base(filepath.Dir(path))
			file := filepath.Base(path)
			short := fmt.Sprintf("%s/%s", dir, file)
			if len(short) > 25 {
				short = short[:25]
			}
			return fmt.Sprintf("%-30s", short)
		},
	}).Level(level).With().Timestamp().CallerWithSkipFrameCount(3).Logger()
}
