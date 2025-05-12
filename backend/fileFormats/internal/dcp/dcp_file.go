package dcp

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/baseFormats"
	"ffxresources/backend/fileFormats/internal/dcp/internal/integrity"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/loggingService"
	"fmt"
)

type DcpFile struct {
	baseFormats.IBaseFileFormat

	formatter   interfaces.ITextFormatter
	fileOptions core.IDcpFileOptions

	log loggingService.ILoggerService
}

func NewDcpFile(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	common.CheckArgumentNil(source, "source")
	common.CheckArgumentNil(destination, "destination")

	return &DcpFile{
		IBaseFileFormat: baseFormats.NewFormatsBase(source, destination),
		formatter:       interactions.NewInteractionService().TextFormatter(),
		log:             loggingService.NewLoggerHandler("dcp_file"),
	}
}

func (df *DcpFile) Extract() error {
	df.ensureFileOptions()

	if err := df.extract(); err != nil {
		return err
	}

	if err := df.extractVerify(); err != nil {
		return err
	}

	df.log.LogInfo("System macrodic file extracted successfully to: %s", df.GetDestination().Extract().GetTargetPath())

	return nil
}

func (df *DcpFile) extract() error {
	df.log.LogInfo("Extracting DCP file inside path: %s", df.GetDestination().Extract().GetTargetPath())

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

	df.log.LogInfo("Verifying extracted macrodic file: %s", targetPath)

	dcpFileIntegrity := integrity.NewDcpFileExtractorIntegrity(df.log)

	if err := dcpFileIntegrity.Verify(targetPath, formatters.NewTxtFormatter(), df.fileOptions); err != nil {
		return fmt.Errorf("error verifying system macrodic file: %s", targetPath)
	}

	return nil
}

func (d *DcpFile) Compress() error {
	d.ensureFileOptions()

	if err := d.compress(); err != nil {
		return err
	}

	if err := d.compressVerify(); err != nil {
		return err
	}

	d.log.LogInfo("Macrodic file compressed: %s", d.GetSource().GetName())

	return nil
}

func (df *DcpFile) compress() error {
	df.log.LogInfo("Compressing DCP file inside path: %s", df.GetDestination().Import().GetTargetPath())

	compressor := NewDcpFileCompressor(
		df.GetSource(),
		df.GetDestination(),
		df.formatter,
		df.fileOptions,
		df.log)

	return compressor.Compress()
}

func (df *DcpFile) compressVerify() error {
	targetFile := df.GetDestination().Import().GetTargetFile()

	df.log.LogInfo("Verifying reimported macrodic file: %s", targetFile)

	dcpFileCompressorIntegrity := integrity.NewDcpFileCompressorVerify(df.log)

	if err := dcpFileCompressorIntegrity.Verify(targetFile, formatters.NewTxtFormatter(), df.fileOptions); err != nil {
		return fmt.Errorf("error verifying system macrodic file: %s", targetFile)
	}

	return nil
}
func (df *DcpFile) ensureFileOptions() {
	if df.fileOptions == nil {
		df.fileOptions = core.NewDcpFileOptions(interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber())
	}
}
