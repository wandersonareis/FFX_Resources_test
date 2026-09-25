package ffxencoding

import (
	"fmt"

	"ffxresources/backend/common"
)

// PrepareCharset publica os mapas decode/encode do charset indicado a partir
// da tabela embutida em charset_tables.go — sem I/O, sem caminho de arquivo,
// sem fail-fast de carga. Erro apenas para charset/versão desconhecidos.
//
// A tabela é imutável (as originais não são modificáveis na prática: mudá-las
// exigiria rebuild dos font atlases), então a carga é determinística e
// idempotente — chamar de novo apenas sobrescreve os mapas.
func PrepareCharset(version common.GameVersion, charset string) error {
	// lastmiss reaproveita os mapas de ffx2 (mesma árvore de encoding).
	version = common.CharsetVersion(version)
	table, ok := getTable(version, charset)
	if !ok {
		return fmt.Errorf("charset %q: tabela embutida não encontrada para a versão %v", charset, version)
	}

	byteToChar, charToByte := buildMappings([]rune(table), charset)
	SetCharMap(version, charset, byteToChar, charToByte)
	return nil
}

// buildMappings constrói os mapas do charset.
//
//	decode (byteToChar): todos os slots, repetidos incluídos — muitos-para-um.
//	encode (charToByte): 1 slot canônico por rune (1ª ocorrência) — uma
//	  escrita por chave, nunca sobrescreve.
//
// Slots repetidos: decodificam o rune da tabela normalmente, mas o encode
// canônico os apontaria para outro byte. O decode do converter detecta essa
// divergência (byte canônico != byte lido) e envolve o slot no token
// {PUA:XX:CHAR}, que preserva o byte exato no round-trip — o font atlas pode
// desenhar glifos distintos por slot mesmo quando a tabela os mapeia ao
// mesmo rune (ex.: Œ/œ onde a tabela diz espaço). Ver converter.ParseCommand.
func buildMappings(runes []rune, charset string) (map[uint]rune, map[rune]uint) {
	byteToChar := make(map[uint]rune, len(runes))
	charToByte := make(map[rune]uint, len(runes))
	dups := 0

	for i, r := range runes {
		idx := uint(0x30 + i)
		byteToChar[idx] = r

		if _, exists := charToByte[r]; exists {
			dups++
			if common.IsVerboseMode() {
				fmt.Printf("[charset %s] REPETIDO: %q em 0x%02X e 0x%02X → encode usa 0x%02X, decode emite {PUA:%X:%c}\n",
					charset, r, charToByte[r], idx, charToByte[r], idx, r)
			}
			continue
		}
		charToByte[r] = idx
	}

	if common.IsVerboseMode() {
		fmt.Printf("[charset %s] resumo: %d slots, %d runes únicos, %d repetidos\n",
			charset, len(runes), len(charToByte), dups)
	}
	return byteToChar, charToByte
}
