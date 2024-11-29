package splitter

import (
	"ffxresources/backend/fileFormats/internal/dcp/internal/file"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type IDcpFileSpliter interface {
	Split(dataInfo interactions.IGameDataInfo) error
}

type DcpFileSpliter struct {
	log zerolog.Logger
}

func NewDcpFileSpliter() IDcpFileSpliter {
	return &DcpFileSpliter{
		log: logger.Get().With().Str("module", "dcp_file_splitter").Logger(),
	}
}

func (ds *DcpFileSpliter) Split(dataInfo interactions.IGameDataInfo) error {
	targetFile := dataInfo.GetGameData().FullFilePath

	extractLocation := dataInfo.GetExtractLocation()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		ds.log.Error().
			Err(err).
			Str("path", extractLocation.TargetPath).
			Msg("error when providing the extraction directory")

		return fmt.Errorf("error when creating the extraction directory")
	}

	if err := ds.dcpReader(targetFile, extractLocation.TargetPath); err != nil {
		return err
	}

	return nil
}

func (ds *DcpFileSpliter) dcpReader(dcpFilePath, outputDir string) error {
	dcpFileStream, err := os.Open(dcpFilePath)
	if err != nil {
		ds.log.Error().
			Err(err).
			Str("file", dcpFilePath).
			Msg("error when opening the file")

		return fmt.Errorf("error when opening the file %s", dcpFilePath)
	}

	defer dcpFileStream.Close()

	header := file.NewHeader()
	header.FromFile(dcpFilePath)

	if err := header.DataLengths(header, dcpFileStream); err != nil {
		ds.log.Error().
			Err(err).
			Msg("error when calculating the data intervals")

		return fmt.Errorf("error when calculating the data intervals")
	}

	content := file.NewContent(header, outputDir)

	if err := content.Read(dcpFileStream); err != nil {
		ds.log.Error().
			Err(err).
			Msg("error when reading the data")

		return fmt.Errorf("error reading the data")
	}

	return nil
}
