package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/dcp/internal/lib"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/fileFormats/internal/dcp/internal/splitter"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog"
)

type IPartsVerifier interface {
	Verify(path string, options interactions.DcpFileOptions) error
	EnsurePartsLength(partsLength, expectedLength int) error
}

type partsVerifier struct {
	PartsComparer IPartComparer
	fileSplitter  splitter.IDcpFileSpliter
	log           zerolog.Logger
	worker        common.IWorker[parts.DcpFileParts]
}

func newPartsVerifier() IPartsVerifier {
	worker := common.NewWorker[parts.DcpFileParts]()
	return &partsVerifier{
		PartsComparer: newPartComparer(),
		fileSplitter:  new(splitter.DcpFileSpliter),
		log:           logger.Get().With().Str("module", "dcp_parts_verify").Logger(),
		worker:        worker,
	}
}

func (pv *partsVerifier) Verify(path string, options interactions.DcpFileOptions) error {
	partsList := &[]parts.DcpFileParts{}

	if err := util.FindFileParts(partsList, path, lib.DCP_FILE_PARTS_PATTERN, parts.NewDcpFileParts); err != nil {
		pv.log.Error().Err(err).Msgf("Error when finding lockit parts: %v", err)
		return fmt.Errorf("error when finding lockit parts")
	}

	if err := pv.EnsurePartsLength(len(*partsList), options.PartsLength); err != nil {
		pv.log.Error().Err(err).Msgf("Error when ensuring lockit parts length: %v", err)
		return fmt.Errorf("error when ensuring lockit parts length")
	}

	if err := pv.PartsComparer.CompareGameDataBinaryParts(partsList); err != nil {
		return fmt.Errorf("error when comparing binary parts")
	}

	tmpParts := pv.createExtractTemporaryPartsList(partsList, path)

	pv.worker.ParallelForEach(partsList, func(i int, part parts.DcpFileParts) {
		if err := part.Validate(); err != nil {
			pv.log.Error().Err(err).Msgf("Error processing macrodic file part: %s", part.GetGameData().Name)
			return
		}

		part.Extract()
	})

	if err := pv.PartsComparer.CompareTranslatedTextParts(tmpParts); err != nil {
		return fmt.Errorf("error when comparing text parts")
	}

	return nil
}

func (lc *partsVerifier) EnsurePartsLength(partsLength, expectedLength int) error {
	if partsLength != expectedLength {
		return fmt.Errorf("parts length is %d, expected %d", partsLength, expectedLength)
	}

	return nil
}

func (pv *partsVerifier) createExtractTemporaryPartsList(partsList *[]parts.DcpFileParts, tmpDir string) *[]parts.DcpFileParts {
	tmpParts := make([]parts.DcpFileParts, 0, len(*partsList))

	setTemporaryDirectoryForPart := func(index int, part parts.DcpFileParts) {
		tmpPart := &part
		newPartFile := filepath.Join(tmpDir, part.GetExtractLocation().TargetFileName)

		tmpPart.GetExtractLocation().SetTargetFile(newPartFile)
		tmpPart.GetExtractLocation().SetTargetPath(tmpDir)

		tmpParts = append(tmpParts, *tmpPart)
	}

	pv.worker.VoidForEach(partsList, setTemporaryDirectoryForPart)

	return &tmpParts
}
