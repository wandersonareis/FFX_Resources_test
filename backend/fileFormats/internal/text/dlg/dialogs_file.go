package dlg

import (
	"ffxresources/backend/fileFormats/internal/text/dlg/internal"
	verify "ffxresources/backend/fileFormats/internal/text/lib/dlg_krnl_verify"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"fmt"
	"slices"

	"github.com/rs/zerolog"
)

type DialogsFile struct {
	dialogsClones internal.IDlgClones
	decoder       internal.IDlgDecoder
	encoder       internal.IDlgEncoder
	textVerifyer  verify.ITextsVerify
	dataInfo      interactions.IGameDataInfo
	log           zerolog.Logger
}

func NewDialogs(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	return &DialogsFile{
		dialogsClones: internal.NewDlgClones(dataInfo),
		decoder:       internal.NewDlgDecoder(),
		encoder:       internal.NewDlgEncoder(),
		textVerifyer:  verify.NewTextsVerify(),

		dataInfo: dataInfo,
		log:      logger.Get().With().Str("module", "dialogs_file").Logger(),
	}
}

func (d DialogsFile) GetFileInfo() interactions.IGameDataInfo {
	return d.dataInfo
}

func (d DialogsFile) Extract() error {
	if slices.Contains(d.dataInfo.GetGameData().ClonedItems, d.dataInfo.GetGameData().RelativeGameDataPath) {
		return nil
	}

	if err := d.decoder.Decoder(d.GetFileInfo()); err != nil {
		d.log.Error().
			Err(err).
			Str("file", d.GetFileInfo().GetGameData().FullFilePath).
			Msg("Error decoding dialog file")

		return fmt.Errorf("failed to decode dialog file: %s", d.GetFileInfo().GetGameData().Name)
	}

	if err := d.textVerifyer.VerifyExtract(d.dataInfo.GetExtractLocation()); err != nil {
		d.log.Error().
			Err(err).
			Str("file", d.GetFileInfo().GetExtractLocation().TargetFile).
			Msg("Error verifying text file")

		return fmt.Errorf("failed to verify text file: %s", d.GetFileInfo().GetExtractLocation().TargetFile)
	}

	d.log.Info().
		Str("file", d.GetFileInfo().GetExtractLocation().TargetFile).
		Msg("Dialog file extracted successfully")

	return nil
}

func (d DialogsFile) Compress() error {
	if err := d.encoder.Encoder(d.GetFileInfo()); err != nil {
		d.log.Error().
			Err(err).
			Str("file", d.GetFileInfo().GetTranslateLocation().TargetFile).
			Msg("Error compressing dialog file")

		return fmt.Errorf("failed to compress dialog file: %s", d.GetFileInfo().GetTranslateLocation().TargetFile)
	}

	if err := d.textVerifyer.VerifyCompress(d.GetFileInfo(), d.decoder.Decoder); err != nil {
		d.log.Error().
			Err(err).
			Str("file", d.GetFileInfo().GetImportLocation().TargetFile).
			Msg("Error verifying text file")

		return fmt.Errorf("failed to verify text file: %s", d.GetFileInfo().GetImportLocation().TargetFile)
	}

	d.dialogsClones.Clone()

	d.log.Info().
		Str("file", d.GetFileInfo().GetImportLocation().TargetFile).
		Msg("Dialog file compressed")

	return nil
}
