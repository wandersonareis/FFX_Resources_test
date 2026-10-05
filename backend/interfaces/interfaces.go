package interfaces

type (
	// IBinaryFile orquestra o ciclo de vida de um arquivo binário de localização:
	// carregar do binário e salvar de volta no binário.
	//
	// Exportação/importação de texto não fazem parte deste contrato
	// compartilhado: o texto flui como DTO (backend/dto), montado por
	// backend/builders a partir dos dados brutos e serializado por
	// backend/formatters/json, sem `any` e sem reflexão. O pacote event
	// desconhece JSON; objectsfile/macrodic seguem o mesmo caminho.
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
