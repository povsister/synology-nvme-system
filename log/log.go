package log

import "github.com/rs/zerolog"

func Trace() *zerolog.Event {
	return defaultLogger.Trace()
}

func Debug() *zerolog.Event {
	return defaultLogger.Debug()
}

func Info() *zerolog.Event {
	return defaultLogger.Info()
}

func Warn() *zerolog.Event {
	return defaultLogger.Warn()
}

func Error(errs ...error) *zerolog.Event {
	return defaultLogger.Error().Errs("errors", errs)
}
