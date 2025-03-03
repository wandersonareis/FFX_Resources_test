package dlg

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/lib/textVerifier"
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

		log: &logger.LogHandler{
			Logger: logger.Get().With().Str("module", "dialogs_file").Logger(),
		},
	}
}

func (d *DialogsFile) GetSource() interfaces.ISource {
	return d.source
}

func (d *DialogsFile) Extract() error {
	extractorInstance := rentDlgExtractor()
	defer returnDlgExtractor(extractorInstance)

	d.log.LogInfo("Extracting dialog file: %s", d.source.Get().Name)

	if err := extractorInstance.Extract(d.source, d.destination); err != nil {
		d.log.LogError(err, "Error decoding dialog file: %s", d.source.Get().Name)
		return err
	}

	verifierInstance := rentTextVerifier()
	defer returnTextVerifier(verifierInstance)

	d.log.LogInfo("Verifying extracted dialog file: %s", d.destination.Extract().Get().GetTargetFile())

	if err := verifierInstance.Verify(d.source, d.destination, textVerifier.ExtractIntegrityCheck); err != nil {
		d.log.LogError(err, "Error verifying dialog file: %s", d.source.Get().Name)

		return fmt.Errorf("failed to integrity dialog file: %s", d.source.Get().Name)
	}

	d.log.LogInfo("Dialog file extracted: %s", d.source.Get().Name)

	return nil
}

func (d *DialogsFile) Compress() error {
	compressorInstance := rentDlgCompressor()
	defer returnDlgCompressor(compressorInstance)

	d.log.LogInfo("Compressing dialog file: %s", d.destination.Import().Get().GetTargetFile())
	if err := compressorInstance.Compress(d.source, d.destination); err != nil {
		return err
	}

	d.log.LogInfo("Verifying compressed dialog file: %s", d.destination.Import().Get().GetTargetFile())

	tmpSource, tmpDestination := d.createTemp(d.source, d.destination)
	defer tmpDestination.Extract().Get().Dispose()

	tmpFile := NewDialogs(tmpSource, tmpDestination)

	if err := tmpFile.Extract(); err != nil {
		d.log.LogError(err, "Error decoding dialog file: %s", d.destination.Import().Get().GetTargetFile())

		return fmt.Errorf("failed to decode dialog file: %s", d.destination.Import().Get().GetTargetFile())
	}

	verifierInstance := rentTextVerifier()
	defer returnTextVerifier(verifierInstance)

	if err := verifierInstance.Verify(tmpSource, tmpDestination, textVerifier.CompressIntegrityCheck); err != nil {
		d.log.LogError(err, "Error verifying dialog file: %s", d.source.Get().Name)

		return fmt.Errorf("failed to integrity dialog file: %s", d.source.Get().Name)
	}

	d.log.LogInfo("Dialog file compressed: %s", d.destination.Import().Get().GetTargetFile())

	return nil
}

func (d *DialogsFile) createTemp(source interfaces.ISource, destination locations.IDestination) (interfaces.ISource, locations.IDestination) {
	tmp := common.NewTempProvider("tmp", ".txt")

	tmpSource := source
	tmpDestination := destination

	tmpDestination.Extract().Get().SetTargetFile(tmp.TempFile)
	tmpDestination.Extract().Get().SetTargetPath(tmp.TempFilePath)

	tmpSource.Get().Path = destination.Import().Get().GetTargetFile()

	return tmpSource, tmpDestination
}
