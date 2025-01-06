package verify

import (
	"bytes"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/lockit/internal/splitter"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"fmt"
	"os"
)

type IFileValidator interface {
	Validate(filePath string, options interactions.LockitFileOptions) error
}

type FileValidator struct {
	fileSplitter  splitter.IFileSplitter
	partsVerifier IPartsVerifier
}

func newFileValidator() IFileValidator {
	return &FileValidator{
		fileSplitter:  splitter.NewLockitFileSplitter(),
		partsVerifier: newPartsVerifier(),
	}
}

func (fv *FileValidator) Validate(filePath string, options interactions.LockitFileOptions) error {
	lockitFileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error when reading file %s: | %w", filePath, err)
	}

	targetLineBreaksCount := bytes.Count(lockitFileData, []byte("\r\n"))

	if err := fv.ensureLineBreaksCount(targetLineBreaksCount, options.LineBreaksCount); err != nil {
		return fmt.Errorf("error when ensuring line breaks count: %w", err)
	}

	source, destination := fv.createTemporaryFileInfo(filePath)

	if err := fv.fileSplitter.FileSplitter(source, destination.Extract().Get(), options); err != nil {
		return fmt.Errorf("error when splitting file %s | %w", filePath, err)
	}

	if err := fv.partsVerifier.Verify(destination.Extract().Get().TargetPath, options); err != nil {
		return fmt.Errorf("error when verifying monted lockit file parts: %w", err)
	}

	return nil
}

func (fv *FileValidator) ensureLineBreaksCount(targetCount, expectedCount int) error {
	if targetCount != expectedCount {
		return fmt.Errorf("parts length is %d, expected %d", targetCount, expectedCount)
	}

	return nil
}

func (fv *FileValidator) createTemporaryFileInfo(filePath string) (interfaces.ISource, locations.IDestination) {
	//tmpProvider := common.NewTempProviderDev("", "")

	//tmpInfo := interactions.NewGameDataInfo(filePath)

	source, destination := util.CreateTemporaryFileInfo(filePath, formatters.NewTxtFormatter())

	//destination.InitializeLocations(source, formatters.NewTxtFormatter())

	//destination.Extract().Get().SetTargetPath(tmpDir)

	return source, destination
}
