package mt2

import (
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/internal/mt2/internal"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/loggingService"
	"fmt"
)

type (
	IKrnlCompressor interface {
		Compress(source interfaces.ISource, destination locations.IDestination) error
	}

	KrnlCompressor struct {
		KernelEncoder internal.IKrnlEncoder
		Log           loggingService.ILoggerService
	}
)

func NewKrnlCompressor(logger loggingService.ILoggerService) *KrnlCompressor {
	return &KrnlCompressor{
		KernelEncoder: internal.NewKrnlEncoder(),
		Log:           logger,
	}
}

func (k *KrnlCompressor) Compress(source interfaces.ISource, destination locations.IDestination) error {
	if err := destination.Import().ProvideTargetPath(); err != nil {
		outputPath := destination.Import().GetTargetPath()

		return fmt.Errorf("error providing import path: %s | error: %w", outputPath, err)
	}

	textEncoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextKrnlEncoding()
	defer textEncoding.Dispose()

	if err := k.KernelEncoder.Encoder(source, destination, textEncoding); err != nil {
		k.Log.Error(err, "Error compressing kernel file")
		return fmt.Errorf("error compressing kernel file: %s", err)
	}

	return nil
}
