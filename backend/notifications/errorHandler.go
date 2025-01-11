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

func PanicRecoverLogger(logger zerolog.Logger) {
	if r := recover(); r != nil {
		logger.Error().Interface("recover", r).Msg("Panic occurred")
	}
}
