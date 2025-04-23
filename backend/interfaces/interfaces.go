package interfaces

import (
	"ffxresources/backend/core"
	"ffxresources/backend/models"
)

type (
	IExtractor interface {
		Extract() error
	}

	ICompressor interface {
		Compress() error
	}

	ISource interface {
		Get() *core.SpiraFileInfo
		Set(source *core.SpiraFileInfo)
		PopulateDuplicatesFiles(gamePart models.GameVersion)
	}

	IFileProcessor interface {
		ICompressor
		IExtractor
		GetSource() ISource
	}

	ITextFormatter interface {
		ReadFile(source ISource, targetDirectory string) (string, string)
		WriteFile(source ISource, targetDirectory string) (string, string)
	}

	IValidate interface {
		Validate() error
	}

	IInteractionBase interface {
		SetTargetDirectory(path string) error
		GetTargetDirectory() string
		ProvideTargetDirectory() error
	}
)

