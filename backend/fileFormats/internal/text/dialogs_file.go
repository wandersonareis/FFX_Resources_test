package text

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/internal/dlg"
	"ffxresources/backend/fileFormats/internal/text/textVerifier"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
)

type DialogsFile struct {
	source      interfaces.ISource
	destination locations.IDestination

	log logger.ILoggerHandler
}

func NewDialogs(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	return &DialogsFile{
		source:      source,
		destination: destination,

		log: logger.NewLoggerHandler("DialogsFile"),
	}
}

func (d *DialogsFile) GetSource() interfaces.ISource {
	return d.source
}

func (d *DialogsFile) Extract() error {
	if err := common.CheckPathExists(d.source.GetPath()); err != nil {
		return fmt.Errorf("dialog file not found: %s", d.source.GetName())
	}

	d.log.LogInfo("Initiating extraction of kernel file: %s", d.source.GetName())

	if err := d.extract(); err != nil {
		return err
	}

	d.log.LogInfo("Verifying the integrity of the extracted dialog file: %s", d.destination.Extract().GetTargetFile())

	if err := d.extractVerify(); err != nil {
		return err
	}

	d.log.LogInfo("Successfully extracted dialog file: %s", d.source.GetName())

	return nil
}

func (d *DialogsFile) extract() error {
	dlg.InitExtractorsPool(d.log)
	extractorInstance := dlg.RentDlgExtractor()
	defer dlg.ReturnDlgExtractor(extractorInstance)

	if err := extractorInstance.Extract(d.source, d.destination); err != nil {
		d.log.LogError(err, "Error decoding dialog file: %s", d.source.GetName())
		return err
	}

	return nil
}

func (d *DialogsFile) extractVerify() error {
	dlg.InitTextVerifierPool(d.log)
	verifierInstance := dlg.RentTextVerifier()
	defer dlg.ReturnTextVerifier(verifierInstance)

	if err := verifierInstance.Verify(d.source, d.destination, textVerifier.NewTextExtractionVerificationStrategy()); err != nil {
		d.log.LogError(err, "Error verifying dialog file: %s", d.source.GetName())

		return fmt.Errorf("failed to integrity dialog file: %s", d.source.GetName())
	}

	return nil
}
func (d *DialogsFile) Compress() error {
	dlg.InitCompressorsPool(d.log)
	compressorInstance := dlg.RentDlgCompressor()
	defer dlg.ReturnDlgCompressor(compressorInstance)

	d.log.LogInfo("Compressing dialog file: %s", d.destination.Import().GetTargetFile())

	if err := d.ensureTranslatedText(); err != nil {
		return err
	}

	if err := compressorInstance.Compress(d.source, d.destination); err != nil {
		return err
	}

	d.log.LogInfo("Verifying compressed dialog file: %s", d.destination.Import().GetTargetFile())

	tmpSource, tmpDestination := d.createTemp(d.source, d.destination)
	defer tmpDestination.Extract().Dispose()

	tmpFile := NewDialogs(tmpSource, tmpDestination)

	if err := tmpFile.Extract(); err != nil {
		d.log.LogError(err, "Error decoding dialog file: %s", d.destination.Import().GetTargetFile())

		return fmt.Errorf("failed to decode dialog file: %s", d.destination.Import().GetTargetFile())
	}

	dlg.InitTextVerifierPool(d.log)
	verifierInstance := dlg.RentTextVerifier()
	defer dlg.ReturnTextVerifier(verifierInstance)

	if err := verifierInstance.Verify(tmpSource, tmpDestination, textVerifier.NewTextCompressionVerificationStrategy()); err != nil {
		d.log.LogError(err, "Error verifying dialog file: %s", d.destination.Import().GetTargetFile())
		return fmt.Errorf("failed to integrity dialog file: %s", d.source.GetName())
	}

	d.log.LogInfo("Dialog file compressed: %s", d.destination.Import().GetTargetFile())

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

	if err := textVerifierInstance.Verify(d.source, d.destination, textVerifier.NewTextSegmentsVerificationStrategy()); err != nil {
		return fmt.Errorf("translated file segments count mismatch: %s", targetFile)
	}

	return nil
}

func (d *DialogsFile) createTemp(source interfaces.ISource, destination locations.IDestination) (interfaces.ISource, locations.IDestination) {
	tmp := common.NewTempProvider("tmp", ".txt")

	tmpSource := source
	tmpSource.Get().Path = destination.Import().GetTargetFile()

	extractLocation := destination.Extract().Copy()
	extractLocation.SetTargetFile(tmp.TempFile)
	extractLocation.SetTargetPath(tmp.TempFilePath)

	tmpDestination := &locations.Destination{
		ExtractLocation:   &extractLocation,
		TranslateLocation: destination.Translate(),
		ImportLocation:    destination.Import(),
	}

	return tmpSource, tmpDestination
}
