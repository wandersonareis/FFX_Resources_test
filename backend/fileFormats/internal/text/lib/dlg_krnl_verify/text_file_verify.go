package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"os"

	"github.com/rs/zerolog"
)

type ITextsVerify interface {
	VerifyExtract(extract locations.IExtractLocation) error
	VerifyCompress(source interfaces.ISource, destination locations.IDestination, extractor func(source interfaces.ISource, destination locations.IDestination) error) error
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

func (dv *TextsVerify) VerifyExtract(extract locations.IExtractLocation) error {
	if err := dv.segmentCounter.CountBinary(extract.GetTargetFile()); err != nil {
		dv.log.Error().
			Err(err).
			Str("file", extract.GetTargetFile()).
			Msg("Error on text file")

		if err := os.Remove(extract.GetTargetFile()); err != nil {
			dv.log.Error().
				Err(err).
				Msgf("Error removing the text file: %s", extract.GetTargetFile())
		}

		return err
	}

	if err := dv.segmentCounter.CountText(extract.GetTargetFile()); err != nil {
		dv.log.Error().Err(err).Send()

		if err := os.Remove(extract.GetTargetFile()); err != nil {
			dv.log.Error().Err(err).Msgf("Error removing the text file: %s", extract.GetTargetFile())
		}

		return err
	}

	dv.log.Info().Msgf("Text file verified successfully: %s", extract.GetTargetFile())

	return nil
}

func (dv *TextsVerify) VerifyCompress(source interfaces.ISource, destination locations.IDestination, extractor func(source interfaces.ISource, destination locations.IDestination) error) error {
	extractLocation := destination.Extract().Get()
	translateLocation := destination.Translate().Get()
	importLocation := destination.Import().Get()
	if err := importLocation.Validate(); err != nil {
		dv.log.Error().Msgf("Reimport file not exists: %s", importLocation.GetTargetFile())
		return err
	}

	dv.createTemporaryFileInfo(source, destination)
	defer extractLocation.DisposeTargetFile()

	if err := extractor(source, destination); err != nil {
		dv.log.Error().Err(err).Msg("Error on reimported dialog file")
		return err
	}

	if err := dv.filesComparer.CompareTranslatedTextParts(translateLocation.GetTargetFile(), extractLocation.GetTargetFile()); err != nil {
		dv.log.Error().Err(err).Msg("Error os reimported text file")
		return err
	}

	dv.log.Info().Msgf("Compressed text file verified successfully: %s", source.Get().Name)

	return nil
}

func (dv *TextsVerify) createTemporaryFileInfo(source interfaces.ISource, destination locations.IDestination) {
	tmp := common.NewTempProviderDev("tmp", ".txt")

	destination.Extract().Get().SetTargetFile(tmp.TempFile)
	destination.Extract().Get().SetTargetPath(tmp.TempFilePath)

	s := source.Get()
	s.Path = destination.Import().Get().GetTargetFile()
	source.Set(s)
}
