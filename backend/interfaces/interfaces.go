package interfaces

import (
	"ffxresources/backend/common"
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
		GetVersion() common.GameVersion
		IsDir() bool
		PopulateDuplicatesFiles()
	}

	IFileProcessor interface {
		ICompressor
		IExtractor
		GetSource() ISource
	}

	ITextFormatter interface {
		GetTargetExtension() string
		ReadFile(source ISource, targetDirectory string) (string, string)
		WriteFile(source ISource, targetDirectory string) (string, string)
	}

	IValidate interface {
		Validate() error
	}

	// IBinaryFile orquestra o ciclo de vida de um arquivo binário de localização:
	// carregar do binário e salvar de volta no binário.
	//
	// Exportação/importação de texto não fazem parte deste contrato
	// compartilhado: cada domínio usa um formatter tipado recebido diretamente
	// nos métodos ExportToJson/ImportFromJson dos tipos concretos
	// (datastore.IObjectsFormatter, event.IEventsFormatter,
	// macrodic.IMacroFormatter), sem `any` e sem reflexão.
	//
	// É genérica no tipo do objeto de texto (T) de propósito: referenciar aqui
	// core/components ou datastore formaria um import cíclico, pois
	// core/components já depende deste package (IList/IMap). A especialização
	// concreta usada pelos readers continua em datastore.IBinaryFile.
	IBinaryFile[T any] interface {
		LoadFromBinary() error
		SaveToBinary(filePath string) error
		GetObjects() IList[T]
	}

	IInteractionBase interface {
		SetTargetDirectory(path string) error
		GetTargetDirectory() string
		ProvideTargetDirectory() error
	}
)
