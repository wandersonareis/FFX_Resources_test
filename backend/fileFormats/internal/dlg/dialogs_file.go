package dlg

import (
	"ffxresources/backend/fileFormats/internal/dlg/internal"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"path/filepath"
	"slices"
)

type DialogsFile struct {
	*util.DlgKrnlVerify
}

func NewDialogs(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	return &DialogsFile{
		DlgKrnlVerify: util.NewDlgKrnlVerify(dataInfo),
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

	if err := d.VerifyExtract(d.GetExtractLocation()); err != nil {
		d.Log.Error().Err(err).Interface("DialogFile", util.ErrorObject(d.GetFileInfo())).Msg("Error verifying dialog file")
		return
	}

	d.Log.Info().Msgf("Dialog file extracted: %s", d.GetGameData().Name)
}

func (d DialogsFile) Compress() {
	if err := internal.DialogsFileCompressor(d.GetFileInfo()); err != nil {
		d.Log.Error().Err(err).Interface("DialogFile", util.ErrorObject(d.GetFileInfo())).Msg("Error compressing dialog file")
		return
	}

	if err := d.VerifyCompress(d.GetFileInfo(), internal.DialogsFileExtractor); err != nil {
		d.Log.Error().Err(err).Interface("DialogFile", util.ErrorObject(d.GetFileInfo())).Msg("Error verifying compressed dialog file")
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

		d.Log.Info().Msgf("All duplicated dialog files for %s have been created", d.GetGameData().Name)
	}

	d.Log.Info().Msgf("Dialog file compressed: %s", d.GetGameData().Name)
}
