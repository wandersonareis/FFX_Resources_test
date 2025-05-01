package dcpCore

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"fmt"
	"os"
)

type (
	IDcpFileSpliter interface {
		FileSplitter(source interfaces.ISource, destination locations.IDestination, fileOptions core.IDcpFileOptions) error
	}

	dcpFileSpliter struct{}
)

func NewDcpFileSpliter() IDcpFileSpliter {
	return &dcpFileSpliter{}
}

func (ds *dcpFileSpliter) FileSplitter(source interfaces.ISource, destination locations.IDestination, fileOptions core.IDcpFileOptions) error {
	targetFile := source.Get().Path

	extractLocation := destination.Extract()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		return fmt.Errorf("error when providing the extraction directory: %s", extractLocation.GetTargetPath())
	}

	if err := ds.dcpReader(targetFile, extractLocation.GetTargetPath(), fileOptions); err != nil {
		return err
	}

	return nil
}

func (ds *dcpFileSpliter) dcpReader(dcpFilePath, outputDir string, fileOptions core.IDcpFileOptions) error {
	dcpFileStream, err := os.Open(dcpFilePath)
	if err != nil {
		return fmt.Errorf("error when opening the file %s", dcpFilePath)
	}

	defer dcpFileStream.Close()

	header := newHeader()
	if err := header.FromFile(dcpFilePath); err != nil {
		return err
	}

	if err := header.DataLengths(header, dcpFileStream); err != nil {
		return err
	}

	content := newContent(header, outputDir, fileOptions)

	if err := content.Read(dcpFileStream); err != nil {
		return err
	}

	return nil
}
