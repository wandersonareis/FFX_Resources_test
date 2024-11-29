package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/dcp/internal/splitter"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"fmt"
)

type IFileValidator interface {
	Validate(filePath string, options interactions.DcpFileOptions) error
}

type FileValidator struct {
	fileSplitter  splitter.IDcpFileSpliter
	partsVerifier IPartsVerifier
}

func newFileValidator() IFileValidator {
	return &FileValidator{
		fileSplitter:  splitter.NewDcpFileSpliter(),
		partsVerifier: newPartsVerifier(),
	}
}

func (fv *FileValidator) Validate(filePath string, options interactions.DcpFileOptions) error {
	tmpInfo, tmpDir := fv.createTemporaryFileInfo(filePath)
	defer tmpInfo.GetExtractLocation().DisposeTargetPath()

	if err := fv.fileSplitter.Split(tmpInfo); err != nil {
		return fmt.Errorf("error when splitting file %s | %w", filePath, err)
	}

	if err := fv.partsVerifier.Verify(tmpDir, options); err != nil {
		return fmt.Errorf("error when verifying monted lockit file parts: %w", err)
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
