package mt2

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/internal/mt2/internal"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
)

type (
	IKrnlExtractor interface {
		Extract(source interfaces.ISource, destination locations.IDestination) error
	}

	krnlExtractor struct {
		decoder internal.IKrnlDecoder
		log     logger.ILoggerHandler
	}
)

func newKrnlExtractor(logger logger.ILoggerHandler) *krnlExtractor {
	return &krnlExtractor{
		decoder: internal.NewKrnlDecoder(),
		log: logger,
	}
}

func (k *krnlExtractor) Extract(source interfaces.ISource, destination locations.IDestination) error {
	if err := destination.Extract().Get().ProvideTargetDirectory(); err != nil {
		return fmt.Errorf("failed to provide target directory: %s", err)
	}
	
	if err := k.decoder.Decoder(source, destination); err != nil {
		k.log.LogError(err, "Error decoding kernel file: %s", source.Get().Name)

		return fmt.Errorf("failed to decode kernel file: %s", source.Get().Name)
	}

	return nil
}
