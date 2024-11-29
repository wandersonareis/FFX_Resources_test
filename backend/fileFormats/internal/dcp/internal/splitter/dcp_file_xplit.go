package splitter

import (
	"ffxresources/backend/fileFormats/internal/dcp/internal/file"
	"ffxresources/backend/interactions"
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
		log: zerolog.New(os.Stdout).With().Str("module", "dcp_file_splitter").Logger(),
	}
}

func (ds *DcpFileSpliter) Split(dataInfo interactions.IGameDataInfo) error {
	targetFile := dataInfo.GetGameData().FullFilePath

	extractLocation := dataInfo.GetExtractLocation()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		ds.log.Error().Err(err).Msgf("error when providing the extraction directory: %v", err)
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
		ds.log.Error().Err(err).Msgf("error when opening the file %s", dcpFilePath)
		ds.log.Error().Err(err).Msgf("%s", err.Error())
		return fmt.Errorf("error when opening the file %s", dcpFilePath)
	}

	defer dcpFileStream.Close()

	header := file.NewHeader()
	header.FromFile(dcpFilePath)

	if err := header.DataLengths(header, dcpFileStream); err != nil {
		return fmt.Errorf("error when calculating the data intervals: %w", err)
	}

	content := file.NewContent(header, outputDir)

	if err := content.Read(dcpFileStream); err != nil {
		return fmt.Errorf("error reading the data: %w", err)
	}

	return nil
}
