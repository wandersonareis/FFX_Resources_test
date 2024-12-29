package notifications

import (
	"ffxresources/backend/logger"
	"fmt"

	"github.com/rs/zerolog"
)

func PanicRecover(errChan chan error, logger zerolog.Logger) {
	if r := recover(); r != nil {
		logger.Error().Interface("recover", r).Msg("Panic occurred")
		errChan <- fmt.Errorf("panic occurred")
	}
}

func PanicRecoverLogger(logger zerolog.Logger) {
	if r := recover(); r != nil {
		logger.Error().Interface("recover", r).Msg("Panic occurred")
	}
}

func ProcessError(errChan chan error, logHandler logger.ILoggerHandler) {
	for {
		select {
		case err := <-errChan:
			if err != nil {
				logHandler.LogError(err, "Error occurred")
			}
		}
	}
}
