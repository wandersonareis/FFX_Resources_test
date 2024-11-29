package util

/* import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type DlgKrnlVerify struct {
	*base.FormatsBase

	log zerolog.Logger
}

func NewDlgKrnlVerify(dataInfo interactions.IGameDataInfo) *DlgKrnlVerify {
	return &DlgKrnlVerify{
		FormatsBase: base.NewFormatsBase(dataInfo),
		log:         logger.Get().With().Str("module", "dlg_krnl_verify").Logger(),
	}
}

func (dv *DlgKrnlVerify) VerifyExtract(extract *interactions.ExtractLocation) error {
	dv.log.Info().Msgf("Verifying text file: %s", extract.TargetFile)

	if !dv.verifyText(extract.TargetFile) {
		dv.log.Error().Msgf("No segments found in the text file: %s", extract.TargetFile)

		if err := os.Remove(extract.TargetFile); err != nil {
			dv.log.Error().Err(err).Msgf("Error removing the text file: %s", extract.TargetFile)
		}

		return fmt.Errorf("error when verifying dialog file")
	}

	dv.log.Info().Msg("Text file verified")

	return nil
}

func (dv *DlgKrnlVerify) VerifyCompress(dataInfo interactions.IGameDataInfo, extractor func(dataInfo interactions.IGameDataInfo) error) error {
	dv.Log.Info().Msgf("Verifying compressed file: %s", dataInfo.GetGameData().Name)

	if err := dataInfo.GetImportLocation().Validate(); err != nil {
		dv.log.Error().Msgf("Reimport file not exists: %s", dataInfo.GetImportLocation().TargetFile)
		return err
	}

	tmp := common.NewTempProvider().ProvideTempFileWithExtension("tmp", ".txt")
	defer tmp.Dispose()

	tmpInfo := interactions.NewGameDataInfo(dataInfo.GetImportLocation().TargetFile)

	tmpInfo.GetExtractLocation().TargetFile = tmp.File
	tmpInfo.GetExtractLocation().TargetPath = tmp.FilePath

	if err := extractor(tmpInfo); err != nil {
		dv.log.Error().Err(err).Msg("Error on reimported dialog file")
		return err
	}

	if err := tmpInfo.GetExtractLocation().Validate(); err != nil {
		dv.log.Error().Msgf("Temp file text extracted not exists: %s", tmpInfo.GetExtractLocation().TargetFile)
		return err
	}

	translatedTextFile, err := os.ReadFile(dataInfo.GetTranslateLocation().TargetFile)
	if err != nil {
		dv.log.Error().Err(err).Msgf("Error reading the file: %s", dataInfo.GetTranslateLocation().TargetFile)
		return err
	}

	tmpReimportedTextFile, err := os.ReadFile(tmpInfo.GetExtractLocation().TargetFile)
	if err != nil {
		dv.log.Error().Err(err).Msgf("Error reading the file: %s", tmpInfo.GetExtractLocation().TargetFile)
		return err
	}

	if !bytes.Equal(translatedTextFile, tmpReimportedTextFile) {
		dv.log.Error().Msgf("Translated and reimported are not equal content: %s and %s", dataInfo.GetTranslateLocation().TargetFile, tmpInfo.GetExtractLocation().TargetFile)

		if err := os.Remove(tmpInfo.GetExtractLocation().TargetFile); err != nil {
			dv.log.Error().Err(err).Msgf("Error removing the file: %s", tmpInfo.GetExtractLocation().TargetFile)
		}

		return fmt.Errorf("error verifying dialog file: %s", dataInfo.GetImportLocation().TargetFile)
	}

	dv.log.Info().Msg("Compressed file verified")

	return nil
}

func (dv *DlgKrnlVerify) verifyText(targetFile string) bool {
	dialogsCount := common.CountSegments(targetFile)

	return dialogsCount != 0
}
 */