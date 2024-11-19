package internal

import (
	"bytes"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interactions"
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type LockitFileVerify struct {
	*base.FormatsBase
}

func NewLockitFileVerify(dataInfo interactions.IGameDataInfo) *LockitFileVerify {
	return &LockitFileVerify{
		FormatsBase: base.NewFormatsBase(dataInfo),
	}
}

func (l *LockitFileVerify) VerifyExtract(targetPath string, options *interactions.LockitFileOptions) error {
	if !l.verifyLockitFileParts(targetPath, options) && !l.verifyLockitTextParts(targetPath, options) {
		os.RemoveAll(targetPath)
		return fmt.Errorf("error when verifying splited lockit file")
	}

	l.Log.Info().Msgf("Lockit file splited successfully to: %s", targetPath)

	return nil
}

func (lv *LockitFileVerify) VerifyCompress(dataInfo interactions.IGameDataInfo, options *interactions.LockitFileOptions) error {
	if !lv.verifyLockitTextParts(dataInfo.GetTranslateLocation().TargetPath, options) &&
		!lv.verifyLockitFileParts(dataInfo.GetImportLocation().TargetPath, options) &&
		!lv.valideLockitFile(dataInfo.GetImportLocation().TargetFile, options) {
		os.Remove(dataInfo.GetImportLocation().TargetFile)
		return fmt.Errorf("error when verifying lockit file")
	}

	lv.Log.Info().Msgf("Lockit file reimported successfully to: %s", dataInfo.GetImportLocation().TargetFile)

	return nil
}

func (lv *LockitFileVerify) verifyLockitFileParts(targetPath string, options *interactions.LockitFileOptions) bool {
	parts := []LockitFileParts{}

	if err := util.FindFileParts(&parts, targetPath, LOCKIT_FILE_PARTS_PATTERN, NewLockitFileParts); err != nil {
		lv.Log.Error().Err(err).Msgf("error when finding lockit parts in %s", targetPath)
		return false
	}

	if len(parts) != options.PartsLength {
		lv.Log.Error().Int("Current parts", len(parts)).Int("Expected parts", options.PartsLength).Msg("invalid number of parts")
		return false
	}

	if !lv.countLockitLineBreaks(parts, options, lv.Log) {
		lv.Log.Error().Msg("invalid line breaks count")
		return false
	}

	return true
}

func (lv *LockitFileVerify) verifyLockitTextParts(targetPath string, options *interactions.LockitFileOptions) bool {
	parts := []LockitFileParts{}

	if err := util.FindFileParts(&parts, targetPath, LOCKIT_TXT_PARTS_PATTERN, NewLockitFileParts); err != nil {
		lv.Log.Error().Err(err).Msgf("error when finding lockit text parts in %s", targetPath)
		return false
	}

	if len(parts) != options.PartsLength {
		lv.Log.Error().Int("Text parts", len(parts)).Int("Expected text parts", options.PartsLength).Msg("invalid number of text parts")
		return false
	}

	if !lv.countLockitLineBreaks(parts, options, lv.Log) {
		lv.Log.Error().Msg("invalid line breaks count")
		return false
	}

	return true
}

func (lv *LockitFileVerify) valideLockitFile(lockedFilePath string, options *interactions.LockitFileOptions) bool {
	lockitFileData, err := os.ReadFile(lockedFilePath)
	if err != nil {
		return false
	}

	reimportedLineBreaksCount := countAllLineEndings(lockitFileData)

	if options.LineBreaksCount != reimportedLineBreaksCount {
		lv.Log.Error().Msgf("line breaks count does not match. Expected: %d, got: %d", options.LineBreaksCount, reimportedLineBreaksCount)
		return false
	}

	return true
}

func (lv *LockitFileVerify) countLockitLineBreaks(parts []LockitFileParts, options *interactions.LockitFileOptions, log zerolog.Logger) bool {
	partsSize := options.PartsSizes
	size := 0

	for i, part := range parts {
		switch true {
		case i == 0:
			size = partsSize[i]
		case i > 0 && i < len(partsSize):
			size = partsSize[i] - partsSize[i-1]
		case i <= len(partsSize):
			size = options.LineBreaksCount - partsSize[i-1]
		}

		partData, err := os.ReadFile(part.GetGameData().FullFilePath)
		if err != nil {
			log.Error().Err(err).Msgf("error when reading file %s", part.GetGameData().Name)
			return false
		}

		ocorrences := bytes.Count(partData, []byte{0x0d, 0x0a})

		if ocorrences != size {
			log.Error().Msgf("File %s has %d line breaks, expected %d", part.GetGameData().Name, ocorrences, size)
			return false
		}

	}

	return true
}

func countAllLineEndings(buffer []byte) int {
	return bytes.Count(buffer, []byte("\r\n"))
}
