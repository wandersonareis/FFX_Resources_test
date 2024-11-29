package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/dcp/internal/splitter"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"fmt"

	"github.com/rs/zerolog"
)

type IFileValidator interface {
	Validate(filePath string, options interactions.DcpFileOptions) error
}

type FileValidator struct {
	fileSplitter  splitter.IDcpFileSpliter
	partsVerifier IPartsVerifier

	log zerolog.Logger
}

func newFileValidator() IFileValidator {
	return &FileValidator{
		fileSplitter:  splitter.NewDcpFileSpliter(),
		partsVerifier: newPartsVerifier(),

		log: logger.Get().With().Str("module", "dcp_file_validator").Logger(),
	}
}

func (fv *FileValidator) Validate(filePath string, options interactions.DcpFileOptions) error {
	tmpInfo, tmpDir := fv.createTemporaryFileInfo(filePath)
	defer tmpInfo.GetExtractLocation().DisposeTargetPath()

	if err := fv.fileSplitter.Split(tmpInfo); err != nil {
		fv.log.Error().
			Err(err).
			Str("file", filePath).
			Msg("error when splitting the file")

		return fmt.Errorf("error when splitting file")
	}

	if err := fv.partsVerifier.Verify(tmpDir, options); err != nil {
		fv.log.Error().
			Err(err).
			Str("file", filePath).
			Msg("error when verifying monted lockit file parts")

		return fmt.Errorf("error when verifying monted lockit file parts")
	}

	return nil
}

func (fv *FileValidator) createTemporaryFileInfo(filePath string) (interactions.IGameDataInfo, string) {
	tmpDir := common.NewTempProviderDev("", "").TempFilePath

	tmpInfo := interactions.NewGameDataInfo(filePath)
	tmpInfo.InitializeLocations(formatters.NewTxtFormatter())

	tmpInfo.GetExtractLocation().TargetPath = tmpDir

	return tmpInfo, tmpDir
}
