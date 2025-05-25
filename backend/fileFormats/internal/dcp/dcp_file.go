package dcp

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/dcp/internal/integrity"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/loggingService"
	"fmt"
)

type DcpFile struct {
	source                interfaces.ISource
	destination           locations.IDestination
	formatter             interfaces.ITextFormatter
	checkIntegrityService components.IVerificationService

	log loggingService.ILoggerService
}

func NewDcpFile(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	return &DcpFile{
		formatter:             interactions.NewInteractionService().TextFormatter(),
		checkIntegrityService: components.NewVerificationService(),
		log:                   loggingService.NewLoggerHandler("dcp_file"),
		source:                source,
		destination:           destination,
	}
}

func (df *DcpFile) GetSource() interfaces.ISource {
	return df.source
}

func (df *DcpFile) Extract() error {
	df.ensureFileOptions()

	if err := df.extract(); err != nil {
		return err
	}

	if err := df.extractVerify(); err != nil {
		return err
	}

	df.log.Info("System macrodic file extracted successfully to: %s", df.GetDestination().Extract().GetTargetPath())

	return nil
}

func (df *DcpFile) extract() error {
	df.log.Info("Extracting DCP file inside path: %s", df.GetDestination().Extract().GetTargetPath())

	extractor := NewDcpFileExtractor(
		df.GetSource(),
		df.GetDestination(),
		df.formatter,
		df.fileOptions,
		df.log)

	return extractor.Extract()
}

func (df *DcpFile) extractVerify() error {
	targetPath := df.GetDestination().Extract().GetTargetPath()

	df.log.Info("Verifying extracted macrodic file: %s", targetPath)

	dcpFileIntegrity := integrity.NewDcpFileExtractorIntegrity(df.log)

	if err := dcpFileIntegrity.Verify(targetPath, formatters.NewTxtFormatter(), df.fileOptions); err != nil {
		return fmt.Errorf("error verifying system macrodic file: %s", targetPath)
	}

	return nil
}

func (df *DcpFile) Compress() error {
	if err := df.compress(); err != nil {
		return err
	}

	if err := df.compressVerify(); err != nil {
		return err
	}

	df.log.Info("Macrodic file compressed: %s", df.GetSource().GetName())

	return nil
}

func (df *DcpFile) compress() error {
	df.log.Info("Compressing DCP file inside path: %s", df.destination.Import().GetTargetPath())

	compressor := NewDcpFileCompressor(
		df.source,
		df.destination,
		df.formatter,
		df.log)

	return compressor.Compress()
}

func (df *DcpFile) compressVerify() error {
	targetFile := df.destination.Import().GetTargetFile()

	df.log.Info("Verifying reimported macrodic file: %s", targetFile)

	if err := df.checkIntegrityService.Verify(
		df.source,
		df.destination,
		integrity.NewDcpCompressionVerificationStrategy(targetFile)); err != nil {
		return fmt.Errorf("error verifying system macrodic file: %s", targetFile)
	}

	return nil
}
func (df *DcpFile) ensureFileOptions() {
	if df.fileOptions == nil {
		df.fileOptions = core.NewDcpFileOptions(interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber())
	}
}
