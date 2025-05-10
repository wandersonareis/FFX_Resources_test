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
		GetPath() string
		SetPath(path string)
		GetRelativePath() string
		SetRelativePath(relativePath string)
		GetName() string
		GetNameWithoutExtension() string
		GetParentPath() string
		GetSize() int64
		GetType() models.NodeType
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
