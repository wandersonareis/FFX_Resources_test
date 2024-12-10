package dcp

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/dcp/internal/joinner"
	"ffxresources/backend/fileFormats/internal/dcp/internal/lib"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/fileFormats/internal/dcp/internal/splitter"
	"ffxresources/backend/fileFormats/internal/dcp/internal/verify"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"ffxresources/backend/notifications"
	"fmt"

	"github.com/rs/zerolog"
)

type DcpFile struct {
	*base.FormatsBase

	dcpFileVerify *verify.DcpFileVerify
	PartsList     components.IList[parts.DcpFileParts]
	fileSplitter  splitter.IDcpFileSpliter
	options       interactions.DcpFileOptions
	log           zerolog.Logger
}

func NewDcpFile(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	interactions := interactions.NewInteraction()

	dcpFileParts := components.NewEmptyList[parts.DcpFileParts]()
	dataInfo.CreateRelativePath()

	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	err := components.GenerateGameFileParts(
		dcpFileParts,
		dataInfo.GetExtractLocation().TargetPath,
		lib.DCP_FILE_PARTS_PATTERN,
		parts.NewDcpFileParts)

	if err != nil {
		notifications.NotifyError(err)
		return nil
	}

	return &DcpFile{
		FormatsBase:   base.NewFormatsBase(dataInfo),
		dcpFileVerify: verify.NewDcpFileVerify(dataInfo),
		PartsList:     dcpFileParts,
		fileSplitter:  splitter.NewDcpFileSpliter(),
		options:       interactions.GamePartOptions.GetDcpFileOptions(),
		log:           logger.Get().With().Str("module", "dcp_file").Logger(),
	}
}

func (d *DcpFile) Extract() error {
	expectedDcpPartsLength := d.options.PartsLength

	if err := d.ensureDcpPartsLength(expectedDcpPartsLength); err != nil {
		d.log.Error().
			Str("file", d.GetFileInfo().GetImportLocation().TargetFile).
			Err(err).
			Msg("Failed to unpack DCP file")

		return fmt.Errorf("failed to unpack DCP file: %s", d.GetGameData().Name)
	}

	if d.PartsList.GetLength() != expectedDcpPartsLength {
		d.log.Error().
			Int("expected", expectedDcpPartsLength).
			Int("actual", d.PartsList.GetLength()).
			Msg("Invalid number of split files")

		return fmt.Errorf("invalid number of split files for %s: Expected: %d, Get: %d", d.GetGameData().Name, expectedDcpPartsLength, d.PartsList.GetLength())
	}

	errChan := make(chan error, d.PartsList.GetLength())

	go notifications.ProcessError(errChan, d.log)

	extractParts := func(_ int, part parts.DcpFileParts) {
		if err := part.Validate(); err != nil {
			d.log.Error().
				Str("part_file", part.GetGameData().Name).
				Err(err).
				Msg("Failed to validate file part")

			errChan <- fmt.Errorf("failed to validate file part: %s", part.GetGameData().Name)
		}

		if err := part.Extract(); err != nil {
			d.log.Error().
				Str("part_file", part.GetGameData().Name).
				Err(err).
				Msg("Failed to extract file part")

			errChan <- fmt.Errorf("failed to extract file part: %s", part.GetGameData().Name)
		}
	}

	d.PartsList.ParallelForEach(extractParts)

	defer close(errChan)

	d.log.Info().
		Str("file", d.GetFileInfo().GetImportLocation().TargetFile).
		Msgf("Verifying monted macrodic file")

	if err := d.dcpFileVerify.VerifyExtract(d.PartsList, d.options); err != nil {
		d.log.Error().
			Str("file", d.GetFileInfo().GetExtractLocation().TargetFile).
			Err(err).
			Msg("Failed to verify DCP file")

		return fmt.Errorf("failed to verify DCP file: %s", d.GetGameData().Name)
	}

	d.log.Info().
		Str("file", d.GetFileInfo().GetExtractLocation().TargetFile).
		Msgf("System macrodic file extracted")

	return nil
}

func (d DcpFile) Compress() error {
	d.log.Info().
		Str("file", d.GetFileInfo().GetTranslateLocation().TargetFile).
		Msgf("Compressing macrodic file")

	expectedDcpPartsLength := d.options.PartsLength

	if err := d.ensureDcpPartsLength(expectedDcpPartsLength); err != nil {
		d.log.Error().
			Str("file", d.GetFileInfo().GetImportLocation().TargetFile).
			Err(err).
			Msg("Failed to packing DCP file")

		return fmt.Errorf("failed to packing DCP file: %s", d.GetGameData().Name)
	}

	dcpTranslatedPartsTextPath := d.GetFileInfo().GetTranslateLocation().TargetPath

	if d.PartsList.GetLength() != expectedDcpPartsLength {
		d.log.Error().
			Int("expected", expectedDcpPartsLength).
			Int("actual", d.PartsList.GetLength()).
			Msg("Invalid number of split files")

		return fmt.Errorf("invalid number of split files for %s: Expected: %d, Get: %d",
			d.GetGameData().Name, expectedDcpPartsLength, d.PartsList.GetLength())
	}

	dcpXplitedTextFiles := components.NewList[string](expectedDcpPartsLength)

	if err := components.ListFilesByRegex(dcpXplitedTextFiles, dcpTranslatedPartsTextPath, lib.DCP_TXT_PARTS_PATTERN); err != nil {
		d.log.Error().
			Err(err).
			Str("Path", dcpTranslatedPartsTextPath).
			Msg("Error listing xplited text files")

		return fmt.Errorf("error listing xplited text files: %s", dcpTranslatedPartsTextPath)
	}

	if dcpXplitedTextFiles.GetLength() != expectedDcpPartsLength {
		d.log.Error().
			Int("expected", expectedDcpPartsLength).
			Int("actual", d.PartsList.GetLength()).
			Str("Path", dcpTranslatedPartsTextPath).
			Msg("Invalid number of split text files")

		return fmt.Errorf("invalid number of split text files for %s: Expected: %d, Get: %d", dcpTranslatedPartsTextPath, expectedDcpPartsLength, dcpXplitedTextFiles.GetLength())
	}

	compressor := func(_ int, part parts.DcpFileParts) {
		part.Compress()
	}

	d.PartsList.ParallelForEach(compressor)

	targetReimportFile := d.GetFileInfo().GetImportLocation().TargetFile

	if err := joinner.DcpFileJoiner(d.GetFileInfo(), d.PartsList, targetReimportFile); err != nil {
		d.log.Error().
			Err(err).
			Str("file", targetReimportFile).
			Msg("Error joining macrodic file")

		return fmt.Errorf("error joining macrodic file: %s", targetReimportFile)
	}

	d.log.Info().
		Str("file", d.GetFileInfo().GetImportLocation().TargetFile).
		Msg("Verifying reimported macrodic file")

	if err := d.dcpFileVerify.VerifyCompress(d.GetFileInfo(), d.options); err != nil {
		d.log.Error().
			Err(err).
			Str("file", d.GetFileInfo().GetImportLocation().TargetFile).
			Msg("Error verifying system macrodic file")

		return fmt.Errorf("error verifying system macrodic file: %s", d.GetFileInfo().GetImportLocation().TargetFile)
	}

	d.log.Info().
		Str("file", d.GetFileInfo().GetImportLocation().TargetFile).
		Msgf("Macrodic file compressed")

	return nil
}

func (d *DcpFile) ensureDcpPartsLength(expected int) error {
	if d.PartsList.GetLength() != expected {
		if err := d.fileSplitter.Split(d.GetFileInfo()); err != nil {
			return err
		}

		newDcpFile := NewDcpFile(d.GetFileInfo()).(*DcpFile)
		d.SetFileInfo(newDcpFile.GetFileInfo())
		d.PartsList = newDcpFile.PartsList
	}

	return nil
}
