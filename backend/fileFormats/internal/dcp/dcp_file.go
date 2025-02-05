package dcp

import (
	"ffxresources/backend/core"
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
)

type DcpFile struct {
	*base.FormatsBase

	dcpFileVerify *verify.DcpFileVerify
	PartsList     components.IList[parts.DcpFileParts]
	fileSplitter  splitter.IDcpFileSpliter
	options       core.IDcpFileOptions

	log logger.ILoggerHandler
}

func NewDcpFile(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	dcpFileParts := components.NewEmptyList[parts.DcpFileParts]()

	destination.InitializeLocations(source, formatters.NewTxtFormatter())

	err := components.PopulateGameFilePartsList(
		dcpFileParts,
		destination.Extract().Get().GetTargetPath(),
		lib.DCP_FILE_PARTS_PATTERN,
		parts.NewDcpFileParts)

	if err != nil {
		notifications.NotifyError(err)
		return nil
	}

	logger := &logger.LogHandler{
		Logger: logger.Get().With().Str("module", "dcp_file").Logger(),
	}

	return &DcpFile{
		FormatsBase: base.NewFormatsBase(source, destination),
		dcpFileVerify: verify.NewDcpFileVerify(logger),
		PartsList:     dcpFileParts,
		fileSplitter:  splitter.NewDcpFileSpliter(logger),
		options:       core.NewDcpFileOptions(interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber()),
		log: logger,
	}
}

func (d *DcpFile) Extract() error {
	expectedDcpPartsLength := d.options.GetPartsLength()

	if err := d.ensureDcpPartsLength(expectedDcpPartsLength); err != nil {
		d.log.LogError(err, "failed to unpack DCP file: %s", d.Source().Get().Name)

		return fmt.Errorf("failed to unpack DCP file: %s", d.Source().Get().Name)
	}

	if d.PartsList.GetLength() != expectedDcpPartsLength {
		d.log.LogError(nil, "invalid number of split files: Expected: %d, Get: %d", expectedDcpPartsLength, d.PartsList.GetLength())
		return fmt.Errorf("invalid number of split files for %s: Expected: %d, Get: %d", d.Source().Get().Name, expectedDcpPartsLength, d.PartsList.GetLength())
	}

	errChan := make(chan error, d.PartsList.GetLength())
	defer close(errChan)

	extractParts := func(_ int, part parts.DcpFileParts) {
		if err := part.Validate(); err != nil {
			d.log.LogError(err, "failed to validate file part: %s", part.Source().Get().Name)

			errChan <- fmt.Errorf("failed to validate file part: %s", part.Source().Get().Name)
		}

		if err := part.Extract(); err != nil {
			d.log.LogError(err, "failed to extract file part: %s", part.Source().Get().Name)

			errChan <- fmt.Errorf("failed to extract file part: %s", part.Source().Get().Name)
		}
	}

	d.PartsList.ParallelForEach(extractParts)

	if err := <-errChan; err != nil {
		d.log.LogError(err, "error extracting DCP file: %s", d.Source().Get().Name)
		return fmt.Errorf("error extracting DCP file: %s", d.Source().Get().Name)
	}

	d.log.LogInfo("Verifying monted macrodic file: %s", d.Destination().Extract().Get().GetTargetFile())

	if err := d.dcpFileVerify.VerifyExtract(d.PartsList, d.options); err != nil {
		d.log.LogError(err, "failed to verify DCP file: %s", d.Source().Get().Name)

		return fmt.Errorf("failed to verify DCP file: %s", d.Source().Get().Name)
	}

	d.log.LogInfo("System macrodic file extracted: %s", d.Source().Get().Name)

	return nil
}

func (d *DcpFile) Compress() error {
	translateDestination := d.Destination().Translate().Get()
	importDestination := d.Destination().Import().Get()

	d.log.LogInfo("Compressing DCP file parts...")

	expectedDcpPartsLength := d.options.GetPartsLength()

	if err := d.ensureDcpPartsLength(expectedDcpPartsLength); err != nil {
		d.log.LogError(err, "failed to packing DCP file: %s", d.Source().Get().Name)

		return fmt.Errorf("failed to packing DCP file: %s", d.Source().Get().Name)
	}

	dcpTranslatedPartsTextPath := translateDestination.GetTargetPath()

	if d.PartsList.GetLength() != expectedDcpPartsLength {
		d.log.LogError(nil, "invalid number of split files: Expected: %d, Get: %d", expectedDcpPartsLength, d.PartsList.GetLength())

		return fmt.Errorf("invalid number of split files for %s: Expected: %d, Get: %d",
			d.Source().Get().Name, expectedDcpPartsLength, d.PartsList.GetLength())
	}

	dcpSplitedTextFiles := components.NewList[string](expectedDcpPartsLength)

	if err := components.ListFilesByRegex(dcpSplitedTextFiles, dcpTranslatedPartsTextPath, lib.DCP_TXT_PARTS_PATTERN); err != nil {
		d.log.LogError(err, "error listing xplited text files: %s", dcpTranslatedPartsTextPath)

		return fmt.Errorf("error listing xplited text files: %s", dcpTranslatedPartsTextPath)
	}

	if dcpSplitedTextFiles.GetLength() != expectedDcpPartsLength {
		d.log.LogError(nil, "invalid number of split text files: Expected: %d, Get: %d, Path: %s", expectedDcpPartsLength, dcpSplitedTextFiles.GetLength(), dcpTranslatedPartsTextPath)

		return fmt.Errorf("invalid number of split text files for %s: Expected: %d, Get: %d",
			dcpTranslatedPartsTextPath, expectedDcpPartsLength, dcpSplitedTextFiles.GetLength())
	}

	compressor := func(_ int, part parts.DcpFileParts) {
		if err := part.Compress(); err != nil {
			d.log.LogError(err, "failed to compress file part: %s", part.Source().Get().Name)
		}
	}

	d.PartsList.ParallelForEach(compressor)

	outputFile := importDestination.GetTargetFile()

	if err := joinner.DcpFileJoiner(d.Source(), d.Destination(), d.PartsList, outputFile); err != nil {
		d.log.LogError(err, "error joining macrodic file: %s", outputFile)

		return fmt.Errorf("error joining macrodic file: %s", outputFile)
	}

	d.log.LogInfo("Verifying reimported macrodic file: %s", outputFile)

	if err := d.dcpFileVerify.VerifyCompress(d.Destination(), formatters.NewTxtFormatter(), d.options); err != nil {
		d.log.LogError(err, "error verifying system macrodic file: %s", outputFile)

		return fmt.Errorf("error verifying system macrodic file: %s", outputFile)
	}

	d.log.LogInfo("Macrodic file compressed: %s", d.Source().Get().Name)

	return nil
}

func (d *DcpFile) ensureDcpPartsLength(expected int) error {
	if d.PartsList.GetLength() != expected {
		if err := d.fileSplitter.Split(d.Source(), d.Destination(), d.options); err != nil {
			return err
		}

		newDcpFile := NewDcpFile(d.Source(), d.Destination()).(*DcpFile)
		d.SetFileInfo(newDcpFile.GetFileInfo())
		d.PartsList = newDcpFile.PartsList
	}

	return nil
}
