package mt2

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/mt2/internal"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
)

type (
	IKrnlCompressor interface {
		Compress(source interfaces.ISource, destination locations.IDestination) error
	}

	KrnlCompressor struct {
		encoder internal.IKrnlEncoder
		log     logger.ILoggerHandler
	}
)

func newKrnlCompressor() *KrnlCompressor {
	return &KrnlCompressor{
		encoder: internal.NewKrnlEncoder(),
		log: &logger.LogHandler{
			Logger: logger.Get().With().Str("module", "kernel_file_compressor").Logger(),
		},
	}
}

func (k *KrnlCompressor) Compress(source interfaces.ISource, destination locations.IDestination) error {
	if err := k.encoder.Encoder(source, destination); err != nil {
		k.log.LogError(err, "Error compressing kernel file: %s", destination.Translate().Get().GetTargetFile())

		return fmt.Errorf("failed to compress kernel file: %s", source.Get().Name)
	}

	return nil
}
