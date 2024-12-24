package file

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

type Content struct {
	header    *Header
	container *bytes.Buffer
	options   interactions.DcpFileOptions
	outputDir string

	log zerolog.Logger
}

func NewContent(header *Header, outputDir string) *Content {
	return &Content{
		header:    header,
		options:   interactions.NewInteractionService().DcpAndLockitOptions.GetDcpFileOptions(),
		outputDir: outputDir,

		log: logger.Get().With().Str("module", "dcp_file_content").Logger(),
	}
}

func NewContentWithBuffer(container *bytes.Buffer) *Content {
	return &Content{
		container: container,

		log: logger.Get().With().Str("module", "dcp_file_content").Logger(),
	}
}

func (c Content) Read(file *os.File) error {
	for i, dataRange := range c.header.DataRanges {
		dataLentgh := dataRange.End - dataRange.Start

		if _, err := file.Seek(dataRange.Start, io.SeekStart); err != nil {
			c.log.Error().
				Err(err).
				Str("file", file.Name()).
				Msg("error when positioning in the file")

			return err
		}

		data := make([]byte, dataLentgh)

		outputFileName := fmt.Sprintf("%s.%03d", c.options.NameBase, i)

		outputFilePartsPath := filepath.Join(c.outputDir, outputFileName)

		if _, err := io.ReadFull(file, data); err != nil {
			c.log.Error().
				Err(err).
				Str("file", file.Name()).
				Msg("error reading data")

			return err
		}

		if err := common.WriteBytesToFile(outputFilePartsPath, data); err != nil {
			c.log.Error().
				Err(err).
				Str("file", outputFilePartsPath).
				Msg("error saving the file")

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
			c.log.Error().
				Err(err).
				Str("file", filePath).
				Msg("error when opening the file")

			return err
		}

		defer file.Close()

		if _, err := io.Copy(c.container, file); err != nil {
			c.log.Error().
				Err(err).
				Str("file", filePath).
				Msg("error recording the data from the file")

			return err
		}
	}

	return nil
}
