package text

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/internal/mt2"
	"ffxresources/backend/fileFormats/internal/text/textVerifier"
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

func (k *kernelFile) GetSource() interfaces.ISource {
	return k.source
}

func (k *kernelFile) Extract() error {
	if !common.IsFileExists(k.source.Get().Path) {
		k.log.LogError(nil, "Kernel file not found: %s", k.source.Get().Name)

		return fmt.Errorf("kernel file not found: %s", k.source.Get().Name)
	}

	k.log.LogInfo("Extracting kernel file: %s", k.source.Get().Name)

	extractorInstance := mt2.RentKrnlExtractor()
	defer mt2.ReturnKrnlExtractor(extractorInstance)
	fmt.Println("extractorInstance", extractorInstance)

	if err := extractorInstance.Extract(k.source, k.destination); err != nil {
		k.log.LogError(err, "Error extracting kernel file: %s", k.source.Get().Name)

		return fmt.Errorf("failed to decode kernel file: %s", k.source.Get().Name)
	}

	k.log.LogInfo("Verifying extracted kernel file: %s", k.destination.Extract().GetTargetFile())

	verifierInstance := mt2.RentTextVerifier()
	defer mt2.ReturnTextVerifier(verifierInstance)
	fmt.Println("verifierInstance", verifierInstance)

	if err := verifierInstance.Verify(k.source, k.destination, textVerifier.NewTextExtractVerify()); err != nil {
		k.log.LogError(err, "Error verifying kernel file: %s", k.source.Get().Name)

		return fmt.Errorf("failed to integrity kernel file: %s", k.source.Get().Name)
	}

	k.log.LogInfo("Kernel file extracted: %s", k.source.Get().Name)

	return nil
}

func (k *kernelFile) Compress() error {
	compressorInstance := mt2.RentKrnlCompressor()
	defer mt2.ReturnKrnlCompressor(compressorInstance)

	if err := compressorInstance.Compress(k.source, k.destination); err != nil {
		k.log.LogError(err, "Error compressing kernel file: %s", k.destination.Translate().GetTargetFile())

		return fmt.Errorf("failed to compress kernel file: %s", k.source.Get().Name)
	}

	tmpSource, tmpDestination := k.createTemp(k.source, k.destination)
	defer tmpDestination.Extract().Dispose()

	tmpFile := NewKernel(tmpSource, tmpDestination)

	if err := tmpFile.Extract(); err != nil {
		k.log.LogError(err, "Error decoding kernel file: %s", k.destination.Import().GetTargetFile())

		return fmt.Errorf("failed to decode kernel file: %s", k.destination.Import().GetTargetFile())
	}

	textVerifierInstance := mt2.RentTextVerifier()
	defer mt2.ReturnTextVerifier(textVerifierInstance)

	if err := textVerifierInstance.Verify(tmpSource, tmpDestination, textVerifier.NewTextCompressVerify()); err != nil {
		k.log.LogError(err, "Error verifying kernel file: %s", k.source.Get().Name)

		return fmt.Errorf("failed to integrity kernel file: %s", k.source.Get().Name)
	}

	k.log.LogInfo("Kernel file compressed: %s", k.destination.Import().GetTargetFile())

	return nil
}

func (k *kernelFile) createTemp(source interfaces.ISource, destination locations.IDestination) (interfaces.ISource, locations.IDestination) {
	tmp := common.NewTempProvider("tmp", ".txt")

	tmpSource := source
	tmpDestination := destination

	tmpDestination.Extract().SetTargetFile(tmp.TempFile)
	tmpDestination.Extract().SetTargetPath(tmp.TempFilePath)

	tmpSource.Get().Path = destination.Import().GetTargetFile()

	return tmpSource, tmpDestination
}
