package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"os"

	"github.com/rs/zerolog"
)

type ITextsVerify interface {
	VerifyExtract(extract *interactions.ExtractLocation) error
	VerifyCompress(dataInfo interactions.IGameDataInfo, extractor func(dataInfo interactions.IGameDataInfo) error) error
}

type TextsVerify struct {
	segmentCounter ISegmentCounter
	filesComparer  IComparer

	log zerolog.Logger
}

func NewTextsVerify() ITextsVerify {
	return &TextsVerify{
		segmentCounter: new(segmentCounter),
		filesComparer:  newPartComparer(),

		log: logger.Get().With().Str("module", "texts_verify").Logger(),
	}
}

func (dv *TextsVerify) VerifyExtract(extract *interactions.ExtractLocation) error {
	if err := dv.segmentCounter.CountBinary(extract.TargetFile); err != nil {
		dv.log.Error().
			Err(err).
			Str("file", extract.TargetFile).
			Msg("Error on text file")

		if err := os.Remove(extract.TargetFile); err != nil {
			dv.log.Error().
				Err(err).
				Msgf("Error removing the text file: %s", extract.TargetFile)
		}

		return err
	}

	if err := dv.segmentCounter.CountText(extract.TargetFile); err != nil {
		dv.log.Error().Err(err).Send()

		if err := os.Remove(extract.TargetFile); err != nil {
			dv.log.Error().Err(err).Msgf("Error removing the text file: %s", extract.TargetFile)
		}

		return err
	}

	dv.log.Info().Msgf("Text file verified successfully: %s", extract.TargetFile)

	return nil
}

func (dv *TextsVerify) VerifyCompress(dataInfo interactions.IGameDataInfo, extractor func(dataInfo interactions.IGameDataInfo) error) error {
	if err := dataInfo.GetImportLocation().Validate(); err != nil {
		dv.log.Error().Msgf("Reimport file not exists: %s", dataInfo.GetImportLocation().TargetFile)
		return err
	}

	dv.createTemporaryFileInfo(dataInfo.GetGameDataInfo())
	defer dataInfo.GetExtractLocation().DisposeTargetFile()

	if err := extractor(dataInfo); err != nil {
		dv.log.Error().Err(err).Msg("Error on reimported dialog file")
		return err
	}

	if err := dv.filesComparer.CompareTranslatedTextParts(dataInfo.GetTranslateLocation().TargetFile, dataInfo.GetExtractLocation().TargetFile); err != nil {
		dv.log.Error().Err(err).Msg("Error os reimported text file")
		return err
	}

	dv.log.Info().Msgf("Compressed text file verified successfully: %s", dataInfo.GetGameData().Name)

	return nil
}

func (dv *TextsVerify) createTemporaryFileInfo(dataInfo *interactions.GameDataInfo) {
	tmp := common.NewTempProviderDev("tmp", ".txt")

	dataInfo.GetExtractLocation().TargetFile = tmp.TempFile
	dataInfo.GetExtractLocation().TargetPath = tmp.TempFilePath

	dataInfo.GetGameData().FullFilePath = dataInfo.GetImportLocation().TargetFile
}
