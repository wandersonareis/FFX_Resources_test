package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"ffxresources/backend/fileFormats/internal/lockit/internal/parts"
	"ffxresources/backend/fileFormats/internal/lockit/internal/splitter"
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
	fileSplitter  splitter.IFileSplitter
	worker        common.IWorker[parts.LockitFileParts]
}

func newPartsVerifier() IPartsVerifier {
	worker := common.NewWorker[parts.LockitFileParts]()
	return &partsVerifier{
		PartsComparer: newPartComparer(),
		fileSplitter:  splitter.NewLockitFileSplitter(),
		worker:        worker,
	}
}

func (pv *partsVerifier) Verify(path string, options interactions.LockitFileOptions) error {
	partsList := &[]parts.LockitFileParts{}

	if err := util.FindFileParts(partsList, path, lib.LOCKIT_FILE_PARTS_PATTERN, parts.NewLockitFileParts); err != nil {
		return fmt.Errorf("error when finding lockit parts: %w", err)
	}

	if err := pv.EnsurePartsLength(len(*partsList), options.PartsLength); err != nil {
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

func (pv *partsVerifier) createExtractTemporaryPartsList(partsList *[]parts.LockitFileParts, tmpDir string) *[]parts.LockitFileParts {
	tmpParts := make([]parts.LockitFileParts, 0, len(*partsList))

	setTemporaryDirectoryForPart := func(index int, part parts.LockitFileParts) {
		tmpPart := &part
		newPartFile := filepath.Join(tmpDir, part.GetExtractLocation().TargetFileName)

		tmpPart.GetExtractLocation().SetTargetFile(newPartFile)
		tmpPart.GetExtractLocation().SetTargetPath(tmpDir)

		tmpParts = append(tmpParts, *tmpPart)
	}

	pv.worker.VoidForEach(partsList, setTemporaryDirectoryForPart)

	return &tmpParts
}
