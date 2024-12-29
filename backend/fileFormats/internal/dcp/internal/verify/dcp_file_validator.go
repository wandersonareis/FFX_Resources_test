package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/dcp/internal/splitter"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
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
	source, destination, tmpDir := fv.createTemporaryFileInfo(filePath)

	if err := fv.fileSplitter.Split(source, destination); err != nil {
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

func (fv *FileValidator) createTemporaryFileInfo(filePath string) (interfaces.ISource, locations.IDestination, string) {
	tmpDir := common.NewTempProviderDev("", "").TempFilePath

	gamePart := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()

	//tmpInfo := interactions.NewGameDataInfo(filePath, gamePart)
	source, err := locations.NewSource(filePath, gamePart)
	if err != nil {
		fv.log.Error().
			Err(err).
			Str("file", filePath).
			Msg("error when creating source")

		return nil, nil, ""
	}

	destination := locations.NewDestination()

	destination.InitializeLocations(source, formatters.NewTxtFormatter())

	destination.Extract().Get().SetTargetPath(tmpDir)

	return source, destination, tmpDir
}
