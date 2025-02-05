package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/dcp/internal/lib"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/fileFormats/internal/dcp/internal/splitter"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
	"path/filepath"
)

type IPartsVerifier interface {
	Verify(path string, formatter interfaces.ITextFormatter, options core.IDcpFileOptions) error
	EnsurePartsLength(partsLength, expectedLength int) error
}

type partsVerifier struct {
	PartsComparer IPartComparer
	fileSplitter  splitter.IDcpFileSpliter

	log logger.ILoggerHandler
}

func newPartsVerifier(logger logger.ILoggerHandler) IPartsVerifier {
	return &partsVerifier{
		PartsComparer: newPartComparer(),
		fileSplitter:  new(splitter.DcpFileSpliter),

		log: logger,
	}
}

func (pv *partsVerifier) Verify(path string, formatter interfaces.ITextFormatter, options core.IDcpFileOptions) error {
	partsList := components.NewEmptyList[parts.DcpFileParts]()

	if err := util.FindFileParts(partsList, path, lib.DCP_FILE_PARTS_PATTERN, formatter, parts.NewDcpFileParts); err != nil {
		pv.log.LogError(err, "error when finding lockit parts: %s", path)

		return fmt.Errorf("error when finding lockit parts")
	}

	if err := pv.EnsurePartsLength(partsList.GetLength(), options.GetPartsLength()); err != nil {
		pv.log.LogError(err, "error when ensuring lockit parts length: Expected parts: %d, Found parts: %d", options.GetPartsLength(), partsList.GetLength())

		return fmt.Errorf("error when ensuring lockit parts length")
	}

	if err := pv.PartsComparer.CompareGameDataBinaryParts(partsList); err != nil {
		pv.log.LogError(err, "error when comparing binary parts")

		return fmt.Errorf("error when comparing binary parts")
	}

	tmpParts := pv.createExtractTemporaryPartsList(partsList, path)

	extractorFunc := func(index int, part parts.DcpFileParts) {
		if err := part.Validate(); err != nil {
			pv.log.LogError(err, "error when validating lockit part: %s", part.Source().Get().Path)

			return
		}

		part.Extract()
	}

	partsList.ParallelForEach(extractorFunc)

	if err := pv.PartsComparer.CompareTranslatedTextParts(tmpParts); err != nil {
		pv.log.LogError(err, "error when comparing text parts")

		return fmt.Errorf("error when comparing text parts")
	}

	return nil
}

func (lc *partsVerifier) EnsurePartsLength(partsLength, expectedLength int) error {
	if partsLength != expectedLength {
		return fmt.Errorf("parts length is different from expected")
	}

	return nil
}

func (pv *partsVerifier) createExtractTemporaryPartsList(partsList components.IList[parts.DcpFileParts], tmpDir string) components.IList[parts.DcpFileParts] {
	tmpPartsList := components.NewList[parts.DcpFileParts](partsList.GetLength())

	setTemporaryDirectoryForPart := func(part parts.DcpFileParts) {
		tmpPart := &part
		newPartFile := filepath.Join(tmpDir, common.GetFileName(part.Destination().Extract().Get().GetTargetFile()))

		tmpPart.Destination().Extract().Get().SetTargetFile(newPartFile)
		tmpPart.Destination().Extract().Get().SetTargetPath(tmpDir)

		tmpPartsList.Add(*tmpPart)
	}

	partsList.ForEach(setTemporaryDirectoryForPart)

	return tmpPartsList
}
