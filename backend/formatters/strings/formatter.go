// StringsFormatter e StringsNoDedupFormatter: os formatters do pacote no
// registry de formatters (chaves "strings" e "strings-no-dedup"). O
// byte-level vive em strings.go/parse.go (Marshal/Unmarshal); a camada de
// ARQUIVO padronizada (gravar/lêr por instância) fica em
// formatters.NewWriter/NewReader — nada de métodos Write*/Read* por kind
// aqui: naming é decisão do chamador (helpers *JSONPath do pacote json +
// formatters.SwapExt).
package strings

import (
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters"
)

// StringsFormatter formata a Collection textual para o formato Strings.
//
// Formato java-like, uma linha por (row, idioma):
// <versão>:<id>[:<name>]:<index>:<lang>║$<hash> = <literal|$hash>,
// agrupadas por entrada sob /*key=<key> row_count=<n>*/.
// Cada formato futuro organiza esse DTO à sua maneira, sem reaproveitar
// structs de outro formato.
type StringsFormatter struct{}

// Extension retorna a extensão produzida por este formatter.
func (StringsFormatter) Extension() string {
	return extensionStrings
}

// Marshal organiza a Collection pronta no texto de saída (todos os idiomas).
// Recebe apenas DTO pronto; não toca no domínio.
func (StringsFormatter) Marshal(c dto.Collection) ([]byte, error) {
	return Marshal(c)
}

// MarshalLangs organiza a Collection pronta contendo só os idiomas pedidos
// (nil/vazio = todos).
func (StringsFormatter) MarshalLangs(c dto.Collection, langs []string) ([]byte, error) {
	return MarshalLangs(c, langs)
}

// Unmarshal parseia o texto de volta para a Collection (DTO),
// com rows ordenados por Index para reconstrução posicional do binário.
func (StringsFormatter) Unmarshal(data []byte) (dto.Collection, error) {
	return Unmarshal(data)
}

// StringsNoDedupFormatter serializa SEM o dedup de conteúdo: todo texto
// sai literal, inclusive os que casariam como repetição (não viram ref
// "$hash"). É a direção "dupe/resolve" do menu/CLI: expandir os hashes de
// conteúdo de volta ao texto completo. O campo de identidade (hash "$hex"
// da row) permanece em qualquer modo.
type StringsNoDedupFormatter struct{}

// Extension retorna a extensão produzida por este formatter.
func (StringsNoDedupFormatter) Extension() string {
	return extensionStrings
}

// Marshal serializa SEM dedup (todo texto literal, todos os idiomas).
func (StringsNoDedupFormatter) Marshal(c dto.Collection) ([]byte, error) {
	return MarshalNoDedup(c)
}

// MarshalLangs serializa SEM dedup, só com os idiomas pedidos
// (nil/vazio = todos).
func (StringsNoDedupFormatter) MarshalLangs(c dto.Collection, langs []string) ([]byte, error) {
	return MarshalNoDedupLangs(c, langs)
}

// Unmarshal parseia o texto de volta para a Collection (DTO) — o mesmo
// parse do StringsFormatter (as refs "$hash" saem expandidas).
func (StringsNoDedupFormatter) Unmarshal(data []byte) (dto.Collection, error) {
	return Unmarshal(data)
}

func init() {
	formatters.Register("strings", StringsFormatter{})
	formatters.Register("strings-no-dedup", StringsNoDedupFormatter{})
}
