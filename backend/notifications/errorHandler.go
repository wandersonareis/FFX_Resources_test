package notifications

import (
	"fmt"

	"github.com/rs/zerolog"
)

func PanicRecover(errChan chan error, logger zerolog.Logger) {
	if r := recover(); r != nil {
		logger.Error().Interface("recover", r).Msg("Panic occurred")
		errChan <- fmt.Errorf("panic occurred")
	}
}

func ProcessError(errChan chan error, logger zerolog.Logger) {
	for {
		select {
		case err := <-errChan:
			if err != nil {
				logger.Error().Err(err).Msg("Error occurred")
			}
		}
	}
}
