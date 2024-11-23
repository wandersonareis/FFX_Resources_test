package verify

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/lockit/internal/splitter"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
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

	tmpInfo, tmpDir := fv.createTemporaryFileInfo(filePath)
	defer tmpInfo.GetExtractLocation().DisposeTargetPath()

	if err := fv.fileSplitter.FileSplitter(tmpInfo, options); err != nil {
		return fmt.Errorf("error when splitting file %s | %w", filePath, err)
	}

	if err := fv.partsVerifier.Verify(tmpDir, options); err != nil {
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

func (fv *FileValidator) createTemporaryFileInfo(filePath string) (interactions.IGameDataInfo, string) {
	tmpDir := common.NewTempProvider().ProvideTempDir()

	tmpInfo := interactions.NewGameDataInfo(filePath)
	tmpInfo.InitializeLocations(formatters.NewTxtFormatter())

	tmpInfo.GetExtractLocation().TargetPath = tmpDir
	defer tmpInfo.GetExtractLocation().DisposeTargetPath()

	return tmpInfo, tmpDir
}
