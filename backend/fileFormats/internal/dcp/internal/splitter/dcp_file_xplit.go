package splitter

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/dcp/internal/file"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
	"os"
)

type IDcpFileSpliter interface {
	Split(source interfaces.ISource, destination locations.IDestination, fileOptions core.IDcpFileOptions) error
}

type DcpFileSpliter struct {
	log logger.ILoggerHandler
}

func NewDcpFileSpliter(logger logger.ILoggerHandler) IDcpFileSpliter {
	return &DcpFileSpliter{
		log: logger,
	}
}

func (ds *DcpFileSpliter) Split(source interfaces.ISource, destination locations.IDestination, fileOptions core.IDcpFileOptions) error {
	targetFile := source.Get().Path

	extractLocation := destination.Extract().Get()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		ds.log.LogError(err, "error when providing the extraction directory: %s", extractLocation.GetTargetPath())

		return fmt.Errorf("error when creating the extraction directory")
	}

	if err := ds.dcpReader(targetFile, extractLocation.GetTargetPath(), fileOptions); err != nil {
		return err
	}

	return nil
}

func (ds *DcpFileSpliter) dcpReader(dcpFilePath, outputDir string, fileOptions core.IDcpFileOptions) error {
	dcpFileStream, err := os.Open(dcpFilePath)
	if err != nil {
		ds.log.LogError(err, "error when opening the file: %s", dcpFilePath)

		return fmt.Errorf("error when opening the file %s", dcpFilePath)
	}

	defer dcpFileStream.Close()

	header := file.NewHeader()
	header.FromFile(dcpFilePath)

	if err := header.DataLengths(header, dcpFileStream); err != nil {
		ds.log.LogError(err, "error when calculating the data intervals")

		return fmt.Errorf("error when calculating the data intervals")
	}

	content := file.NewContent(header, outputDir, fileOptions, ds.log)

	if err := content.Read(dcpFileStream); err != nil {
		ds.log.LogError(err, "error when reading the data")

		return fmt.Errorf("error reading the data")
	}

	return nil
}
