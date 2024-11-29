package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/parts"
	"ffxresources/backend/interactions"
	"fmt"
)

type ILockitFileVerifier interface {
	VerifyExtract(partsList *[]parts.LockitFileParts, extractLocation *interactions.ExtractLocation, options interactions.LockitFileOptions) error
	VerifyCompress(dataInfo interactions.IGameDataInfo, options interactions.LockitFileOptions) error
}

type LockitFileVerifier struct {
	*base.FormatsBase

	FileValidator    IFileValidator
	LineBreakCounter ILineBreakCounter

	worker common.IWorker[parts.LockitFileParts]
}

func NewLockitFileVerifier(dataInfo interactions.IGameDataInfo) ILockitFileVerifier {
	worker := common.NewWorker[parts.LockitFileParts]()

	return &LockitFileVerifier{
		FormatsBase:      base.NewFormatsBase(dataInfo),
		FileValidator:    newFileValidator(),
		LineBreakCounter: new(LineBreakCounter),
		worker:           worker,
	}
}

func (lv *LockitFileVerifier) VerifyExtract(parts *[]parts.LockitFileParts, location *interactions.ExtractLocation, options interactions.LockitFileOptions) error {
	if len(*parts) != options.PartsLength {
		return fmt.Errorf("error when ensuring splited lockit parts: expected %d | got %d", options.PartsLength, len(*parts))
	}

	if err := lv.LineBreakCounter.CountBinaryParts(parts, options); err != nil {
		return fmt.Errorf("error when counting binary line breaks in splited files: %w", err)
	}

	if err := lv.LineBreakCounter.CountTextParts(parts, options); err != nil {
		return fmt.Errorf("error when counting text line breaks in splited files: %w", err)
	}

	return nil
}

func (lv *LockitFileVerifier) VerifyCompress(dataInfo interactions.IGameDataInfo, options interactions.LockitFileOptions) error {
	lv.Log.Info().Msgf("Verifying reimported lockit file: %s", dataInfo.GetImportLocation().TargetFile)

	if err := dataInfo.GetImportLocation().Validate(); err != nil {
		return fmt.Errorf("reimport file not exists: %s | %w", dataInfo.GetImportLocation().TargetFile, err)
	}

	if err := lv.FileValidator.Validate(dataInfo.GetImportLocation().TargetFile, options); err != nil {
		return err
	}

	lv.Log.Info().Msgf("Lockit file monted successfully to: %s", dataInfo.GetImportLocation().TargetFile)

	return nil
}
