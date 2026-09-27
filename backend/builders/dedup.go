// Dedup de textos idênticos (refs "$hash"): regra única usada na entrega
// do DTO ao frontend (view oculta refs) e espelhada na serialização dos
// artefatos (marshalCollectionLangs / MarshalLangs). O store nunca guarda
// dedup — é sempre uma cópia aplicada na borda: entrega, export ou apply.
package builders

import (
	"unicode/utf8"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/hash"
)

// DedupDTO devolve uma cópia da Collection com o dedup de textos 'us'
// idênticos aplicado — a mesma regra dos artefatos (hash.MinDedupRunes
// runes, XXH64 byte-exato; 1ª ocorrência em ordem de chave permanece
// literal = def, demais viram "$hash" = ref). Idiomas ≠ us nunca viram
// ref. A entrada não é mutada; hash e rows ficam intactos.
//
// O escopo do dedup é a Collection recebida: o view passa a versão
// inteira (dedup global) e o export granular passa só as entradas
// pedidas — nunca há refs cuja def esteja fora do payload/artefato.
func DedupDTO(c dto.Collection) dto.Collection {
	seen := make(map[string]bool)
	out := make(dto.Collection, len(c))
	for _, k := range c.SortedKeys() {
		entry := c[k]
		rows := make([]dto.TextRow, len(entry.Rows))
		copy(rows, entry.Rows)
		dto.SortRows(rows)
		for i := range rows {
			r := &rows[i]
			t := r.Text[common.DefaultLocalization]
			h := r.Hash[common.DefaultLocalization]
			if t == "" || h == "" || utf8.RuneCountInString(t) < hash.MinDedupRunes {
				continue
			}
			if !seen[h] {
				seen[h] = true
				continue
			}
			text := make(map[string]string, len(r.Text))
			for lk, lv := range r.Text {
				text[lk] = lv
			}
			text[common.DefaultLocalization] = hash.Prefix(h)
			r.Text = text
		}
		entry.Rows = rows
		out[k] = entry
	}
	return out
}

// refBare devolve o hash por trás de uma ref "$hash" da própria row.
// Espelha a segurança do unmarshal do JSON/strings: prefixo `$` só conta
// se casar com o hash PRÓPRIO da row (literal que comece com `$` passa
// reto — o texto vence o hash).
func refBare(row dto.TextRow, text string) (string, bool) {
	bare, ok := hash.Strip(text)
	if !ok || bare != row.Hash[common.DefaultLocalization] {
		return "", false
	}
	return bare, true
}
