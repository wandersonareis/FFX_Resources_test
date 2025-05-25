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
	targetPath := df.destination.Extract().GetTargetPath()
	df.log.Info("Extracting DCP file inside path: %s", targetPath)

	if err := df.extract(); err != nil {
		return err
	}

	df.log.Info("Verifying extracted macrodic file: %s", targetPath)

	if err := df.extractVerify(); err != nil {
		return err
	}

	df.log.Info("System macrodic file extracted successfully to: %s", targetPath)

	return nil
}

func (df *DcpFile) extract() error {
	extractor := NewDcpFileExtractor(
		df.source,
		df.destination,
		df.formatter,
		df.log)

	if err := extractor.Extract(); err != nil {
		return fmt.Errorf("error extracting system macrodic file: %s", err.Error())
	}

	return nil
}

func (df *DcpFile) extractVerify() error {
	targetPath := df.destination.Extract().GetTargetPath()

	if err := df.checkIntegrityService.Verify(
		df.source,
		df.destination,
		integrity.NewDcpExtractionVerificationStrategy()); err != nil {
		return fmt.Errorf("error verifying macrodic file: %s", targetPath)
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
