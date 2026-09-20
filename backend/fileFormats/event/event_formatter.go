package event

import "ffxresources/backend/common"

// IEventsFormatter é o contrato de um formato de texto para os dados
// originais de eventos (array de EventFileData). Cada formato (JSON, CSV,
// ...) implementa esta interface com uma struct própria e organiza a saída
// à sua maneira.
//
// Sem `any` e sem genéricos no contrato: os tipos são concretos, de modo que
// passar a representação errada não compila.
type IEventsFormatter interface {
	// Extension retorna a extensão do arquivo do formato (ex: ".json").
	Extension() string
	// Marshal formata os eventos originais para os bytes do formato de saída.
	// version contextualiza metadados quando o formato os exige.
	Marshal(events []EventFileData, version common.GameVersion) ([]byte, error)
	// Unmarshal parseia os bytes do formato de volta para o array original
	// de eventos, ordenado por ID.
	Unmarshal(data []byte) ([]EventFileData, error)
}
