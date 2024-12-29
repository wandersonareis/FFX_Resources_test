package logger

import "github.com/rs/zerolog"

type (
	ILoggerHandler interface {
		GetLogger() *LogHandler
		LogInfo(message string, args ...interface{})
		LogError(err error, message string, args ...interface{})
	}

	LogHandler struct {
		Logger zerolog.Logger
	}
)

func (l *LogHandler) GetLogger() *LogHandler {
	return l
}

func (l *LogHandler) LogInfo(message string, args ...interface{}) {
	if len(args) > 0 {
		l.Logger.Info().Msgf(message, args...)
		return
	}

	l.Logger.Info().Msg(message)
}

func (l *LogHandler) LogError(err error, message string, args ...interface{}) {
	if err != nil {
		if len(args) > 0 {
			l.Logger.Error().Err(err).Msgf(message, args...)
			return
		}

		l.Logger.Error().Err(err).Msg(message)
		return
	}

	if len(args) > 0 {
		l.Logger.Error().Msgf(message, args...)
		return
	}

	l.Logger.Error().Msg(message)
}
