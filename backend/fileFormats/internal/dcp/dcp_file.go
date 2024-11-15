package dcp

import (
	"ffxresources/backend/common"
	"ffxresources/backend/events"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/dcp/internal"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"fmt"
)

type DcpFile struct {
	*base.FormatsBase
	Parts *[]internal.DcpFileParts
}

func NewDcpFile(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	interactions := interactions.NewInteraction()

	gameVersionDcpPartsLength := interactions.GamePartOptions.GetGamePartOptions().DcpPartsLength

	parts := make([]internal.DcpFileParts, 0, gameVersionDcpPartsLength)

	dataInfo.CreateRelativePath()

	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	if err := util.FindFileParts(&parts,
		dataInfo.GetExtractLocation().TargetPath,
		util.DCP_FILE_PARTS_PATTERN,
		internal.NewDcpFileParts); err != nil {
		events.NotifyError(err)
		return nil
	}

	return &DcpFile{
		FormatsBase: base.NewFormatsBase(dataInfo),
		Parts:       &parts,
	}
}

func (d *DcpFile) Extract() {
	currentDcpPartsLength := len(*d.Parts)
	expectedDcpPartsLength := interactions.NewInteraction().GamePartOptions.GetGamePartOptions().DcpPartsLength

	if err := d.ensureDcpPartsLength(&currentDcpPartsLength, expectedDcpPartsLength); err != nil {
		d.Log.Error().Err(err).Interface("object", util.ErrorObject(d.GetFileInfo())).Msg("Error unpacking macrodic file")
		return
	}

	if currentDcpPartsLength != expectedDcpPartsLength {
		d.Log.Error().Err(fmt.Errorf("invalid number of xplited files: %d", expectedDcpPartsLength)).Interface("object", util.ErrorObject(d.GetFileInfo())).Send()

		return
	}

	worker := common.NewWorker[internal.DcpFileParts]()
	worker.ParallelForEach(*d.Parts, func(i int, extractor internal.DcpFileParts) {
		if err := extractor.Validate(); err != nil {
			d.Log.Error().Err(err).Interface("object", util.ErrorObject(extractor.GetFileInfo())).Msg("Error processing macrodic file parts")
			return
		}

		extractor.Extract()
	})
}

func (d DcpFile) Compress() {
	currentDcpPartsLength := len(*d.Parts)
	expectedDcpPartsLength := interactions.NewInteraction().GamePartOptions.GetGamePartOptions().DcpPartsLength

	if err := d.ensureDcpPartsLength(&currentDcpPartsLength, expectedDcpPartsLength); err != nil {
		d.Log.Error().Err(err).Interface("object", util.ErrorObject(d.GetFileInfo())).Msg("Error packing macrodic file")
		return
	}

	dcpTranslatedPartsTextPath := d.GetFileInfo().GetTranslateLocation().TargetPath

	if currentDcpPartsLength != expectedDcpPartsLength {
		d.Log.Error().Err(fmt.Errorf("invalid number of xplited files: %d", expectedDcpPartsLength)).Msg("Error packing macrodic file")
		return
	}

	dcpXplitedTextFiles := make([]string, 0, expectedDcpPartsLength)
	if err := common.ListFilesMatchingPattern(&dcpXplitedTextFiles, dcpTranslatedPartsTextPath, util.DCP_TXT_PARTS_PATTERN); err != nil {
		d.Log.Error().Err(err).Str("Path", dcpTranslatedPartsTextPath).Send()
		return
	}

	if len(dcpXplitedTextFiles) != expectedDcpPartsLength {
		d.Log.Error().Err(fmt.Errorf("invalid number of xplited text files: %d", len(dcpXplitedTextFiles))).Str("Path", dcpTranslatedPartsTextPath).Send()
		return
	}

	worker := common.NewWorker[internal.DcpFileParts]()
	worker.ParallelForEach(*d.Parts, func(i int, compressor internal.DcpFileParts) {
		compressor.Compress()
	})

	targetReimportFile := d.GetFileInfo().GetImportLocation().TargetFile

	if err := internal.DcpFileJoiner(d.GetFileInfo(), d.Parts, targetReimportFile); err != nil {
		d.Log.Error().Err(err).Interface("DcpParts", d.Parts).Str("ImportTo", targetReimportFile).Msg("Error joining macrodic file")
		return
	}
}

func (d *DcpFile) ensureDcpPartsLength(current *int, expected int) error {
	if *current != expected {
		if err := internal.DcpFileXpliter(d.GetFileInfo()); err != nil {
			return err
		}

		newDcpFile := NewDcpFile(d.GetFileInfo()).(*DcpFile)
		d.SetFileInfo(newDcpFile.GetFileInfo())
		d.Parts = newDcpFile.Parts

		*current = len(*d.Parts)
	}

	return nil
}
