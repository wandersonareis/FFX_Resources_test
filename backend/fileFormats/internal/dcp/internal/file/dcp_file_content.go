package file

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/interactions"
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
	log       zerolog.Logger
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
		log:       zerolog.New(os.Stdout).With().Str("module", "dcp_file_content").Logger(),
	}
}

func (c Content) Read(file *os.File) error {
	for i, dataRange := range c.header.DataRanges {
		dataLentgh := dataRange.End - dataRange.Start

		if _, err := file.Seek(dataRange.Start, io.SeekStart); err != nil {
			c.log.Error().Err(err).Msgf("error when positioning in the file: %s", file.Name())
			c.log.Error().Err(err).Msgf("%s", err.Error())
			return err
		}

		data := make([]byte, dataLentgh)

		outputFileName := fmt.Sprintf("%s.%03d", c.options.NameBase, i)

		outputFilePartsPath := filepath.Join(c.outputDir, outputFileName)

		if _, err := io.ReadFull(file, data); err != nil {
			c.log.Error().Err(err).Msgf("error reading data: %s", file.Name())
			c.log.Error().Err(err).Msgf("%s", err.Error())
			return err
		}

		if err := common.WriteBytesToFile(outputFilePartsPath, data); err != nil {
			c.log.Error().Err(err).Msgf("error saving the file %s", outputFilePartsPath)
			return err
		}
	}

	return nil
}
func (c Content) Write(header *Header, parts *[]parts.DcpFileParts) error {
	for _, part := range *parts {
		filePath := part.GetImportLocation().TargetFile
		fileName := filepath.Base(filePath)

		file, err := os.Open(filePath)
		if err != nil {
			c.log.Error().Err(err).Msgf("error when opening the file %s", fileName)
			c.log.Error().Err(err).Msgf("%s", err.Error())
			return err
		}

		defer file.Close()

		if _, err := io.Copy(c.container, file); err != nil {
			c.log.Error().Err(err).Msgf("error recording the data from the file %s to the container", fileName)
			c.log.Error().Err(err).Msgf("%s", err.Error())
			return err
		}
	}

	return nil
}
