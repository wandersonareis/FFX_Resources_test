package verify

import (
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"fmt"

	"github.com/rs/zerolog"
)

type DcpFileVerify struct {
	fileValidator  IFileValidator
	segmentCounter ISegmentCounter

	log    zerolog.Logger
}

func NewDcpFileVerify(dataInfo interactions.IGameDataInfo) *DcpFileVerify {
	return &DcpFileVerify{
		fileValidator:  newFileValidator(),
		segmentCounter: new(segmentCounter),
		log:            logger.Get().With().Str("module", "dcp_file_verify").Logger(),
	}
}

func (lv *DcpFileVerify) VerifyExtract(dcpFileParts *[]parts.DcpFileParts, options interactions.DcpFileOptions) error {
	if len(*dcpFileParts) != options.PartsLength {
		lv.log.Error().Msgf("Error when ensuring splited lockit parts: expected %d | got %d", options.PartsLength, len(*dcpFileParts))
		return fmt.Errorf("error when ensuring splited lockit parts")
	}

	if err := lv.segmentCounter.CountBinaryParts(dcpFileParts, options); err != nil {
		lv.log.Error().Err(err).Msgf("Error when counting binary parts in splited files: %s", err.Error())
		return fmt.Errorf("error when counting binary line breaks in splited files")
	}

	if err := lv.segmentCounter.CountTextParts(dcpFileParts, options); err != nil {
		lv.log.Error().Err(err).Msgf("Error when counting text segments in splited files: %s", err.Error())
		return fmt.Errorf("error when counting text segments in splited files")
	}

	return nil
}

func (lv *DcpFileVerify) VerifyCompress(dataInfo interactions.IGameDataInfo, options interactions.DcpFileOptions) error {
	lv.log.Info().Msgf("Verifying reimported macrodic file: %s", dataInfo.GetImportLocation().TargetFile)

	if err := dataInfo.GetImportLocation().Validate(); err != nil {
		lv.log.Error().Err(err).Msgf("Error when validating reimported macrodic file: %s", dataInfo.GetImportLocation().TargetFile)
		return fmt.Errorf("reimport file not exists: %w", err)
	}

	if err := lv.fileValidator.Validate(dataInfo.GetImportLocation().TargetFile, options); err != nil {
		lv.log.Error().Err(err).Msgf("Error when validating reimported macrodic file: %s", dataInfo.GetImportLocation().TargetFile)
		return fmt.Errorf("error when validating reimported macrodic file: %w", err)
	}

	return nil
}
