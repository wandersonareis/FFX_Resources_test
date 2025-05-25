package dcpFileHandler

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/dcp/internal/dcpParts"
	"ffxresources/backend/interfaces"
	"fmt"
	"os"
)

type (
	IDcpFileMerger interface {
		Merge(source interfaces.ISource, destination locations.IDestination, filePartsList components.IList[dcpParts.DcpFileParts]) error
	}

	dcpFileMerger struct {
		reader *dcpFileReader
		writer *dcpFileWriter
	}
)

func NewDcpFileMerger() IDcpFileMerger {
	return &dcpFileMerger{
		reader: newDcpFileReader(),
		writer: newDcpFileWriter(),
	}
}

func (dfm *dcpFileMerger) Merge(source interfaces.ISource, destination locations.IDestination, filePartsList components.IList[dcpParts.DcpFileParts]) error {
	inputDcpFile := source.GetPath()
	newDcpFile := destination.Import().GetTargetFile()

	if err := dfm.validateDestination(destination); err != nil {
		return fmt.Errorf("error validating destination: %w", err)
	}

	inputFileData, err := dfm.readInputFile(inputDcpFile)
	if err != nil {
		return err
	}

	chunks, err := dfm.reader.GetChunks(inputFileData)
	if err != nil {
		return fmt.Errorf("error when getting chunks: %w", err)
	}

	updatedChunks, err := dfm.writer.UpdateChunks(chunks, filePartsList)
	if err != nil {
		return fmt.Errorf("error when updating chunks: %w", err)
	}

	if err := dfm.writer.SaveContainerFile(newDcpFile, updatedChunks); err != nil {
		return fmt.Errorf("error when saving the container file: %w", err)
	}

	return nil
}

func (dfm *dcpFileMerger) readInputFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (dfm *dcpFileMerger) validateDestination(destination locations.IDestination) error {
	return destination.Import().ProvideTargetPath()
}
