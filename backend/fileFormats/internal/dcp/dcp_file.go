package dcp

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/dcp/internal/joinner"
	"ffxresources/backend/fileFormats/internal/dcp/internal/lib"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/fileFormats/internal/dcp/internal/splitter"
	"ffxresources/backend/fileFormats/internal/dcp/internal/verify"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
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

func NewDcpFile(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	interactions := interactions.NewInteraction()

	dcpFileParts := components.NewEmptyList[parts.DcpFileParts]()
	//dataInfo.CreateRelativePath()

	destination.InitializeLocations(source, formatters.NewTxtFormatterDev())

	err := components.GenerateGameFilePartsDev(
		dcpFileParts,
		destination.Extract().Get().GetTargetPath(),
		lib.DCP_FILE_PARTS_PATTERN,
		parts.NewDcpFileParts)

	if err != nil {
		notifications.NotifyError(err)
		return nil
	}

	return &DcpFile{
		FormatsBase:   base.NewFormatsBaseDev(source, destination),
		dcpFileVerify: verify.NewDcpFileVerify(),
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
			Str("file", d.Destination().Import().Get().GetTargetFile()).
			Err(err).
			Msg("Failed to unpack DCP file")

		return fmt.Errorf("failed to unpack DCP file: %s", d.Source().Get().Name)
	}

	if d.PartsList.GetLength() != expectedDcpPartsLength {
		d.log.Error().
			Int("expected", expectedDcpPartsLength).
			Int("actual", d.PartsList.GetLength()).
			Msg("Invalid number of split files")

		return fmt.Errorf("invalid number of split files for %s: Expected: %d, Get: %d", d.Source().Get().Name, expectedDcpPartsLength, d.PartsList.GetLength())
	}

	errChan := make(chan error, d.PartsList.GetLength())

	go notifications.ProcessError(errChan, d.log)

	extractParts := func(_ int, part parts.DcpFileParts) {
		if err := part.Validate(); err != nil {
			d.log.Error().
				Str("part_file", part.Source().Get().Name).
				Err(err).
				Msg("Failed to validate file part")

			errChan <- fmt.Errorf("failed to validate file part: %s", part.Source().Get().Name)
		}

		if err := part.Extract(); err != nil {
			d.log.Error().
				Str("part_file", part.Source().Get().Name).
				Err(err).
				Msg("Failed to extract file part")

			errChan <- fmt.Errorf("failed to extract file part: %s", part.Source().Get().Name)
		}
	}

	d.PartsList.ParallelForEach(extractParts)

	defer close(errChan)

	d.log.Info().
		Str("file", d.Destination().Import().Get().GetTargetFile()).
		Msgf("Verifying monted macrodic file")

	if err := d.dcpFileVerify.VerifyExtract(d.PartsList, d.options); err != nil {
		d.log.Error().
			Str("file", d.Destination().Extract().Get().GetTargetFile()).
			Err(err).
			Msg("Failed to verify DCP file")

		return fmt.Errorf("failed to verify DCP file: %s", d.Source().Get().Name)
	}

	d.log.Info().
		Str("file", d.Destination().Extract().Get().GetTargetFile()).
		Msgf("System macrodic file extracted")

	return nil
}

func (d DcpFile) Compress() error {
	d.log.Info().
		Str("file", d.Destination().Translate().Get().GetTargetFile()).
		Msgf("Compressing macrodic file")

	expectedDcpPartsLength := d.options.PartsLength

	if err := d.ensureDcpPartsLength(expectedDcpPartsLength); err != nil {
		d.log.Error().
			Str("file", d.Destination().Import().Get().GetTargetFile()).
			Err(err).
			Msg("Failed to packing DCP file")

		return fmt.Errorf("failed to packing DCP file: %s", d.Source().Get().Name)
	}

	dcpTranslatedPartsTextPath := d.Destination().Translate().Get().GetTargetPath()

	if d.PartsList.GetLength() != expectedDcpPartsLength {
		d.log.Error().
			Int("expected", expectedDcpPartsLength).
			Int("actual", d.PartsList.GetLength()).
			Msg("Invalid number of split files")

		return fmt.Errorf("invalid number of split files for %s: Expected: %d, Get: %d",
			d.Source().Get().Name, expectedDcpPartsLength, d.PartsList.GetLength())
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

	targetReimportFile := d.Destination().Import().Get().GetTargetFile()

	if err := joinner.DcpFileJoiner(d.Source(), d.Destination(), d.PartsList, targetReimportFile); err != nil {
		d.log.Error().
			Err(err).
			Str("file", targetReimportFile).
			Msg("Error joining macrodic file")

		return fmt.Errorf("error joining macrodic file: %s", targetReimportFile)
	}

	d.log.Info().
		Str("file", d.Destination().Import().Get().GetTargetFile()).
		Msg("Verifying reimported macrodic file")

	if err := d.dcpFileVerify.VerifyCompress(d.Destination(), d.options); err != nil {
		d.log.Error().
			Err(err).
			Str("file", d.Destination().Import().Get().GetTargetFile()).
			Msg("Error verifying system macrodic file")

		return fmt.Errorf("error verifying system macrodic file: %s", d.Destination().Import().Get().GetTargetFile())
	}

	d.log.Info().
		Str("file", d.Destination().Import().Get().GetTargetFile()).
		Msgf("Macrodic file compressed")

	return nil
}

func (d *DcpFile) ensureDcpPartsLength(expected int) error {
	if d.PartsList.GetLength() != expected {
		if err := d.fileSplitter.Split(d.Source(), d.Destination()); err != nil {
			return err
		}

		newDcpFile := NewDcpFile(d.Source(), d.Destination()).(*DcpFile)
		d.SetFileInfo(newDcpFile.GetFileInfo())
		d.PartsList = newDcpFile.PartsList
	}

	return nil
}
