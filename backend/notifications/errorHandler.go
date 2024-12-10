package notifications

import "github.com/rs/zerolog"

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
