package dcp

import (
	"ffxresources/backend/common"
	"ffxresources/backend/events"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/dcp/internal/joinner"
	"ffxresources/backend/fileFormats/internal/dcp/internal/lib"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/fileFormats/internal/dcp/internal/splitter"
	"ffxresources/backend/fileFormats/internal/dcp/internal/verify"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"

	"github.com/rs/zerolog"
)

type DcpFile struct {
	*base.FormatsBase

	dcpFileVerify *verify.DcpFileVerify
	PartsList     *[]parts.DcpFileParts
	fileSplitter  splitter.IDcpFileSpliter
	options       interactions.DcpFileOptions
	log           zerolog.Logger
	worker        common.IWorker[parts.DcpFileParts]
}

func NewDcpFile(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	interactions := interactions.NewInteraction()

	dcpFileParts := &[]parts.DcpFileParts{}
	dataInfo.CreateRelativePath()

	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	if err := util.FindFileParts(dcpFileParts,
		dataInfo.GetExtractLocation().TargetPath,
		lib.DCP_FILE_PARTS_PATTERN,
		parts.NewDcpFileParts); err != nil {
		events.NotifyError(err)
		return nil
	}

	return &DcpFile{
		FormatsBase:   base.NewFormatsBase(dataInfo),
		dcpFileVerify: verify.NewDcpFileVerify(dataInfo),
		PartsList:     dcpFileParts,
		fileSplitter:  splitter.NewDcpFileSpliter(),
		options:       interactions.GamePartOptions.GetDcpFileOptions(),
		log:           logger.Get().With().Str("module", "dcp_file").Logger(),
		worker:        common.NewWorker[parts.DcpFileParts](),
	}
}

func (d *DcpFile) Extract() {
	expectedDcpPartsLength := d.options.PartsLength

	if err := d.ensureDcpPartsLength(expectedDcpPartsLength); err != nil {
		d.log.Error().
			Str("file", d.GetFileInfo().GetImportLocation().TargetFile).
			Err(err).
			Msg("Failed to unpack DCP file")
		return
	}

	if len(*d.PartsList) != expectedDcpPartsLength {
		d.log.Error().
			Int("expected", expectedDcpPartsLength).
			Int("actual", len(*d.PartsList)).
			Msg("Invalid number of split files")
		return
	}

	extractParts := func(i int, part parts.DcpFileParts) {
		if err := part.Validate(); err != nil {
			d.log.Error().
				Str("part_file", part.GetGameData().Name).
				Err(err).
				Msg("Failed to validate file part")
			return
		}

		part.Extract()
	}

	d.worker.ParallelForEach(d.PartsList, extractParts)

	d.log.Info().
		Str("file", d.GetFileInfo().GetImportLocation().TargetFile).
		Msgf("Verifying monted macrodic file")

	if err := d.dcpFileVerify.VerifyExtract(d.PartsList, d.options); err != nil {
		d.log.Error().
			Str("file", d.GetFileInfo().GetExtractLocation().TargetFile).
			Err(err).
			Msg("Failed to verify DCP file")
		return
	}

	d.log.Info().
		Str("file", d.GetFileInfo().GetExtractLocation().TargetFile).
		Msgf("System macrodic file extracted")
}

func (d DcpFile) Compress() {
	d.log.Info().
		Str("file", d.GetFileInfo().GetTranslateLocation().TargetFile).
		Msgf("Compressing macrodic file")

	expectedDcpPartsLength := d.options.PartsLength

	if err := d.ensureDcpPartsLength(expectedDcpPartsLength); err != nil {
		d.log.Error().
			Str("file", d.GetFileInfo().GetImportLocation().TargetFile).
			Err(err).
			Msg("Failed to packing DCP file")
		return
	}

	dcpTranslatedPartsTextPath := d.GetFileInfo().GetTranslateLocation().TargetPath

	if len(*d.PartsList) != expectedDcpPartsLength {
		d.log.Error().
			Int("expected", expectedDcpPartsLength).
			Int("actual", len(*d.PartsList)).
			Msg("Invalid number of split files")
		return
	}

	dcpXplitedTextFiles := make([]string, 0, expectedDcpPartsLength)

	if err := common.ListFilesMatchingPattern(&dcpXplitedTextFiles, dcpTranslatedPartsTextPath, lib.DCP_TXT_PARTS_PATTERN); err != nil {
		d.log.Error().
			Err(err).
			Str("Path", dcpTranslatedPartsTextPath).
			Msg("Error listing xplited text files")
		return
	}

	if len(dcpXplitedTextFiles) != expectedDcpPartsLength {
		d.log.Error().
			Int("expected", expectedDcpPartsLength).
			Int("actual", len(*d.PartsList)).
			Str("Path", dcpTranslatedPartsTextPath).
			Msg("Invalid number of split text files")
		return
	}

	d.worker.ParallelForEach(d.PartsList, func(i int, part parts.DcpFileParts) {
		part.Compress()
	})

	targetReimportFile := d.GetFileInfo().GetImportLocation().TargetFile

	if err := joinner.DcpFileJoiner(d.GetFileInfo(), d.PartsList, targetReimportFile); err != nil {
		d.log.Error().
			Err(err).
			Str("file", targetReimportFile).
			Msg("Error joining macrodic file")
		return
	}

	d.log.Info().
		Str("file", d.GetFileInfo().GetImportLocation().TargetFile).
		Msg("Verifying reimported macrodic file")

	if err := d.dcpFileVerify.VerifyCompress(d.GetFileInfo(), d.options); err != nil {
		d.log.Error().
			Err(err).
			Str("file", d.GetFileInfo().GetImportLocation().TargetFile).
			Msg("Error verifying system macrodic file")
		return
	}

	d.log.Info().
		Str("file", d.GetFileInfo().GetImportLocation().TargetFile).
		Msgf("Macrodic file compressed")
}

func (d *DcpFile) ensureDcpPartsLength(expected int) error {
	if len(*d.PartsList) != expected {
		if err := d.fileSplitter.Split(d.GetFileInfo()); err != nil {
			return err
		}

		newDcpFile := NewDcpFile(d.GetFileInfo()).(*DcpFile)
		d.SetFileInfo(newDcpFile.GetFileInfo())
		d.PartsList = newDcpFile.PartsList
	}

	return nil
}
