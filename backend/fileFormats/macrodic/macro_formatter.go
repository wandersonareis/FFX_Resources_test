package macrodic

// IMacroFormatter é o contrato de um formato de texto para os dados originais
// do dicionário de macros (containers por localization). Cada formato (JSON,
// CSV, ...) implementa esta interface com uma struct própria e organiza a
// saída à sua maneira.
//
// Sem `any` e sem genéricos no contrato: os tipos são concretos, de modo que
// passar a representação errada não compila.
type IMacroFormatter interface {
	// Extension retorna a extensão do arquivo do formato (ex: ".json").
	Extension() string
	// Marshal formata os containers originais para os bytes do formato de saída.
	Marshal(containers map[string]*MacroDictionaryBinaryFile) ([]byte, error)
	// Unmarshal desserializa os bytes do formato de volta para a representação
	// de importação usada hoje (MacroDictionaryJsonImport).
	Unmarshal(data []byte) (*MacroDictionaryJsonImport, error)
}
