package interfaces

import (
	"ffxresources/backend/core"
	"ffxresources/backend/models"
)

type IFileProcessor interface {
	Source() ISource
	Extract() error
	Compress() error
}

type ITextFormatter interface {
	ReadFile(source ISource, targetDirectory string) (string, string)
	WriteFile(source ISource, targetDirectory string) (string, string)
}

type ISource interface {
	Get() *core.SpiraFileInfo
	Set(source *core.SpiraFileInfo)
	GetGamePartDuplicates(namePrefix string, gamePart models.GameVersion) []string
}

type IValidate interface {
	Validate() error
}

type IInteractionBase interface {
	SetTargetDirectory(path string)
	GetTargetDirectory() string
	ProvideTargetDirectory() error
}
