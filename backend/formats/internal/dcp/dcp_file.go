package dcp

import (
	"ffxresources/backend/common"
	"ffxresources/backend/events"
	"ffxresources/backend/formats/internal/dcp/internal"
	"ffxresources/backend/formats/lib"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"fmt"
)

type DcpFile struct {
	dataInfo *interactions.GameDataInfo
	Parts    *[]internal.DcpFileParts
}

func NewDcpFile(dataInfo *interactions.GameDataInfo) *DcpFile {
	interactions := interactions.NewInteraction()

	gameVersionDcpPartsLength := interactions.GamePartOptions.GetGamePartOptions().DcpPartsLength

	parts := make([]internal.DcpFileParts, 0, gameVersionDcpPartsLength)

	gameFilesPath := interactions.GameLocation.TargetDirectory

	relative := common.GetDifferencePath(dataInfo.GameData.FullFilePath, gameFilesPath)

	dataInfo.GameData.RelativeGameDataPath = relative

	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	if err := lib.FindFileParts(&parts,
		dataInfo.ExtractLocation.TargetPath,
		lib.DCP_FILE_PARTS_PATTERN,
		internal.NewDcpFileParts); err != nil {
		events.NotifyError(err)
		return nil
	}

	return &DcpFile{
		dataInfo: dataInfo,
		Parts:    &parts,
	}
}

func (d DcpFile) GetFileInfo() *interactions.GameDataInfo {
	return d.dataInfo
}

func (d *DcpFile) Extract() {
	currentDcpPartsLength := len(*d.Parts)
	expectedDcpPartsLength := interactions.NewInteraction().GamePartOptions.GetGamePartOptions().DcpPartsLength

	if err := d.ensureDcpPartsLength(&currentDcpPartsLength, expectedDcpPartsLength); err != nil {
		events.NotifyError(err)
		return
	}

	if currentDcpPartsLength != expectedDcpPartsLength {
		events.NotifyError(fmt.Errorf("invalid number of xplited files: %d", expectedDcpPartsLength))
		return
	}

	worker := common.NewWorker[internal.DcpFileParts]()
	worker.ParallelForEach(*d.Parts, func(i int, extractor internal.DcpFileParts) {
		if err := extractor.Validate(); err != nil {
			events.NotifyError(err)
			return
		}

		extractor.Extract()
	})
}

func (d DcpFile) Compress() {
	currentDcpPartsLength := len(*d.Parts)
	expectedDcpPartsLength := interactions.NewInteraction().GamePartOptions.GetGamePartOptions().DcpPartsLength

	if err := d.ensureDcpPartsLength(&currentDcpPartsLength, expectedDcpPartsLength); err != nil {
		events.NotifyError(err)
		return
	}

	dcpTranslatedPartsTextPath := d.dataInfo.TranslateLocation.TargetPath

	if currentDcpPartsLength != expectedDcpPartsLength {
		events.NotifyError(fmt.Errorf("invalid number of xplited files"))
		return
	}

	dcpXplitedTextFiles := make([]string, 0, expectedDcpPartsLength)
	if err := common.ListFilesMatchingPattern(&dcpXplitedTextFiles, dcpTranslatedPartsTextPath, lib.DCP_TXT_PARTS_PATTERN); err != nil {
		events.NotifyError(err)
		return
	}

	if len(dcpXplitedTextFiles) != expectedDcpPartsLength {
		events.NotifyError(fmt.Errorf("invalid number of xplited text files: %d", len(dcpXplitedTextFiles)))
		return
	}

	worker := common.NewWorker[internal.DcpFileParts]()
	worker.ParallelForEach(*d.Parts, func(i int, compressor internal.DcpFileParts) {
		compressor.Compress()
	})

	targetReimportFile := d.dataInfo.ImportLocation.TargetFile

	if err := internal.DcpFileJoiner(d.dataInfo, d.Parts, targetReimportFile); err != nil {
		events.NotifyError(err)
		return
	}
}

func (d *DcpFile) ensureDcpPartsLength(current *int, expected int) error {
	if *current != expected {
		if err := internal.DcpFileXpliter(d.GetFileInfo()); err != nil {
			return err
		}

		newDcpFile := NewDcpFile(d.GetFileInfo())
		d.dataInfo = newDcpFile.GetFileInfo()
		d.Parts = newDcpFile.Parts

		*current = len(*d.Parts)
	}

	return nil
}
