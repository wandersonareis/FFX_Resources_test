package internal

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Content struct {
	header    *Header
	container *bytes.Buffer
	options   *interactions.DcpFileOptions
	outputDir string
}

func NewContent(header *Header, outputDir string) *Content {
	return &Content{
		header:    header,
		options:   interactions.NewInteraction().GamePartOptions.GetDcpFileOptions(),
		outputDir: outputDir,
	}
}

func NewContentWithBuffer(container *bytes.Buffer) *Content {
	return &Content{
		container: container,
	}
}

func (c Content) Read(file *os.File) error {
	for i, dataRange := range c.header.DataRanges {
		dataLentgh := dataRange.End - dataRange.Start

		if _, err := file.Seek(dataRange.Start, io.SeekStart); err != nil {
			return fmt.Errorf("error when positioning in the file: %w", err)
		}

		data := make([]byte, dataLentgh)

		outputFileName := fmt.Sprintf("%s.%03d", c.options.NameBase, i)

		outputFilePartsPath := filepath.Join(c.outputDir, outputFileName)

		if _, err := io.ReadFull(file, data); err != nil {
			return fmt.Errorf("error reading data: %w", err)
		}

		if err := common.WriteBytesToFile(outputFilePartsPath, data); err != nil {
			return fmt.Errorf("erro ao salvar o arquivo %03d: %w", i, err)
		}
	}

	return nil
}
func (c Content) Write(header *Header, parts *[]DcpFileParts) error {
	for _, part := range *parts {
		filePath := part.GetGameData().FullFilePath
		fileName := filepath.Base(filePath)

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("error when opening the file %s: %w", fileName, err)
		}

		defer file.Close()

		if _, err := io.Copy(c.container, file); err != nil {
			return fmt.Errorf("error recording the data from the file %s to the container: %w", fileName, err)
		}
	}

	return nil
}
