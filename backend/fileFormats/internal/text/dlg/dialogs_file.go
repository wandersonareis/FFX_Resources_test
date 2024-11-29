package dlg

import (
	"ffxresources/backend/fileFormats/internal/text/dlg/internal"
	"ffxresources/backend/fileFormats/internal/text/lib/dlg_krnl_verify"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"slices"

	"github.com/rs/zerolog"
)

type DialogsFile struct {
	dialogsClones internal.IDlgClones
	textVerifyer  *verify.DlgKrnlVerify
	dataInfo      interactions.IGameDataInfo
	log           zerolog.Logger
}

func NewDialogs(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	return &DialogsFile{
		dialogsClones: internal.NewDlgClones(dataInfo),
		textVerifyer:  verify.NewDlgKrnlVerify(),
		dataInfo:      dataInfo,
		log:           logger.Get().With().Str("module", "dialogs_file").Logger(),
	}
}

func (d DialogsFile) GetFileInfo() interactions.IGameDataInfo {
	return d.dataInfo
}

func (d DialogsFile) Extract() {
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			d.log.Info().Msgf("Disposing target file: %s", d.GetFileInfo().GetImportLocation().TargetFile)

			d.GetFileInfo().GetImportLocation().DisposeTargetFile()

			close(errChan)
		}()

		for err := range errChan {
			d.log.Error().Err(err).Msg("error when verifying monted macrodic file")

			return
		}
	}()

	if slices.Contains(d.dataInfo.GetGameData().ClonedItems, d.dataInfo.GetGameData().RelativeGameDataPath) {
		return
	}

	if err := internal.DialogsFileExtractor(d.GetFileInfo()); err != nil {
		d.log.Error().Err(err).Interface("DialogFile", util.ErrorObject(d.GetFileInfo())).Msg("Error extracting dialog file")
		return
	}

	if err := d.textVerifyer.VerifyExtract(d.dataInfo.GetExtractLocation()); err != nil {
		d.log.Error().Err(err).Msg("Error verifying text file")
		return
	}

	d.log.Info().Msgf("Dialog file extracted successfully: %s", d.dataInfo.GetGameData().Name)
}

func (d DialogsFile) Compress() {
	if err := internal.DialogsFileCompressor(d.GetFileInfo()); err != nil {
		d.log.Error().Err(err).Interface("DialogFile", util.ErrorObject(d.GetFileInfo())).Msg("Error compressing dialog file")
		return
	}

	if err := d.textVerifyer.VerifyCompress(d.GetFileInfo(), internal.DialogsFileExtractor); err != nil {
		d.log.Error().Err(err).Msg("Error verifying text file")
		return
	}
	
	d.dialogsClones.Clone()

	d.log.Info().Msgf("Dialog file compressed: %s", d.dataInfo.GetGameData().Name)
}
