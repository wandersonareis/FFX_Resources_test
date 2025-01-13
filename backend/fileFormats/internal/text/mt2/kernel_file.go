package mt2

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/lib/textVerifier"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
)

type kernelFile struct {
	source      interfaces.ISource
	destination locations.IDestination

	log logger.ILoggerHandler
}

func NewKernel(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
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
	if !common.IsFileExists(k.source.Get().Path) {
		k.log.LogError(nil, "Kernel file not found: %s", k.source.Get().Name)

		return fmt.Errorf("kernel file not found: %s", k.source.Get().Name)
	}

	k.log.LogInfo("Extracting kernel file: %s", k.source.Get().Name)

	extractorInstance := rentKrnlExtractor()
	defer returnKrnlExtractor(extractorInstance)
	fmt.Println("extractorInstance", extractorInstance)

	if err := extractorInstance.Extract(k.source, k.destination); err != nil {
		k.log.LogError(err, "Error extracting kernel file: %s", k.source.Get().Name)

		return fmt.Errorf("failed to decode kernel file: %s", k.source.Get().Name)
	}

	k.log.LogInfo("Verifying extracted kernel file: %s", k.destination.Extract().Get().GetTargetFile())

	verifierInstance := rentTextVerifier()
	defer returnTextVerifier(verifierInstance)
	fmt.Println("verifierInstance", verifierInstance)

	if err := verifierInstance.Verify(k.source, k.destination, textVerifier.ExtractIntegrityCheck); err != nil {
		k.log.LogError(err, "Error verifying kernel file: %s", k.source.Get().Name)

		return fmt.Errorf("failed to verify kernel file: %s", k.source.Get().Name)
	}

	k.log.LogInfo("Kernel file extracted: %s", k.source.Get().Name)

	return nil
}

func (k *kernelFile) Compress() error {
	compressorInstance := rentKrnlCompressor()
	defer returnKrnlCompressor(compressorInstance)

	if err := compressorInstance.Compress(k.source, k.destination); err != nil {
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

	textVerifierInstance := rentTextVerifier()
	defer returnTextVerifier(textVerifierInstance)

	if err := textVerifierInstance.Verify(tmpSource, tmpDestination, textVerifier.CompressIntegrityCheck); err != nil {
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
