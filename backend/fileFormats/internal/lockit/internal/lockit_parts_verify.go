package internal

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/util"
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
	fileSplitter  IFileSplitter
	worker        common.IWorker[LockitFileParts]
}

func newPartsVerifier() IPartsVerifier {
	worker := common.NewWorker[LockitFileParts]()
	return &partsVerifier{
		PartsComparer: newPartComparer(),
		fileSplitter:  NewLockitFileSplitter(),
		worker:        worker,
	}
}

func (pv *partsVerifier) Verify(path string, options interactions.LockitFileOptions) error {
	parts := []LockitFileParts{}

	if err := util.FindFileParts(&parts, path, LOCKIT_FILE_PARTS_PATTERN, NewLockitFileParts); err != nil {
		return fmt.Errorf("error when finding lockit parts: %w", err)
	}

	if err := pv.EnsurePartsLength(len(parts), options.PartsLength); err != nil {
		return fmt.Errorf("error when ensuring lockit parts exist: %w", err)
	}

	if err := pv.PartsComparer.CompareGameDataBinaryParts(parts); err != nil {
		return fmt.Errorf("error when comparing binary parts: %w", err)
	}

	tmpParts := pv.createExtractTemporaryPartsList(parts, path)

	pv.fileSplitter.DecoderPartsFiles(&tmpParts)

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

func (pv *partsVerifier) createExtractTemporaryPartsList(parts []LockitFileParts, tmpDir string) []LockitFileParts {
	tmpParts := make([]LockitFileParts, 0, len(parts))

	setTemporaryDirectoryForPart := func(index int, part LockitFileParts) {
		tmpPart := &part
		newPartFile := filepath.Join(tmpDir, part.GetExtractLocation().TargetFileName)

		tmpPart.GetExtractLocation().SetTargetFile(newPartFile)
		tmpPart.GetExtractLocation().SetTargetPath(tmpDir)

		tmpParts = append(tmpParts, *tmpPart)
	}

	pv.worker.VoidForEach(&parts, setTemporaryDirectoryForPart)

	return tmpParts
}
