package text

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/internal/lib"
	"ffxresources/backend/fileFormats/internal/text/internal/mt2"
	"ffxresources/backend/fileFormats/internal/text/textverify"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/loggingService"
	"fmt"
)

type KernelFile struct {
	source      interfaces.ISource
	destination locations.IDestination

	log loggingService.ILoggerService
}

func NewKernel(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	return &KernelFile{
		source:      source,
		destination: destination,

		log: loggingService.NewLoggerHandler("KernelFile"),
	}
}

func (k *KernelFile) GetSource() interfaces.ISource {
	return k.source
}

func (k *KernelFile) Extract() error {
	if err := common.CheckPathExists(k.source.GetPath()); err != nil {
		return fmt.Errorf("kernel file not found: %s", k.source.GetName())
	}

	k.log.Info("Initiating extraction of kernel file: %s", k.source.GetName())

	if err := k.extract(); err != nil {
		return err
	}

	k.log.Info("Verifying the integrity of the extracted kernel file: %s", k.destination.Extract().GetTargetFile())

	if err := k.extractVerify(); err != nil {
		return err
	}

	k.log.Info("Successfully extracted kernel file: %s", k.source.GetName())

	return nil
}

func (k *KernelFile) extract() error {
	mt2.InitExtractionServicePool(k.log)
	extractorInstance := mt2.RentKrnlExtractor()
	defer mt2.ReturnKrnlExtractor(extractorInstance)

	if err := extractorInstance.Extract(k.source, k.destination); err != nil {
		k.log.Error(err, "Error decoding kernel file: %s", k.source.GetName())
		return err
	}

	return nil
}

func (k *KernelFile) extractVerify() error {
	mt2.InitTextVerificationServicePool(k.log)
	verificationService := mt2.RentTextVerifier()
	defer mt2.ReturnTextVerifier(verificationService)

	if err := verificationService.Verify(k.source, k.destination, textverify.NewTextExtractionVerificationStrategy()); err != nil {
		return fmt.Errorf("failed to integrity kernel file: %s", k.source.GetName())
	}

	return nil
}

func (k *KernelFile) Compress() error {
	mt2.InitCompressionServicePool(k.log)
	compressorInstance := mt2.RentKrnlCompressor()
	defer mt2.ReturnKrnlCompressor(compressorInstance)

	if err := compressorInstance.Compress(k.source, k.destination); err != nil {
		k.log.Error(err, "Error compressing kernel file: %s", k.destination.Translate().GetTargetFile())

		return fmt.Errorf("failed to compress kernel file: %s", k.source.GetName())
	}

	tmpSource, tmpDestination := lib.CreateTemp(k.source, k.destination)
	defer tmpDestination.Extract().Dispose()

	tmpFile := NewKernel(tmpSource, tmpDestination)

	if err := tmpFile.Extract(); err != nil {
		k.log.Error(err, "Error decoding kernel file: %s", k.destination.Import().GetTargetFile())

		return fmt.Errorf("failed to decode kernel file: %s", k.destination.Import().GetTargetFile())
	}

	textVerifierInstance := mt2.RentTextVerifier()
	defer mt2.ReturnTextVerifier(textVerifierInstance)

	if err := textVerifierInstance.Verify(tmpSource, tmpDestination, textverify.NewTextCompressionVerificationStrategy()); err != nil {
		k.log.Error(err, "Error verifying kernel file: %s", k.source.GetName())

		return fmt.Errorf("failed to integrity kernel file: %s", k.source.GetName())
	}

	k.log.Info("Kernel file compressed: %s", k.destination.Import().GetTargetFile())

	return nil
}
