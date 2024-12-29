package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitFileParts"
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
	partsComparer IPartComparer
	partsDecoder  lockitFileParts.ILockitFilePartsDecoder
	fileSplitter  splitter.IFileSplitter
}

func newPartsVerifier() IPartsVerifier {
	return &partsVerifier{
		partsComparer: newPartComparer(),
		fileSplitter: splitter.NewLockitFileSplitter(),
	}
}

func (pv *partsVerifier) Verify(path string, options interactions.LockitFileOptions) error {
	partsList := components.NewEmptyList[lockitFileParts.LockitFileParts]()

	if err := components.GenerateGameFilePartsDev(partsList, path, lib.LOCKIT_FILE_PARTS_PATTERN, lockitFileParts.NewLockitFileParts); err != nil {
		return fmt.Errorf("error when finding lockit parts: %w", err)
	}

	if err := pv.EnsurePartsLength(partsList.GetLength(), options.PartsLength); err != nil {
		return fmt.Errorf("error when ensuring lockit parts exist: %w", err)
	}

	if err := pv.partsComparer.CompareGameDataBinaryParts(partsList); err != nil {
		return fmt.Errorf("error when comparing binary parts: %w", err)
	}

	tmpParts := pv.createExtractTemporaryPartsList(partsList, path)

	//pv.fileSplitter.DecoderPartsFiles(tmpParts)

	pv.partsDecoder = lockitFileParts.NewLockitFilePartsDecoder(tmpParts)
	pv.partsDecoder.DecodeFileParts()

	if err := pv.partsComparer.CompareTranslatedTextParts(tmpParts); err != nil {
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

func (pv *partsVerifier) createExtractTemporaryPartsList(partsList components.IList[lockitFileParts.LockitFileParts], tmpDir string) components.IList[lockitFileParts.LockitFileParts] {
	tmpParts := components.NewList[lockitFileParts.LockitFileParts](partsList.GetLength())

	setTemporaryDirectoryForPart := func(part lockitFileParts.LockitFileParts) {
		newPartFile := filepath.Join(tmpDir, common.GetFileName(part.Destination().Extract().Get().GetTargetFile()))

		part.Destination().Extract().Get().SetTargetFile(newPartFile)
		part.Destination().Extract().Get().SetTargetPath(tmpDir)

		tmpParts.Add(part)
	}

	partsList.ForEach(setTemporaryDirectoryForPart)

	return tmpParts
}
