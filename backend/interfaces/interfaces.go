package interfaces

import "ffxresources/backend/models"

type (
	IExtractor interface {
		Extract() error
	}

	ICompressor interface {
		Compress() error
	}

	ISource interface {
		Get() models.SpiraFileInfo
		GetName() string
		GetNameWithoutExtension() string
		GetExtension() string
		GetPath() string
		SetPath(path string)
		GetRelativePath() string
		SetRelativePath(relativePath string)
		GetParentPath() string
		GetSize() int64
		GetType() models.NodeType
		GetVersion() models.GameVersion
		IsDir() bool
		PopulateDuplicatesFiles()
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
