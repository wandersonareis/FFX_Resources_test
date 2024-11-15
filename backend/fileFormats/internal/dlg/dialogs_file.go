package dlg

import (
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/dlg/internal"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"path/filepath"
	"slices"
)

type DialogsFile struct {
	*base.FormatsBase
}

func NewDialogs(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	return &DialogsFile{
		FormatsBase: base.NewFormatsBase(dataInfo),
	}
}

func (d DialogsFile) Extract() {
	if slices.Contains(d.GetGameData().ClonedItems, d.GetGameData().RelativeGameDataPath) {
		return
	}

	if err := internal.DialogsFileExtractor(d.GetFileInfo()); err != nil {
		d.Log.Error().Err(err).Interface("DialogFile", util.ErrorObject(d.GetFileInfo())).Msg("Error extracting dialog file")
		return
	}
}

func (d DialogsFile) Compress() {
	if err := internal.DialogsFileCompressor(d.GetFileInfo()); err != nil {
		d.Log.Error().Err(err).Interface("DialogFile", util.ErrorObject(d.GetFileInfo())).Msg("Error compressing dialog file")
		return
	}

	if d.GetGameData().ClonedItems != nil {
		for _, clone := range d.GetGameData().ClonedItems {
			cloneReimportPath := filepath.Join(d.GetImportLocation().TargetDirectory, clone)

			if err := util.DuplicateFile(d.GetImportLocation().TargetFile, cloneReimportPath); err != nil {
				d.Log.Error().Err(err).Str("File", clone).Str("TargetPath", cloneReimportPath).Msg("Error duplicating dialog file")
				continue
			}
		}
	}
}
