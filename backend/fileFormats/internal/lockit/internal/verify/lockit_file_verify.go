package verify

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/parts"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"fmt"
)

type ILockitFileVerifier interface {
	VerifyExtract(partsList components.IList[parts.LockitFileParts], options interactions.LockitFileOptions) error
	VerifyCompress(destination locations.IDestination, options interactions.LockitFileOptions) error
}

type LockitFileVerifier struct {
	*base.FormatsBase

	FileValidator    IFileValidator
	LineBreakCounter ILineBreakCounter
}

func NewLockitFileVerifier(source interfaces.ISource, destination locations.IDestination) ILockitFileVerifier {
	return &LockitFileVerifier{
		FormatsBase:      base.NewFormatsBase(source, destination),
		FileValidator:    newFileValidator(),
		LineBreakCounter: new(LineBreakCounter),
	}
}

func (lv *LockitFileVerifier) VerifyExtract(parts components.IList[parts.LockitFileParts], options interactions.LockitFileOptions) error {
	if parts.GetLength() != options.PartsLength {
		return fmt.Errorf("error when ensuring splited lockit parts: expected %d | got %d", options.PartsLength, parts.GetLength())
	}

	if err := lv.LineBreakCounter.CountBinaryParts(parts, options); err != nil {
		return fmt.Errorf("error when counting binary line breaks in splited files: %w", err)
	}

	if err := lv.LineBreakCounter.CountTextParts(parts, options); err != nil {
		return fmt.Errorf("error when counting text line breaks in splited files: %w", err)
	}

	return nil
}

func (lv *LockitFileVerifier) VerifyCompress(destination locations.IDestination, options interactions.LockitFileOptions) error {
	importTargetFile := destination.Import().Get().GetTargetFile()
	lv.Log.Info().Msgf("Verifying reimported lockit file: %s", importTargetFile)

	if err := destination.Import().Get().Validate(); err != nil {
		return fmt.Errorf("reimport file not exists: %s | %w", destination.Import().Get().GetTargetFile(), err)
	}

	if err := lv.FileValidator.Validate(importTargetFile, options); err != nil {
		return err
	}

	lv.Log.Info().Msgf("Lockit file monted successfully to: %s", importTargetFile)

	return nil
}
