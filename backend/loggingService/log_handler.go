package loggingService

import "github.com/rs/zerolog"

type (
	ILoggerService interface {
		GetLogger() *LoggerService
		LogInfo(message string, args ...interface{})
		LogError(err error, message string, args ...interface{})
	}

	LoggerService struct {
		Logger zerolog.Logger
	}
)

func (l *LoggerService) GetLogger() *LoggerService {
	return l
}

func (l *LoggerService) LogInfo(message string, args ...interface{}) {
	if len(args) > 0 {
		l.Logger.Info().Msgf(message, args...)
		return
	}

	l.Logger.Info().Msg(message)
}

func (l *LoggerService) LogError(err error, message string, args ...interface{}) {
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
