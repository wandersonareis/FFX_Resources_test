package lockit

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/baseFormats"
	"ffxresources/backend/fileFormats/internal/lockit/internal"
	"ffxresources/backend/fileFormats/internal/lockit/internal/integrity"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitParts"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/loggingService"
	"fmt"
	"os"
)

type (
	ILockitFileCompressor interface {
		Compress() error
	}

	LockitFileCompressor struct {
		baseFormats.IBaseFileFormat

		lockitFilePartsEncoder   lockitParts.ILockitFilePartsEncoder
		lockitFilePartsIntegrity integrity.ILockitFilePartsIntegrity
		lockitFileSplitter       internal.ILockitFileSplitter
		lockitFileJoiner         internal.ILockitPartsJoiner

		lockitEncoding ffxencoding.IFFXTextLockitEncoding
		options        core.ILockitFileOptions
		logger         loggingService.ILoggerService
	}
)

func NewLockitFileCompressor(
	source interfaces.ISource,
	destination locations.IDestination,
	lockitEncoding ffxencoding.IFFXTextLockitEncoding,
	fileOptions core.ILockitFileOptions,
	logger loggingService.ILoggerService,
) *LockitFileCompressor {
	return &LockitFileCompressor{
		IBaseFileFormat:          baseFormats.NewFormatsBase(source, destination),
		lockitFilePartsEncoder:   lockitParts.NewLockitFilePartsEncoder(logger),
		lockitFilePartsIntegrity: integrity.NewLockitFilePartsIntegrity(logger),
		lockitFileSplitter:       internal.NewLockitFileSplitter(),
		lockitFileJoiner:         internal.NewLockitFileJoiner(logger),

		lockitEncoding: lockitEncoding,
		options:        fileOptions,
		logger:         logger,
	}
}

func (lfc *LockitFileCompressor) Compress() error {
	partsLength := lfc.options.GetPartsLength()

	textTranslatedFilePartsList := components.NewList[lockitParts.LockitFileParts](partsLength)
	defer textTranslatedFilePartsList.Clear()

	if err := lfc.populateLockitTranslatedTextFileParts(textTranslatedFilePartsList); err != nil {
		return err
	}

	if err := lfc.ensureAllLockitTranslatedTextFileParts(textTranslatedFilePartsList, partsLength); err != nil {
		return err
	}

	binaryExtractedFilePartsList := components.NewList[lockitParts.LockitFileParts](partsLength)
	defer binaryExtractedFilePartsList.Clear()

	if err := lfc.populateLockitExtractedBinaryFilePartsList(binaryExtractedFilePartsList); err != nil {
		return err
	}

	if err := lfc.ensureAllLockitExtractedBinaryFileParts(binaryExtractedFilePartsList, partsLength); err != nil {
		return err
	}

	if err := lfc.encodingFilesParts(binaryExtractedFilePartsList); err != nil {
		return err
	}

	binaryTranslatedFilePartsList := components.NewList[lockitParts.LockitFileParts](partsLength)
	defer lfc.disposeList(binaryTranslatedFilePartsList)

	if err := lfc.populateLockitTranslatedBinaryFileParts(binaryTranslatedFilePartsList); err != nil {
		return err
	}

	if err := lfc.ensureAllLockitTranslatedBinaryFileParts(binaryTranslatedFilePartsList, partsLength); err != nil {
		return err
	}

	if err := lfc.joiningLockitBinaryFileParts(binaryTranslatedFilePartsList); err != nil {
		return err
	}

	return nil
}

func (lfc *LockitFileCompressor) populateLockitExtractedBinaryFilePartsList(extractedBinaryPartsList components.IList[lockitParts.LockitFileParts]) error {
	lfc.logger.Info("Populating lockit extracted binary file parts...")

	if err := lockitParts.PopulateLockitBinaryFileParts(
		extractedBinaryPartsList,
		lfc.GetDestination().Extract().GetTargetPath(),
	); err != nil {
		return fmt.Errorf("error populating lockit extracted binary file parts: %w", err)
	}

	return nil
}

func (lfc *LockitFileCompressor) populateLockitTranslatedTextFileParts(translatedTextPartsList components.IList[lockitParts.LockitFileParts]) error {
	lfc.logger.Info("Populating lockit translated text file parts...")

	if err := lockitParts.PopulateLockitTextFileParts(
		translatedTextPartsList,
		lfc.GetDestination().Translate().GetTargetPath(),
	); err != nil {
		return fmt.Errorf("error populating lockit translated text file parts: %w", err)
	}

	return nil
}

func (lfc *LockitFileCompressor) populateLockitTranslatedBinaryFileParts(translatedBinaryPartsList components.IList[lockitParts.LockitFileParts]) error {
	lfc.logger.Info("Populating lockit translated binary file parts...")

	if err := lockitParts.PopulateLockitBinaryFileParts(
		translatedBinaryPartsList,
		lfc.GetDestination().Translate().GetTargetPath(),
	); err != nil {
		return fmt.Errorf("error populating lockit translated binary file parts: %w", err)
	}

	return nil
}

func (lfc *LockitFileCompressor) ensureAllLockitTranslatedTextFileParts(translatedTextPartsList components.IList[lockitParts.LockitFileParts], partsLength int) error {
	if translatedTextPartsList.GetLength() != partsLength {
		return fmt.Errorf("error ensuring translated lockit text parts: expected %d, got %d on path: %s",
			partsLength, translatedTextPartsList.GetLength(), lfc.GetDestination().Translate().GetTargetPath())
	}

	translatedTextList := components.NewList[string](partsLength)
	defer translatedTextList.Clear()

	translatedTextPartsList.ForEach(func(part lockitParts.LockitFileParts) {
		translatedTextList.Add(
			part.GetSource().GetPath())
	})

	if err := lfc.validateLineBreaksCount(translatedTextList); err != nil {
		return fmt.Errorf("error validating line breaks count for lockit translated text parts: %w", err)
	}

	lfc.logger.Info("Lockit file translated text parts validated: %s", lfc.GetDestination().Translate().GetTargetPath())

	return nil
}

func (lfc *LockitFileCompressor) ensureAllLockitTranslatedBinaryFileParts(translatedBinaryPartsList components.IList[lockitParts.LockitFileParts], partsLength int) error {
	if translatedBinaryPartsList.GetLength() != partsLength {
		return fmt.Errorf("error ensuring translated lockit binary parts: expected %d, got %d on path: %s",
			partsLength, translatedBinaryPartsList.GetLength(), lfc.GetDestination().Translate().GetTargetPath())
	}

	translatedBinaryList := components.NewList[string](partsLength)
	defer translatedBinaryList.Clear()

	translatedBinaryPartsList.ForEach(func(part lockitParts.LockitFileParts) {
		translatedBinaryList.Add(
			part.GetSource().GetPath())
	})

	if err := lfc.validateLineBreaksCount(translatedBinaryList); err != nil {
		return fmt.Errorf("error validating line breaks count for lockit translated binary parts: %w", err)
	}

	lfc.logger.Info("Lockit file translated binary parts validated: %s", lfc.GetDestination().Translate().GetTargetPath())

	return nil
}

func (lfc *LockitFileCompressor) ensureAllLockitExtractedBinaryFileParts(extractedBinaryPartsList components.IList[lockitParts.LockitFileParts], partsLength int) error {
	maxAttempts := 3
	for range maxAttempts {
		if extractedBinaryPartsList.GetLength() == partsLength {
			break
		}

		if err := lfc.extractMissingLockitBinaryFileParts(); err != nil {
			return err
		}

		extractedBinaryPartsList.Clear()

		if err := lfc.populateLockitExtractedBinaryFilePartsList(extractedBinaryPartsList); err != nil {
			return err
		}
	}

	if extractedBinaryPartsList.GetLength() != partsLength {
		return fmt.Errorf("error ensuring extracted lockit binary parts: expected %d, got %d on path: %s",
			partsLength, extractedBinaryPartsList.GetLength(), lfc.GetDestination().Extract().GetTargetPath())
	}

	extractedBinaryList := components.NewList[string](partsLength)
	defer extractedBinaryList.Clear()

	extractedBinaryPartsList.ForEach(func(part lockitParts.LockitFileParts) {
		extractedBinaryList.Add(
			part.GetSource().GetPath())
	})

	if err := lfc.validateLineBreaksCount(extractedBinaryList); err != nil {
		return fmt.Errorf("error validating line breaks count for lockit extracted binary parts: %w", err)
	}

	lfc.logger.Info("Lockit file extracted binary parts validated: %s", lfc.GetDestination().Extract().GetTargetPath())

	return nil
}

func (l *LockitFileCompressor) extractMissingLockitBinaryFileParts() error {
	l.logger.Info("Missing lockit file parts detected. Attempting to extract...")

	if err := l.lockitFileSplitter.FileSplitter(
		l.GetSource(),
		l.GetDestination().Extract(),
		l.options); err != nil {
		return fmt.Errorf("error extracting missing lockit binary file parts: %w", err)
	}

	l.logger.Info("Missing lockit file parts extracted: %s", l.GetDestination().Extract().GetTargetPath())

	return nil
}

func (lfc *LockitFileCompressor) encodingFilesParts(extractedBinaryPartsList components.IList[lockitParts.LockitFileParts]) error {
	if extractedBinaryPartsList.IsEmpty() {
		return fmt.Errorf("error when encoding files parts: extracted binary parts list is empty")
	}

	lfc.logger.Info("Encoding files parts to: %s", lfc.GetDestination().Import().GetTargetPath())

	// TODO: Implement a way to get the game version from the source file
	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()

	if err := lfc.lockitFilePartsEncoder.EncodeFilesParts(extractedBinaryPartsList, lfc.lockitEncoding, gameVersion); err != nil {
		return fmt.Errorf("error when encoding files parts: %s", err.Error())
	}

	return nil
}
func (lfc *LockitFileCompressor) joiningLockitBinaryFileParts(translatedBinaryPartsList components.IList[lockitParts.LockitFileParts]) error {
	lfc.logger.Info("Joining process of the translated files to file: %s", lfc.GetDestination().Import().GetTargetFile())

	if err := lfc.lockitFileJoiner.JoinFileParts(lfc.GetDestination(), translatedBinaryPartsList, lfc.options); err != nil {
		return fmt.Errorf("error when joining lockit binary file parts: %s", err.Error())
	}

	return nil
}

func (lfc *LockitFileCompressor) validateLineBreaksCount(filesList components.IList[string]) error {
	if err := lfc.lockitFilePartsIntegrity.ComparePartsLineBreaksCount(
		filesList,
		lfc.options,
	); err != nil {
		return fmt.Errorf("error validating line breaks count for lockit file parts: %w", err)
	}

	return nil
}

func (lfc *LockitFileCompressor) disposeList(list components.IList[lockitParts.LockitFileParts]) {
	list.ForEach(func(part lockitParts.LockitFileParts) {
		os.Remove(part.GetSource().GetPath())
	})

	list.Clear()
}
