package text

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/internal/dlg"
	"ffxresources/backend/fileFormats/internal/text/internal/lib"
	"ffxresources/backend/fileFormats/internal/text/textverify"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/loggingService"
	"fmt"
)

type DialogsFile struct {
	source      interfaces.ISource
	destination locations.IDestination

	log loggingService.ILoggerService
}

func NewDialogs(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	return &DialogsFile{
		source:      source,
		destination: destination,

		log: loggingService.NewLoggerHandler("DialogsFile"),
	}
}

func (d *DialogsFile) GetSource() interfaces.ISource {
	return d.source
}

func (d *DialogsFile) Extract() error {
	if err := common.CheckPathExists(d.source.GetPath()); err != nil {
		return fmt.Errorf("dialog file not found: %s", d.source.GetName())
	}

	d.log.Info("Initiating extraction of dialog file: %s", d.source.GetName())

	if err := d.extract(); err != nil {
		return err
	}

	d.log.Info("Verifying the integrity of the extracted dialog file: %s", d.destination.Extract().GetTargetFile())

	if err := d.extractVerify(); err != nil {
		d.log.Error(err, "Error verifying extracted dialog file")
		return err
	}

	d.log.Info("Successfully extracted dialog file: %s", d.source.GetName())

	return nil
}

func (d *DialogsFile) extract() error {
	dlg.InitExtractorsPool(d.log)
	extractorInstance := dlg.RentDlgExtractor()
	defer dlg.ReturnDlgExtractor(extractorInstance)

	if err := extractorInstance.Extract(d.source, d.destination); err != nil {
		d.log.Error(err, "Error decoding dialog file: %s", d.source.GetName())
		return err
	}

	return nil
}

func (d *DialogsFile) extractVerify() error {
	dlg.InitTextVerifierPool(d.log)
	verifierInstance := dlg.RentTextVerifier()
	defer dlg.ReturnTextVerifier(verifierInstance)

	if err := verifierInstance.Verify(d.source, d.destination, textverify.NewTextExtractionVerificationStrategy()); err != nil {
		return fmt.Errorf("error verifying dialog file at path %s: %v", d.source.GetPath(), err)
	}

	return nil
}
func (d *DialogsFile) Compress() error {
	dlg.InitCompressorsPool(d.log)
	compressorInstance := dlg.RentDlgCompressor()
	defer dlg.ReturnDlgCompressor(compressorInstance)

	d.log.Info("Compressing dialog file: %s", d.destination.Import().GetTargetFile())

	if err := d.ensureTranslatedText(); err != nil {
		return err
	}

	if err := compressorInstance.Compress(d.source, d.destination); err != nil {
		return err
	}

	d.log.Info("Verifying compressed dialog file: %s", d.destination.Import().GetTargetFile())

	tmpSource, tmpDestination := lib.CreateTemp(d.source, d.destination)
	defer tmpDestination.Extract().Dispose()

	tmpFile := NewDialogs(tmpSource, tmpDestination)

	if err := tmpFile.Extract(); err != nil {
		d.log.Error(err, "Error decoding dialog file: %s", d.destination.Import().GetTargetFile())

		return fmt.Errorf("failed to decode dialog file: %s", d.destination.Import().GetTargetFile())
	}

	dlg.InitTextVerifierPool(d.log)
	verifierInstance := dlg.RentTextVerifier()
	defer dlg.ReturnTextVerifier(verifierInstance)

	if err := verifierInstance.Verify(tmpSource, tmpDestination, textverify.NewTextCompressionVerificationStrategy()); err != nil {
		d.log.Error(err, "Error verifying dialog file: %s", d.destination.Import().GetTargetFile())
		return fmt.Errorf("failed to integrity dialog file: %s", d.source.GetName())
	}

	d.log.Info("Dialog file compressed: %s", d.destination.Import().GetTargetFile())

	return nil
}

func (d *DialogsFile) ensureTranslatedText() error {
	dlg.InitTextVerifierPool(d.log)
	textVerifierInstance := dlg.RentTextVerifier()
	defer dlg.ReturnTextVerifier(textVerifierInstance)

	sourceFile := d.source.GetPath()
	targetFile := d.destination.Translate().GetTargetFile()

	if err := common.CheckPathExists(sourceFile); err != nil {
		return fmt.Errorf("failed to check source file path: %s", err)
	}

	if err := common.CheckPathExists(targetFile); err != nil {
		return fmt.Errorf("failed to check target file path: %s", err)
	}

	if err := textVerifierInstance.Verify(d.source, d.destination, textverify.NewTextSegmentsVerificationStrategy(targetFile)); err != nil {
		return fmt.Errorf("translated file segments count mismatch: %s", targetFile)
	}

	return nil
}
