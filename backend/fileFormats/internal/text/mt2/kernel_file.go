package mt2

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/lib/textVerifier"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
	"sync"
)

type kernelFile struct {
	source      interfaces.ISource
	destination locations.IDestination

	compressorOnce   sync.Once
	extractorOnce    sync.Once
	textVerifierOnce sync.Once

	compressor   IKrnlCompressor
	extractor    IKrnlExtractor
	textVerifier textVerifier.ITextVerifier

	log logger.ILoggerHandler
}

func NewKernel(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	//destination.InitializeLocations(source, formatters.NewTxtFormatter())

	return &kernelFile{
		source:      source,
		destination: destination,
		log: &logger.LogHandler{
			Logger: logger.Get().With().Str("module", "kernel_file").Logger(),
		},
	}
}

func (k *kernelFile) Source() interfaces.ISource {
	return k.source
}

func (k *kernelFile) Extract() error {
	k.extractorOnce.Do(func() {
		k.extractor = newKrnlExtractor()
	})

	k.textVerifierOnce.Do(func() {
		k.textVerifier = textVerifier.NewTextsVerify()
	})

	if !common.IsFileExists(k.source.Get().Path) {
		k.log.LogError(nil, "Kernel file not found: %s", k.source.Get().Name)

		return fmt.Errorf("kernel file not found: %s", k.source.Get().Name)
	}

	k.log.LogInfo("Extracting kernel file: %s", k.source.Get().Name)

	if err := k.extractor.Extract(k.source, k.destination); err != nil {
		k.log.LogError(err, "Error extracting kernel file: %s", k.source.Get().Name)

		return fmt.Errorf("failed to decode kernel file: %s", k.source.Get().Name)
	}

	k.log.LogInfo("Verifying extracted kernel file: %s", k.destination.Extract().Get().GetTargetFile())

	if err := k.textVerifier.Verify(k.source, k.destination, textVerifier.ExtractIntegrityCheck); err != nil {
		k.log.LogError(err, "Error verifying kernel file: %s", k.source.Get().Name)

		return fmt.Errorf("failed to verify kernel file: %s", k.source.Get().Name)
	}

	k.log.LogInfo("Kernel file extracted: %s", k.source.Get().Name)

	return nil
}

func (k *kernelFile) Compress() error {
	k.compressorOnce.Do(func() {
		k.compressor = newKrnlCompressor()
	})

	k.textVerifierOnce.Do(func() {
		k.textVerifier = textVerifier.NewTextsVerify()
	})

	if err := k.compressor.Compress(k.source, k.destination); err != nil {
		k.log.LogError(err, "Error compressing kernel file: %s", k.destination.Translate().Get().GetTargetFile())

		return fmt.Errorf("failed to compress kernel file: %s", k.source.Get().Name)
	}

	tmpSource, tmpDestination := k.createTemp(k.source, k.destination)
	defer tmpDestination.Extract().Get().Dispose()

	tmpFile := NewKernel(tmpSource, tmpDestination)

	if err := tmpFile.Extract(); err != nil {
		k.log.LogError(err, "Error decoding kernel file: %s", k.destination.Import().Get().GetTargetFile())

		return fmt.Errorf("failed to decode kernel file: %s", k.destination.Import().Get().GetTargetFile())
	}

	if err := k.textVerifier.Verify(tmpSource, tmpDestination, textVerifier.CompressIntegrityCheck); err != nil {
		k.log.LogError(err, "Error verifying kernel file: %s", k.source.Get().Name)

		return fmt.Errorf("failed to verify kernel file: %s", k.source.Get().Name)
	}

	k.log.LogInfo("Kernel file compressed: %s", k.destination.Import().Get().GetTargetFile())

	return nil
}

func (k *kernelFile) createTemp(source interfaces.ISource, destination locations.IDestination) (interfaces.ISource, locations.IDestination) {
	tmp := common.NewTempProvider("tmp", ".txt")

	tmpSource := source
	tmpDestination := destination

	tmpDestination.Extract().Get().SetTargetFile(tmp.TempFile)
	tmpDestination.Extract().Get().SetTargetPath(tmp.TempFilePath)

	tmpSource.Get().Path = destination.Import().Get().GetTargetFile()

	return tmpSource, tmpDestination
}
