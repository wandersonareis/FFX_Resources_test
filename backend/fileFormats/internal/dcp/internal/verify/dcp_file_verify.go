package verify

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"fmt"

	"github.com/rs/zerolog"
)

type DcpFileVerify struct {
	fileValidator  IFileValidator
	segmentCounter ISegmentCounter

	log zerolog.Logger
}

func NewDcpFileVerify(dataInfo interactions.IGameDataInfo) *DcpFileVerify {
	return &DcpFileVerify{
		fileValidator:  newFileValidator(),
		segmentCounter: new(segmentCounter),

		log: logger.Get().With().Str("module", "dcp_file_verify").Logger(),
	}
}

func (lv *DcpFileVerify) VerifyExtract(dcpFileParts components.IList[parts.DcpFileParts], options interactions.DcpFileOptions) error {
	if dcpFileParts.GetLength() != options.PartsLength {
		lv.log.Error().
			Int("expected", options.PartsLength).
			Int("actual", dcpFileParts.GetLength()).
			Msg("Invalid number of split files")

		return fmt.Errorf("error when ensuring splited lockit parts")
	}

	if err := lv.segmentCounter.CountBinaryParts(dcpFileParts, options); err != nil {
		lv.log.Error().
			Err(err).
			Msg("Error when counting binary parts in splited files")

		return fmt.Errorf("error when counting binary line breaks in splited files")
	}

	if err := lv.segmentCounter.CountTextParts(dcpFileParts, options); err != nil {
		lv.log.Error().
			Err(err).
			Msg("Error when counting text segments in splited files")

		return fmt.Errorf("error when counting text segments in splited files")
	}

	return nil
}

func (lv *DcpFileVerify) VerifyCompress(dataInfo interactions.IGameDataInfo, options interactions.DcpFileOptions) error {
	lv.log.Info().
		Str("file", dataInfo.GetImportLocation().TargetFile).
		Msg("Verifying reimported macrodic file")

	if err := dataInfo.GetImportLocation().Validate(); err != nil {
		lv.log.Error().
			Err(err).
			Str("file", dataInfo.GetImportLocation().TargetFile).
			Msg("Error when validating reimported macrodic file")

		return fmt.Errorf("reimport file not exists: %w", err)
	}

	if err := lv.fileValidator.Validate(dataInfo.GetImportLocation().TargetFile, options); err != nil {
		lv.log.Error().
			Err(err).
			Str("file", dataInfo.GetImportLocation().TargetFile).
			Msg("Error when validating reimported macrodic file")

		return fmt.Errorf("error when validating reimported macrodic file: %w", err)
	}

	return nil
}
