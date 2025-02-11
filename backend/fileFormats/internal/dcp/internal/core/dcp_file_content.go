package dcpCore

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/dcp/internal/dcpParts"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type (
	IDcpFileContent interface {
		Read(file *os.File) error
		Write(header *dcpFileHeader, parts components.IList[dcpParts.DcpFileParts]) error
	}

	dcpFileContent struct {
		header    *dcpFileHeader
		container *bytes.Buffer
		options   core.IDcpFileOptions
		outputDir string
	}
)

func newContent(header *dcpFileHeader, outputDir string, fileOptions core.IDcpFileOptions) IDcpFileContent {
	return &dcpFileContent{
		header:    header,
		outputDir: outputDir,
		options:   fileOptions,
	}
}

func NewContentWithBuffer(container *bytes.Buffer) *dcpFileContent {
	return &dcpFileContent{
		container: container,
	}
}

func (c dcpFileContent) Read(file *os.File) error {
	for i, dataRange := range c.header.DataRanges {
		dataLentgh := dataRange.End - dataRange.Start

		if _, err := file.Seek(dataRange.Start, io.SeekStart); err != nil {
			return fmt.Errorf("error when positioning in the file: %s", file.Name())
		}

		data := make([]byte, dataLentgh)

		outputFileName := fmt.Sprintf("%s.%03d", c.options.GetNameBase(), i)

		outputFilePartsPath := filepath.Join(c.outputDir, outputFileName)

		if _, err := io.ReadFull(file, data); err != nil {
			return fmt.Errorf("error reading data: %s", file.Name())
		}

		if err := common.WriteBytesToFile(outputFilePartsPath, data); err != nil {
			return fmt.Errorf("error saving the file: %s", outputFilePartsPath)
		}
	}

	return nil
}
func (c dcpFileContent) Write(header *dcpFileHeader, parts components.IList[dcpParts.DcpFileParts]) error {
	for _, part := range parts.GetItems() {
		filePath := part.GetDestination().Import().Get().GetTargetFile()

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("error when opening the file: %s", filePath)
		}

		defer file.Close()

		if _, err := io.Copy(c.container, file); err != nil {
			return fmt.Errorf("error recording the data from the file: %s", filePath)
		}
	}

	return nil
}
