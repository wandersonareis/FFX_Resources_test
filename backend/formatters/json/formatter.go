// JSONFormatter e JSONNoDedupFormatter são os ÚNICOS formatters JSON do
// pacote: os bytes são idênticos para todo kind (o motor vive em
// shared.go) — kind só diferencia o NOME do artefato, que é decisão do
// chamador (helpers EventsJSONPath/ObjectsJSONPath/MacroJSONPath/
// HelpJSONPath). Ambos se registram no registry (formatters.Register)
// pelas chaves "json" e "json-no-dedup".
package json

import (
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters"
)

// JSONFormatter serializa/deserializa a Collection no formato JSON de
// artefato: mapa raiz por entrada, metadata, rows com hash "$hex" e dedup
// de conteúdo ($hash) no idioma default.
type JSONFormatter struct{}

// Extension retorna a extensão produzida por este formatter.
func (JSONFormatter) Extension() string {
	return extensionJSON
}

// Marshal organiza a Collection pronta no JSON de saída (todos os idiomas).
// Recebe apenas DTO pronto; não toca no domínio.
func (JSONFormatter) Marshal(c dto.Collection) ([]byte, error) {
	return marshalCollection(c)
}

// MarshalLangs organiza a Collection pronta contendo só os idiomas pedidos
// (nil/vazio = todos).
func (JSONFormatter) MarshalLangs(c dto.Collection, langs []string) ([]byte, error) {
	return marshalCollectionLangs(c, langs)
}

// Unmarshal parseia o JSON de volta para a Collection (DTO), com rows
// ordenados por Index para reconstrução posicional do binário.
func (JSONFormatter) Unmarshal(data []byte) (dto.Collection, error) {
	return unmarshalCollection(data)
}

// JSONNoDedupFormatter serializa SEM o dedup de conteúdo: todo texto do
// idioma default sai literal, inclusive repetições (não colapsam em ref
// "$hash"). É a direção "dupe/resolve": expandir os hashes de conteúdo de
// volta ao texto completo. O hash de identidade da row ($hex) permanece.
type JSONNoDedupFormatter struct{}

// Extension retorna a extensão produzida por este formatter.
func (JSONNoDedupFormatter) Extension() string {
	return extensionJSON
}

// Marshal serializa SEM dedup (todo texto literal, todos os idiomas).
func (JSONNoDedupFormatter) Marshal(c dto.Collection) ([]byte, error) {
	return MarshalNoDedup(c)
}

// MarshalLangs serializa SEM dedup, só com os idiomas pedidos
// (nil/vazio = todos).
func (JSONNoDedupFormatter) MarshalLangs(c dto.Collection, langs []string) ([]byte, error) {
	return MarshalNoDedupLangs(c, langs)
}

// Unmarshal parseia o JSON de volta para a Collection (DTO) — o mesmo
// parse do JSONFormatter (as refs "$hash" saem expandidas).
func (JSONNoDedupFormatter) Unmarshal(data []byte) (dto.Collection, error) {
	return unmarshalCollection(data)
}

func init() {
	formatters.Register("json", JSONFormatter{})
	formatters.Register("json-no-dedup", JSONNoDedupFormatter{})
}
