package log

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

const (
	TRACE = zerolog.TraceLevel
	DEBUG = zerolog.DebugLevel
	INFO  = zerolog.InfoLevel
	WARN  = zerolog.WarnLevel
	ERROR = zerolog.ErrorLevel
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = time.RFC3339

	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	logger := log.Output(writer)
	log.Logger = logger
}

func UpdateLevel(level zerolog.Level) {
	zerolog.SetGlobalLevel(level)
}

func Debug(args ...interface{}) {
	log.Debug().Msg(fmt.Sprint(args...))
}

func Debugf(t string, args ...interface{}) {
	log.Debug().Msgf(t, args...)
}

func Fatal(args ...interface{}) {
	log.Fatal().Msg(fmt.Sprint(args...))
}

func Fatalf(t string, args ...interface{}) {
	log.Fatal().Msgf(t, args...)
}

func Error(args ...interface{}) {
	log.Error().Msg(fmt.Sprint(args...))
}

func Errorf(t string, args ...interface{}) {
	log.Error().Msgf(t, args...)
}

func Info(args ...interface{}) {
	log.Info().Msg(fmt.Sprint(args...))
}

func Infof(t string, args ...interface{}) {
	log.Info().Msgf(t, args...)
}

func Warn(args ...interface{}) {
	log.Warn().Msg(fmt.Sprint(args...))
}

func Warnf(t string, args ...interface{}) {
	log.Warn().Msgf(t, args...)
}

func Println(args ...interface{}) {
	log.Log().Msg(fmt.Sprintln(args...))
}

func Print(args ...interface{}) {
	log.Log().Msg(fmt.Sprint(args...))
}

func Printf(t string, args ...interface{}) {
	log.Log().Msgf(t, args...)
}

func Sync() {
}
