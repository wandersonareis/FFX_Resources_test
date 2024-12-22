package interfaces

import (
	"ffxresources/backend/bases"
	"ffxresources/backend/core"
)

type IFileProcessor interface {
	Source() ISource
	Extract() error
	Compress() error
}

type ITextFormatterDev interface {
	ReadFile(source ISource, targetDirectory string) (string, string)
	WriteFile(source ISource, targetDirectory string) (string, string)
}

type ISource interface {
	Get() *core.SpiraFileInfo
	Set(source *core.SpiraFileInfo)
	GetGamePartDuplicates(namePrefix string, gamePart core.GameVersion) []string
}

type IValidate interface {
	Validate() error
}

type ILocationBase interface {
	bases.ITargetDirectoryBase
	bases.ITargetFileBase

	BuildTargetOutput(source ISource, formatter ITextFormatterDev)
	DisposeTargetFile()
	DisposeTargetPath()
}

type IInteractionBase interface {
	SetTargetDirectory(path string)
	GetTargetDirectory() string
	ProvideTargetDirectory() error
}