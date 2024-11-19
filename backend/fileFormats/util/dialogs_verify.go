package util

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/interactions"
	"fmt"
	"os"
)

type DlgKrnlVerify struct {
	*base.FormatsBase
}

func NewDlgKrnlVerify(dataInfo interactions.IGameDataInfo) *DlgKrnlVerify {
	return &DlgKrnlVerify{
		FormatsBase: base.NewFormatsBase(dataInfo),
	}
}

func (dv *DlgKrnlVerify) VerifyExtract(extract *interactions.ExtractLocation) error {
	if !dv.verifyText(extract.TargetFile) {
		dv.Log.Error().Msgf("No dialogs found in the file: %s", extract.TargetFile)
		if err := os.Remove(extract.TargetFile); err != nil {
			dv.Log.Error().Err(err).Msgf("Error removing the file: %s", extract.TargetFile)
		}

		return fmt.Errorf("error when verifying dialog file")
	}

	dv.Log.Info().Msgf("Dialog file extracted: %s", extract.TargetFileName)

	return nil
}

func (dv *DlgKrnlVerify) VerifyCompress(dataInfo interactions.IGameDataInfo, extractor func(dataInfo interactions.IGameDataInfo) error) error {
	if err := dataInfo.GetImportLocation().Validate(); err != nil {
		dv.Log.Error().Msgf("Reimport file not exists: %s", dataInfo.GetImportLocation().TargetFile)
		return err
	}

	tmp := common.NewTempProvider().ProvideTempFileWithExtension("tmp", ".txt")
	defer tmp.Dispose()

	tmpInfo := interactions.NewGameDataInfo(dataInfo.GetImportLocation().TargetFile)

	tmpInfo.GetExtractLocation().TargetFile = tmp.File
	tmpInfo.GetExtractLocation().TargetPath = tmp.FilePath

	if err := extractor(tmpInfo); err != nil {
		dv.Log.Error().Err(err).Interface("DialogFile", ErrorObject(dv.GetFileInfo())).Msg("Error on reimported dialog file")
		return err
	}

	if err := tmpInfo.GetExtractLocation().Validate(); err != nil {
		dv.Log.Error().Msgf("Temp file text extracted not exists: %s", tmpInfo.GetExtractLocation().TargetFile)
		return err
	}

	translatedTextFile, err := os.ReadFile(dataInfo.GetTranslateLocation().TargetFile)
	if err != nil {
		dv.Log.Error().Err(err).Msgf("Error reading the file: %s", dataInfo.GetTranslateLocation().TargetFile)
		return err
	}

	tmpReimportedTextFile, err := os.ReadFile(tmpInfo.GetExtractLocation().TargetFile)
	if err != nil {
		dv.Log.Error().Err(err).Msgf("Error reading the file: %s", tmpInfo.GetExtractLocation().TargetFile)
		return err
	}

	if !bytes.Equal(translatedTextFile, tmpReimportedTextFile) {
		dv.Log.Error().Msgf("Translated and reimported are not equal content: %s and %s", dataInfo.GetTranslateLocation().TargetFile, tmpInfo.GetExtractLocation().TargetFile)

		if err := os.Remove(tmpInfo.GetExtractLocation().TargetFile); err != nil {
			dv.Log.Error().Err(err).Msgf("Error removing the file: %s", tmpInfo.GetExtractLocation().TargetFile)
		}

		return fmt.Errorf("error verifying dialog file: %s", dataInfo.GetImportLocation().TargetFile)
	}

	dv.Log.Info().Msgf("Dialog file verified: %s", dataInfo.GetGameData().Name)

	return nil
}

func (dv *DlgKrnlVerify) verifyText(targetFile string) bool {
	dialogsCount := common.CountSegments(targetFile)

	return dialogsCount != 0
}