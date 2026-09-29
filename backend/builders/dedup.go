// Dedup de textos idênticos (refs "$hash"). O hash é PONTEIRO de dupe —
// nunca validador de texto: na entrega display ele aponta para o ORIGINAL
// de data/ (imutável, uma vez na leitura da árvore); nos fluxos RAW é o
// hash do texto do arquivo. O store nunca guarda dedup — sempre uma cópia
// aplicada na borda (entrega ou apply); a serialização dos artefatos tem
// a sua própria regra nos formatters.
package builders

import (
	"unicode/utf8"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/hash"
)

// DedupDTO é o dedup RAW (texto atual contra o texto da def): usado só
// onde não há original para comparar (coleções RAW, testes). O colapso
// exige que o texto atual da twin seja IGUAL ao da def — dupe sobre texto
// traduzido (diferente da def) é proibido por decisão: dedup economiza a
// carga do tradutor, e uma tradução divergente é carga real.
//
// Para a entrega display (GetEntry) use o par HashOrder + DedupDisplayDTO:
// o estado "traduzido" é avaliado contra Original (data/) e a 1ª ocorrência
// é resolvida na ordem da versão inteira, sem exigir a def no Collection.
func DedupDTO(c dto.Collection) dto.Collection {
	defs := make(map[string]string) // ponteiro bare -> texto atual da def
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
			if _, ok := defs[h]; !ok {
				defs[h] = t
				continue
			}
			// Dupe: só quando o texto atual é IGUAL ao da def (tradução
			// divergente sob o mesmo original é conteúdo próprio).
			if defs[h] != t {
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

// ResolveDedupRefs devolve uma cópia da Collection com as refs "$hash"
// expandidas: a tabela hash → texto é construída das PRÓPRIAS defs (rows
// literais) do collection — o mesmo mecanismo da 2ª passada do unmarshal
// JSON/strings. Ref sem def no collection permanece como está (identidade:
// o texto atual vence). A entrada não é mutada.
//
// É o passo de aplicação do lote vindo do frontend (view dedupado) nos
// formatos cujo apply não expande refs por conta própria (macro: o DTO
// mesclado vai direto para o rebuild).
func ResolveDedupRefs(c dto.Collection) dto.Collection {
	defs := make(map[string]string)
	for _, k := range c.SortedKeys() {
		for _, row := range c[k].Rows {
			t := row.Text[common.DefaultLocalization]
			if t == "" {
				continue
			}
			if _, isRef := refBare(row, t); isRef {
				continue
			}
			if h := row.Hash[common.DefaultLocalization]; h != "" {
				defs[h] = t
			}
		}
	}
	out := make(dto.Collection, len(c))
	for _, k := range c.SortedKeys() {
		entry := c[k]
		rows := make([]dto.TextRow, len(entry.Rows))
		copy(rows, entry.Rows)
		for i := range rows {
			r := &rows[i]
			bare, isRef := refBare(*r, r.Text[common.DefaultLocalization])
			if !isRef {
				continue
			}
			resolved, ok := defs[bare]
			if !ok {
				continue // ref órfã: identidade (mantém o texto atual)
			}
			text := make(map[string]string, len(r.Text))
			for lk, lv := range r.Text {
				text[lk] = lv
			}
			text[common.DefaultLocalization] = resolved
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

// HashOrder é a régua "1ª ocorrência = def" do dedup de display, versão
// inteira: para cada ponteiro Hash[us] de uma Collection RAW guarda a
// posição da PRIMEIRA ocorrência (ordem SortedKeys → SortRows, contador
// global de rows) e o texto dela.
//
// É construída sobre o estado CRU (sem refs, hashes = ponteiro do texto
// do arquivo). Para rows NÃO traduzidas mods-ponteiros == ponteiro do
// original (mesmo texto ⇒ mesmo XXH64), então a ordem é válida também no
// domínio display. Traduções registram o hash DELAS (não o do original):
// não poluem grupos — colisão exigiria duas frases com o mesmo conteúdo,
// e o ponteiro é endereçado por conteúdo.
type HashOrder struct {
	first map[string]int    // ponteiro bare -> posição global da 1ª ocorrência
	text  map[string]string // texto da 1ª ocorrência (dedup RAW: def literal)
	base  map[string]int    // key -> nº de rows antes dela na ordem canônica
}

// NewHashOrder percorre a Collection RAW em ordem canônica (SortedKeys →
// SortRows) e registra a 1ª ocorrência de cada ponteiro.
func NewHashOrder(c dto.Collection) *HashOrder {
	ho := &HashOrder{
		first: map[string]int{},
		text:  map[string]string{},
		base:  map[string]int{},
	}
	seq := 0
	for _, k := range c.SortedKeys() {
		ho.base[k] = seq
		entry := c[k]
		rows := make([]dto.TextRow, len(entry.Rows))
		copy(rows, entry.Rows)
		dto.SortRows(rows)
		for _, r := range rows {
			t := r.Text[common.DefaultLocalization]
			h := r.Hash[common.DefaultLocalization]
			if t != "" && h != "" {
				if _, seen := ho.first[h]; !seen {
					ho.first[h] = seq
					ho.text[h] = t
				}
			}
			seq++
		}
	}
	return ho
}

// Offset devolve o nº de rows que antecedem a entrada `key` na ordem
// canônica da coleção que originou o HashOrder.
func (ho *HashOrder) Offset(key string) (int, bool) {
	off, ok := ho.base[key]
	return off, ok
}

// firstOf devolve a posição global da 1ª ocorrência do ponteiro, se existe.
func (ho *HashOrder) RawFirst(h string) (int, bool) {
	first, ok := ho.first[h]
	return first, ok
}

// defText devolve o texto da 1ª ocorrência do ponteiro (o estado no build
// do RAW), para o modo RAW comparar o texto atual da twin contra a def.
func (ho *HashOrder) defText(h string) (string, bool) {
	t, ok := ho.text[h]
	return t, ok
}

// collapseInto escreve a ref "$h" no texto 'us' da row (cópia do map).
func collapseInto(r *dto.TextRow, h string) {
	text := make(map[string]string, len(r.Text))
	for lk, lv := range r.Text {
		text[lk] = lv
	}
	text[common.DefaultLocalization] = hash.Prefix(h)
	r.Text = text
}

// DedupDisplayDTO devolve uma cópia da entry (já merged com o Original de
// data/) com o dedup de DISPLAY aplicado: colapsa em ref "$ponteiro" a
// repetição de um ORIGINAL ainda NÃO traduzida (Text[us] == Original[us])
// que não seja a 1ª ocorrência do ponteiro na ordem global (base+i).
//
// Regras:
//   - Sem Original na row (divergência/sem contraparte) → nunca colapsa:
//     o domínio do ponteiro não pode ser validado contra data/, e o texto
//     segue visível.
//   - Célula traduzida (texto difere do original) → nunca colapsa: dupe
//     sobre texto já traduzido é proibido por decisão — é carga real de
//     revisão.
//   - Def pode estar traduzida: a ref aponta para o PONTEIRO do original e
//     o apply resolve defs[ponteiro] → texto atual da def (a tradução).
//
// base é o nº de rows que precedem a entrada na ordem canônica da versão
// (HashOrder.Offset). A entrada não é mutada.
func DedupDisplayDTO(entry dto.FileEntry, base int, ho *HashOrder) dto.FileEntry {
	if ho == nil {
		return entry
	}
	rows := make([]dto.TextRow, len(entry.Rows))
	copy(rows, entry.Rows)
	dto.SortRows(rows)

	for i := range rows {
		r := &rows[i]
		origUS := r.Original[common.DefaultLocalization]
		t := r.Text[common.DefaultLocalization]
		h := r.Hash[common.DefaultLocalization]
		if t == "" || h == "" {
			continue
		}
		// Elegibilidade por domínio: com Original prova-se não traduzida
		// por conteúdo (texto == original, dentro da régua de runes do
		// ORIGINAL); sem Original o estado não é aferível — o colapso
		// volta a ser por texto, igual ao da 1ª ocorrência (comportamento
		// pré-original).
		elegivel := false
		if origUS != "" {
			elegivel = t == origUS && utf8.RuneCountInString(origUS) >= hash.MinDedupRunes
		} else if def, ok := ho.text[h]; ok {
			elegivel = def == t && utf8.RuneCountInString(t) >= hash.MinDedupRunes
		}
		if !elegivel {
			continue
		}
		first, ok := ho.RawFirst(h)
		if !ok || base+i <= first {
			// 1ª ocorrência do ponteiro (ou ponteiro desconhecido): é a def
			// da entrada — permanece literal.
			continue
		}
		collapseInto(r, h)
	}
	entry.Rows = rows
	return entry
}
