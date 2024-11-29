package dcp

import (
	"ffxresources/backend/common"
	"ffxresources/backend/events"
	"ffxresources/backend/fileFormats/internal/dcp/internal/joinner"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/fileFormats/internal/dcp/internal/splitter"
	"ffxresources/backend/fileFormats/internal/dcp/internal/verify"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"fmt"

	"github.com/rs/zerolog"
)

type DcpFile struct {
	*verify.DcpFileVerify

	fileSplitter splitter.IDcpFileSpliter
	options      interactions.DcpFileOptions
	PartsList    *[]parts.DcpFileParts
	log          zerolog.Logger
	worker       common.IWorker[parts.DcpFileParts]
}

func NewDcpFile(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	interactions := interactions.NewInteraction()

	dcpFileParts := &[]parts.DcpFileParts{}
	dataInfo.CreateRelativePath()

	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	if err := util.FindFileParts(dcpFileParts,
		dataInfo.GetExtractLocation().TargetPath,
		util.DCP_FILE_PARTS_PATTERN,
		parts.NewDcpFileParts); err != nil {
		events.NotifyError(err)
		return nil
	}

	return &DcpFile{
		DcpFileVerify: verify.NewDcpFileVerify(dataInfo),
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
		d.log.Error().Err(err).Interface("object", util.ErrorObject(d.GetFileInfo())).Msg("Error unpacking macrodic file")
		return
	}

	if len(*d.PartsList) != expectedDcpPartsLength {
		d.log.Error().Err(fmt.Errorf("invalid number of xplited files: %d", expectedDcpPartsLength)).Send()
		return
	}

	extractParts := func(i int, part parts.DcpFileParts) {
		if err := part.Validate(); err != nil {
			d.log.Error().Err(err).Msg("Error processing macrodic file parts")
			return
		}

		part.Extract()
	}

	d.worker.ParallelForEach(d.PartsList, extractParts)

	d.log.Info().Msgf("Verifying monted macrodic file: %s", d.GetFileInfo().GetImportLocation().TargetFile)

	if err := d.VerifyExtract(d.PartsList, d.options); err != nil {
		d.log.Error().Err(err).Interface("object", util.ErrorObject(d.GetFileInfo())).Msg("Error verifying system macrodic file")
		return
	}

	d.log.Info().Msgf("System macrodic file extracted: %s", d.GetFileInfo().GetGameData().Name)
}

func (d DcpFile) Compress() {
	d.log.Info().Msgf("Compressing macrodic file: %s", d.GetFileInfo().GetGameData().Name)

	expectedDcpPartsLength := d.options.PartsLength

	if err := d.ensureDcpPartsLength(expectedDcpPartsLength); err != nil {
		d.log.Error().Err(err).Interface("object", util.ErrorObject(d.GetFileInfo())).Msg("Error packing macrodic file")
		return
	}

	dcpTranslatedPartsTextPath := d.GetFileInfo().GetTranslateLocation().TargetPath

	d.log.Info().Msg("Ensuring xplited parts length...")

	if len(*d.PartsList) != expectedDcpPartsLength {
		d.log.Error().Err(fmt.Errorf("invalid number of xplited files: Expected parts: %d Got parts: %d", expectedDcpPartsLength, len(*d.PartsList))).Msg("Error packing macrodic file")
		return
	}

	dcpXplitedTextFiles := make([]string, 0, expectedDcpPartsLength)

	if err := common.ListFilesMatchingPattern(&dcpXplitedTextFiles, dcpTranslatedPartsTextPath, util.DCP_TXT_PARTS_PATTERN); err != nil {
		d.log.Error().Err(err).Str("Path", dcpTranslatedPartsTextPath).Send()
		return
	}

	if len(dcpXplitedTextFiles) != expectedDcpPartsLength {
		d.log.Error().Err(fmt.Errorf("invalid number of xplited text files: %d", len(dcpXplitedTextFiles))).Str("Path", dcpTranslatedPartsTextPath).Send()
		return
	}

	d.worker.ParallelForEach(d.PartsList, func(i int, part parts.DcpFileParts) {
		part.Compress()
	})

	targetReimportFile := d.GetFileInfo().GetImportLocation().TargetFile

	if err := joinner.DcpFileJoiner(d.GetFileInfo(), d.PartsList, targetReimportFile); err != nil {
		d.log.Error().Err(err).Interface("DcpParts", d.PartsList).Str("ImportTo", targetReimportFile).Msg("Error joining macrodic file")
		return
	}

	d.log.Info().Msgf("Verifying reimported macrodic file: %s", d.GetFileInfo().GetImportLocation().TargetFile)

	if err := d.VerifyCompress(d.GetFileInfo(), d.options); err != nil {
		d.log.Error().Err(err).Interface("object", util.ErrorObject(d.GetFileInfo())).Msg("Error verifying system macrodic file")
		return
	}

	d.log.Info().Msgf("Macrodic file compressed: %s", d.GetFileInfo().GetGameData().Name)
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
