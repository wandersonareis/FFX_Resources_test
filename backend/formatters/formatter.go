// Package formatters fornece o registro de formatters de exportação/importação
// (JSON, Strings, CSV, XLSX). Cada formatter implementa a interface Formatter
// e é registrado pelo seu nome canônico (extensão sem ponto).
package formatters

import (
	"ffxresources/backend/dto"
)

// Formatter define o contrato mínimo para exportar/importar uma Collection (DTO).
// Os formatters recebem apenas DTO pronto (sem conhecimento do domínio).
type Formatter interface {
	// Extension retorna a extensão do arquivo (ex: ".json", ".csv", ".xlsx").
	Extension() string
	// Marshal serializa a Collection completa (todos os idiomas).
	Marshal(c dto.Collection) ([]byte, error)
	// MarshalLangs serializa a Collection filtrando os idiomas pedidos
	// (nil/vazio = todos os idiomas).
	MarshalLangs(c dto.Collection, langs []string) ([]byte, error)
	// Unmarshal parseia os bytes de volta para uma Collection (DTO).
	Unmarshal(data []byte) (dto.Collection, error)
}

// Registry mapeia nomes canônicos (ex: "json", "csv", "xlsx") para formatters.
var Registry = map[string]Formatter{}

// Register adiciona um formatter ao registro (chamado no init de cada package).
func Register(name string, f Formatter) {
	Registry[name] = f
}

// Get obtém um formatter pelo nome canônico.
func Get(name string) (Formatter, bool) {
	f, ok := Registry[name]
	return f, ok
}

// Extensions retorna todas as extensões registradas (com ponto).
func Extensions() []string {
	exts := make([]string, 0, len(Registry))
	for _, f := range Registry {
		exts = append(exts, f.Extension())
	}
	return exts
}