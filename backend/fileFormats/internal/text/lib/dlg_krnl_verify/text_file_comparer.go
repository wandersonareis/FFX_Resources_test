package verify

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/logger"
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type IComparer interface {
	CompareTranslatedTextParts(fromFile, toFile string) error
}

type Comparer struct {
	log zerolog.Logger
}

type FileInfo struct {
	Filename              string
	HeaderElements        int
	HeaderElementsNonZero int
}

func newPartComparer() IComparer {
	return &Comparer{
		log: logger.Get().With().Str("module", "dcp_parts_verify").Logger(),
	}
}

func (pc Comparer) CompareTranslatedTextParts(fromFile, toFile string) error {
	if err := pc.compare(fromFile, toFile); err != nil {
		return err
	}

	return nil
}

func (pc Comparer) compare(fromFile, toFile string) error {
	newExtractedPartData, err := os.ReadFile(fromFile)
	if err != nil {
		pc.log.Error().Err(err).Msgf("Error when reading extracted part: %s", common.GetFileName(fromFile))
		return fmt.Errorf("error when reading extracted part")
	}

	importedPartData, err := os.ReadFile(toFile)
	if err != nil {
		pc.log.Error().Err(err).Msgf("Error when reading imported part: %s", common.GetFileName(toFile))
		return fmt.Errorf("error when reading imported part")
	}

	if !bytes.Equal(newExtractedPartData, importedPartData) {
		pc.log.Error().Msgf("Extracted part is different from imported part: %s | Expected length: %d Got length: %d", common.GetFileName(fromFile), len(importedPartData), len(newExtractedPartData))
		return fmt.Errorf("extracted part is different from imported part")
	}

	return nil
}
