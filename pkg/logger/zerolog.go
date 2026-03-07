package logger

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

const (
	skipFrameCount = 3
	maxPathSize    = 5
)

var globalLogger zerolog.Logger //nolint:gochecknoglobals

func InitZerolog() {
	zerolog.SetGlobalLevel(getLogLevel(strings.ToLower(os.Getenv("LOG_LEVEL"))))
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		paths := strings.Split(file, "/")
		if len(paths) > maxPathSize {
			file = strings.Join(paths[maxPathSize:], "/")
		}

		return file + ":" + strconv.Itoa(line)
	}

	globalLogger = zerolog.New(os.Stdout).With().Timestamp().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + skipFrameCount).Logger()
}

func WithEmpty() *zerolog.Event {
	return globalLogger.Log()
}

func Debug(message interface{}, args ...interface{}) {
	msg("debug", message, args...)
}

func Info(message string, args ...interface{}) {
	log("info", message, args...)
}

func Warn(message string, args ...interface{}) {
	log("warn", message, args...)
}

func Error(message interface{}, args ...interface{}) {
	if globalLogger.GetLevel() == zerolog.DebugLevel {
		Debug(message, args...)
	}

	msg("error", message, args...)
}

func Fatal(message interface{}, args ...interface{}) {
	msg("fatal", message, args...)

	os.Exit(1)
}

func log(level, message string, args ...interface{}) {
	l := getLogLevel(level)
	event := globalLogger.WithLevel(l)

	if len(args) == 0 {
		event.Msg(message)
	} else {
		event.Msgf(message, args...)
	}
}

func msg(level string, message interface{}, args ...interface{}) {
	switch msg := message.(type) {
	case error:
		log(level, msg.Error(), args...)
	case string:
		log(level, msg, args...)
	default:
		log(level, fmt.Sprintf("%s message %v has unknown type %v", level, message, msg), args...)
	}
}

func getLogLevel(level string) zerolog.Level {
	switch level {
	case "error":
		return zerolog.ErrorLevel
	case "warn":
		return zerolog.WarnLevel
	case "info":
		return zerolog.InfoLevel
	case "debug":
		return zerolog.DebugLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}
