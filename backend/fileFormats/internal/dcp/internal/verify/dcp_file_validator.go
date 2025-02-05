package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/dcp/internal/splitter"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
)

type IFileValidator interface {
	Validate(filePath string, formatter interfaces.ITextFormatter, options core.IDcpFileOptions) error
}

type FileValidator struct {
	fileSplitter  splitter.IDcpFileSpliter
	partsVerifier IPartsVerifier

	log logger.ILoggerHandler
}

func newFileValidator(logger logger.ILoggerHandler) IFileValidator {
	return &FileValidator{
		fileSplitter:  splitter.NewDcpFileSpliter(logger),
		partsVerifier: newPartsVerifier(logger),

		log: logger,
	}
}

func (fv *FileValidator) Validate(filePath string, formatter interfaces.ITextFormatter, fileOptions core.IDcpFileOptions) error {
	source, destination, tmpDir := fv.createTemporaryFileInfo(filePath)

	if err := fv.fileSplitter.Split(source, destination, fileOptions); err != nil {
		fv.log.LogError(err, "error when splitting the file: %s", filePath)

		return fmt.Errorf("error when splitting file")
	}

	if err := fv.partsVerifier.Verify(tmpDir, formatter, fileOptions); err != nil {
		fv.log.LogError(err, "error when verifying monted lockit file parts")

		return fmt.Errorf("error when verifying monted lockit file parts")
	}

	return nil
}

func (fv *FileValidator) createTemporaryFileInfo(filePath string) (interfaces.ISource, locations.IDestination, string) {
	tmpDir := common.NewTempProvider("", "").TempFilePath

	//tmpInfo := interactions.NewGameDataInfo(filePath, gamePart)
	source, err := locations.NewSource(filePath)
	if err != nil {
		fv.log.LogError(err, "error when creating source")

		return nil, nil, ""
	}

	destination := locations.NewDestination()

	destination.InitializeLocations(source, formatters.NewTxtFormatter())

	destination.Extract().Get().SetTargetPath(tmpDir)

	return source, destination, tmpDir
}
