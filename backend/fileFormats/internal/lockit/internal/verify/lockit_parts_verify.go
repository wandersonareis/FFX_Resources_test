package verify

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"ffxresources/backend/fileFormats/internal/lockit/internal/parts"
	"ffxresources/backend/fileFormats/internal/lockit/internal/splitter"
	"ffxresources/backend/interactions"
	"fmt"
	"path/filepath"
)

type IPartsVerifier interface {
	Verify(path string, options interactions.LockitFileOptions) error
	EnsurePartsLength(partsLength, expectedLength int) error
}

type partsVerifier struct {
	PartsComparer IPartComparer
	fileSplitter  splitter.IFileSplitter
}

func newPartsVerifier() IPartsVerifier {
	return &partsVerifier{
		PartsComparer: newPartComparer(),
		fileSplitter:  splitter.NewLockitFileSplitter(),
	}
}

func (pv *partsVerifier) Verify(path string, options interactions.LockitFileOptions) error {
	partsList := components.NewEmptyList[parts.LockitFileParts]()

	if err := components.GenerateGameFileParts(partsList, path, lib.LOCKIT_FILE_PARTS_PATTERN, parts.NewLockitFileParts); err != nil {
		return fmt.Errorf("error when finding lockit parts: %w", err)
	}

	if err := pv.EnsurePartsLength(partsList.GetLength(), options.PartsLength); err != nil {
		return fmt.Errorf("error when ensuring lockit parts exist: %w", err)
	}

	if err := pv.PartsComparer.CompareGameDataBinaryParts(partsList); err != nil {
		return fmt.Errorf("error when comparing binary parts: %w", err)
	}

	tmpParts := pv.createExtractTemporaryPartsList(partsList, path)

	pv.fileSplitter.DecoderPartsFiles(tmpParts)

	if err := pv.PartsComparer.CompareTranslatedTextParts(tmpParts); err != nil {
		return fmt.Errorf("error when comparing text parts: %w", err)
	}

	return nil
}

func (lc *partsVerifier) EnsurePartsLength(partsLength, expectedLength int) error {
	if partsLength != expectedLength {
		return fmt.Errorf("parts length is %d, expected %d", partsLength, expectedLength)
	}

	return nil
}

func (pv *partsVerifier) createExtractTemporaryPartsList(partsList components.IList[parts.LockitFileParts], tmpDir string) components.IList[parts.LockitFileParts] {
	tmpParts := components.NewList[parts.LockitFileParts](partsList.GetLength())

	setTemporaryDirectoryForPart := func(part parts.LockitFileParts) {
		newPartFile := filepath.Join(tmpDir, part.GetExtractLocation().TargetFileName)

		part.GetExtractLocation().SetTargetFile(newPartFile)
		part.GetExtractLocation().SetTargetPath(tmpDir)

		tmpParts.Add(part)
	}

	partsList.ForEach(setTemporaryDirectoryForPart)

	return tmpParts
}
