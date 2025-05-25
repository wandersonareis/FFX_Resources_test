package dcpFileHandler

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"fmt"
	"os"
	"path/filepath"
)

type (
	IDcpFileSplitter interface {
		Split(source interfaces.ISource, destination locations.IDestination, fileOptions models.IDcpFileProperties) error
	}

	dcpFileSplitter struct {
		reader *dcpFileReader
		writer *dcpFileWriter
	}
)

func NewDcpFileSplitter() IDcpFileSplitter {
	return &dcpFileSplitter{
		reader: newDcpFileReader(),
		writer: newDcpFileWriter(),
	}
}

func (dfs *dcpFileSplitter) Split(source interfaces.ISource, destination locations.IDestination, fileOptions models.IDcpFileProperties) error {
	targetFile := source.GetPath()
	extractLocation := destination.Extract()

	if err := dfs.validateDestination(extractLocation); err != nil {
		return fmt.Errorf("error when providing the extraction directory: %w", err)
	}

	fileData, err := dfs.readInputFile(targetFile)
	if err != nil {
		return err
	}

	chunks, err := dfs.reader.GetChunks(fileData)
	if err != nil {
		return err
	}

	if err := dfs.writeChunks(chunks, extractLocation.GetTargetPath(), fileOptions); err != nil {
		return err
	}

	return nil
}

func (ds *dcpFileSplitter) validateDestination(extractLocation locations.IExtractLocation) error {
	return extractLocation.ProvideTargetPath()
}

func (ds *dcpFileSplitter) readInputFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (ds *dcpFileSplitter) writeChunks(chunks []*Chunk, outputDir string, fileOptions models.IDcpFileProperties) error {
	for i, chunk := range chunks {
		if chunk.Data == nil {
			continue
		}

		outputFileName := fmt.Sprintf("%s.%03d", fileOptions.GetNameBase(), i)
		outputFilePath := filepath.Join(outputDir, outputFileName)

		if err := ds.writer.WriteFile(outputFilePath, chunk.Data); err != nil {
			return fmt.Errorf("error writing file: %s, %w", outputFilePath, err)
		}
	}
	return nil
}
