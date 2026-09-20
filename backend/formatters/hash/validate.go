package hash

import (
	"fmt"
	"sort"

	"ffxresources/backend/dto"
)

// TODO: deletar quando colisão xxHash64 for considerada segura —
// guarda temporária de desencargo: reprova DTO com mesmo hash para textos diferentes.
// Dedup legítimo (mesmo hash + mesmo texto) continua válido.
func ValidateNoCollision(c dto.Collection) error {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		entry := c[key]
		// por idioma: hash -> primeiro texto visto
		seen := make(map[string]map[string]string)
		for _, row := range entry.Rows {
			for lang, h := range row.Hash {
				text := row.Text[lang]
				if text == "" {
					continue
				}
				byHash, ok := seen[lang]
				if !ok {
					byHash = make(map[string]string)
					seen[lang] = byHash
				}
				if first, exists := byHash[h]; exists {
					if first != text {
						return fmt.Errorf(
							"hash collision: file %q lang %q hash %s: %q (row %d) vs %q",
							key, lang, h, first, row.Index, text,
						)
					}
					continue
				}
				byHash[h] = text
			}
		}
	}
	return nil
}

// DedupIndex agrupa índices por hash para um idioma, dentro de um arquivo/lote.
// Permite dedup máximo por arquivo por lote e, no import, localizar a frase
// original pelo hash para comparar tags de controle.
func DedupIndex(rows []dto.TextRow, lang string) map[string][]int {
	out := make(map[string][]int)
	for _, row := range rows {
		h, ok := row.Hash[lang]
		if !ok || h == "" {
			continue
		}
		out[h] = append(out[h], row.Index)
	}
	return out
}
