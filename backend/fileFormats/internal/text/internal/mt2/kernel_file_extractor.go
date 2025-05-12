package mt2

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/internal/mt2/internal"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/loggingService"
	"fmt"
)

type (
	IKrnlExtractor interface {
		Extract(source interfaces.ISource, destination locations.IDestination) error
	}

	krnlExtractor struct {
		decoder internal.IKrnlDecoder
		log     loggingService.ILoggerService
	}
)

func NewKrnlExtractor(logger loggingService.ILoggerService) *krnlExtractor {
	return &krnlExtractor{
		decoder: internal.NewKrnlDecoder(),
		log:     logger,
	}
}

func (k *krnlExtractor) Extract(source interfaces.ISource, destination locations.IDestination) error {
	if err := destination.Extract().ProvideTargetDirectory(); err != nil {
		return fmt.Errorf("failed to provide target directory: %s", err)
	}

	if err := k.decoder.Decoder(source, destination); err != nil {
		k.log.Error(err, "Error decoding kernel file: %s", source.GetName())

		return fmt.Errorf("failed to decode kernel file: %s", source.GetName())
	}

	return nil
}
