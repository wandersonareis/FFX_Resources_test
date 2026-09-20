package datastore

// IObjectsFormatter é o contrato de um formato de texto para os dados
// originais de objectsfile (ObjectTextData: array de entries + metadados).
// Cada formato (JSON, CSV, ...) implementa esta interface com uma struct
// própria e organiza a saída à sua maneira.
//
// Sem `any` e sem genéricos no contrato: os tipos são concretos, de modo que
// passar a representação errada não compila.
type IObjectsFormatter interface {
	// Extension retorna a extensão do arquivo do formato (ex: ".json").
	Extension() string
	// Marshal formata os dados originais para os bytes do formato de saída.
	Marshal(data ObjectTextData) ([]byte, error)
	// Unmarshal parseia os bytes do formato de volta para os dados originais.
	Unmarshal(data []byte) (ObjectTextData, error)
}
