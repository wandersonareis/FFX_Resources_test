package interfaces

import "ffxresources/backend/core"

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
	GetGamePartDuplicates(namePrefix string, gamePart core.GamePart) []string
}

type IValidate interface {
	Validate() error
}

/* type IExtractLocation interface {
	ILocationBase
	IValidate

	Get() *locations.ExtractLocation
} */

/* type ITargetExtractLocation interface {
	GetExtractLocation() IExtractLocation
} */

/* type ITranslateLocation interface {
	ILocationBase
	IValidate
}

type ITargetTranslateLocation interface {
	GetTranslateLocation() ITranslateLocation
} */

/* type IImportLocation interface {
	ILocationBase
	IValidate
}

type ITargetImportLocation interface {
	GetImportLocation() IImportLocation
} */

/* type IDestination interface {
	ITargetExtractLocation
	ITargetTranslateLocation
	ITargetImportLocation

	InitializeLocations(source ISource, formatter ITextFormatterDev)
	CreateRelativePath(source ISource, gameLocationPath string)
} */

type ILocationBase interface {
	SetTargetDirectory(path string)
	GetTargetDirectory() string
	GetTargetFile() string
	SetTargetFile(targetFile string)
	GetTargetPath() string
	SetTargetPath(targetPath string)
	ProvideTargetDirectory() error
	ProvideTargetPath() error
	BuildTargetOutput(source ISource, formatter ITextFormatterDev)
	DisposeTargetFile()
	DisposeTargetPath()
}
