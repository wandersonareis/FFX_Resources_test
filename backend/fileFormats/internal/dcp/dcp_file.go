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

	options *interactions.DcpFileOptions
	Parts   *[]internal.DcpFileParts
}

func NewDcpFile(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	interactions := interactions.NewInteraction()

	parts := []internal.DcpFileParts{}
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
		options:     interactions.GamePartOptions.GetDcpFileOptions(),
		Parts:       &parts,
	}
}

func (d *DcpFile) Extract() {
	expectedDcpPartsLength := d.options.PartsLength

	if err := d.ensureDcpPartsLength(expectedDcpPartsLength); err != nil {
		d.Log.Error().Err(err).Interface("object", util.ErrorObject(d.GetFileInfo())).Msg("Error unpacking macrodic file")
		return
	}

	if len(*d.Parts) != expectedDcpPartsLength {
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
	expectedDcpPartsLength := d.options.PartsLength

	if err := d.ensureDcpPartsLength(expectedDcpPartsLength); err != nil {
		d.Log.Error().Err(err).Interface("object", util.ErrorObject(d.GetFileInfo())).Msg("Error packing macrodic file")
		return
	}

	dcpTranslatedPartsTextPath := d.GetFileInfo().GetTranslateLocation().TargetPath

	if len(*d.Parts) != expectedDcpPartsLength {
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

func (d *DcpFile) ensureDcpPartsLength(expected int) error {
	if len(*d.Parts) != expected {
		if err := internal.DcpFileXpliter(d.GetFileInfo()); err != nil {
			return err
		}

		newDcpFile := NewDcpFile(d.GetFileInfo()).(*DcpFile)
		d.SetFileInfo(newDcpFile.GetFileInfo())
		d.Parts = newDcpFile.Parts
	}

	return nil
}

func (d *DcpFile) VerifyExtract() bool {
	extractedParts := []internal.DcpFileParts{}

	if err := util.FindFileParts(&extractedParts, d.GetExtractLocation().TargetPath, util.DCP_FILE_PARTS_PATTERN, internal.NewDcpFileParts); err != nil {
		return false
	}

	if len(extractedParts) != d.options.PartsLength {
		return false
	}

	partsLength := d.options.PartsLength

	if len(*d.Parts) != partsLength {
		return false
	}

	return true
}

func (d *DcpFile) VerifyCompress() error {
	return nil
}