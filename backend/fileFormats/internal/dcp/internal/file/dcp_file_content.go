package file

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/logger"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Content struct {
	header    *Header
	container *bytes.Buffer
	options   core.IDcpFileOptions
	outputDir string

	log logger.ILoggerHandler
}

func NewContent(header *Header, outputDir string, fileOptions core.IDcpFileOptions, logger logger.ILoggerHandler) *Content {
	return &Content{
		header:    header,
		outputDir: outputDir,
		options:   fileOptions,

		log: logger,
	}
}

func NewContentWithBuffer(container *bytes.Buffer) *Content {
	return &Content{
		container: container,

		log: &logger.LogHandler{
			Logger: logger.Get().With().Str("module", "dcp_file_content").Logger(),
		},
	}
}

func (c Content) Read(file *os.File) error {
	for i, dataRange := range c.header.DataRanges {
		dataLentgh := dataRange.End - dataRange.Start

		if _, err := file.Seek(dataRange.Start, io.SeekStart); err != nil {
			c.log.LogError(err, "error when positioning in the file: %s", file.Name())
			return err
		}

		data := make([]byte, dataLentgh)

		outputFileName := fmt.Sprintf("%s.%03d", c.options.GetNameBase(), i)

		outputFilePartsPath := filepath.Join(c.outputDir, outputFileName)

		if _, err := io.ReadFull(file, data); err != nil {
			c.log.LogError(err, "error reading data")
			return err
		}

		if err := common.WriteBytesToFile(outputFilePartsPath, data); err != nil {
			c.log.LogError(err, "error saving the file: %s", outputFilePartsPath)

			return err
		}
	}

	return nil
}
func (c Content) Write(header *Header, parts components.IList[parts.DcpFileParts]) error {
	for _, part := range parts.GetItems() {
		filePath := part.Destination().Import().Get().GetTargetFile()

		file, err := os.Open(filePath)
		if err != nil {
			c.log.LogError(err, "error when opening the file: %s", filePath)

			return err
		}

		defer file.Close()

		if _, err := io.Copy(c.container, file); err != nil {
			c.log.LogError(err, "error recording the data from the file: %s", filePath)
			return err
		}
	}

	return nil
}
