package hash

import "unicode/utf8"

// MinDedupRunes é o tamanho mínimo (em runes) do TEXTO VISÍVEL para um texto
// repetido virar referência de dedup ($hash). A elegibilidade é ESTRITAMENTE
// maior que este valor (> MinDedupRunes), contado sobre o miolo devolvido
// por visibleCore: tags {...} e espaços/quebras de linha das BORDAS são
// ignorados. O objetivo do limite é excluir do dedup palavras/exclamações
// curtas em inglês que se repetem muito mas podem ter tradução diferente a
// cada contexto: "Whoa!", "What?!", "Yeah!", "No way!", "Let's go!",
// "Thank you." nunca viram ref — mesmo quando vêm embrulhadas em tags do
// falante/newline ("{PC:00:Tidus}{TEXT_NEWLINE}Whoa!", genk0100 do FFX).
// Com > 10 no miolo, só textos visíveis de 11+ runes colapsam (frases
// completas, tipicamente tradução única).
// A contagem é em runes para não miscaracterizar CJK (ex.: "你好" tem 2 runes
// mas 6 bytes). Regra única: formatters JSON/strings/CSV/XLSX e o dedup do
// DTO de help.
const MinDedupRunes = 10

// IsDedupEligible informa se o texto pode virar referência de dedup:
// o miolo visível (visibleCore) tem mais de MinDedupRunes runes. É a régua
// única usada pelos formatters JSON/strings/CSV/XLSX e pelo dedup do DTO —
// nunca compare MinDedupRunes diretamente com >=, senão partículas de
// exatamente MinDedupRunes runes ("Whoa!" era dedupado em genk0100/genk1000)
// colapsam indevidamente.
func IsDedupEligible(s string) bool {
	return utf8.RuneCountInString(visibleCore(s)) > MinDedupRunes
}

// visibleCore devolve o miolo VISÍVEL do texto: sem as tags {...} nem os
// espaços/quebras de linha das bordas. Nas duas pontas, em loop, consome
// qualquer quantidade de tokens {..} intercalados com espaços/`\n`/`\r`/`\t`
// — o varredor espelha o de converter.tagCounts (tags.go): '{' sem '}' de
// fechamento é literal e encerra o trim. Tags no MEIO do texto permanecem
// no miolo e contam como runes: "Foo{TEXT_NEWLINE}Bar" é uma frase de duas
// linhas, conteúdo real.
//
// É a base da elegibilidade do dedup: "{PC:00:Tidus}{TEXT_NEWLINE}Whoa!"
// tem ~30 runes crus, mas o miolo é "Whoa!" (5) — partícula curta, nunca
// vira ref, aconteça o número de tags que for na frente.
func visibleCore(s string) string {
	r := []rune(s)
	lo, hi := 0, len(r)

	for trimEdge(r, &lo, &hi) {
	}

	return string(r[lo:hi])
}

// trimEdge consume uma rodada de bordas: espaços e uma tag no início,
// espaços e uma tag no fim. Devolve true se consumiu algo (o chamador
// repete até estabilizar — tags e espaços podem vir em qualquer ordem,
// qualquer quantidade).
func trimEdge(r []rune, lo *int, hi *int) bool {
	changed := false

	// Borda inicial: run de espaços/quebras.
	for *lo < *hi && isEdgeBlank(r[*lo]) {
		*lo++
		changed = true
	}
	// Borda inicial: uma tag {..} completa.
	if *lo < *hi && r[*lo] == '{' {
		if end := indexRune(r, *lo+1, *hi, '}'); end >= 0 {
			*lo = end + 1
			changed = true
		}
	}

	// Borda final: run de espaços/quebras.
	for *hi > *lo && isEdgeBlank(r[*hi-1]) {
		*hi--
		changed = true
	}
	// Borda final: uma tag {..} completa — o '}' do fim casa com um '{'
	// sem chaves entre eles; senão é '}' literal e o trim para.
	if *hi > *lo && r[*hi-1] == '}' {
		if start := lastIndexRune(r, *lo, *hi-1, '{'); start >= 0 &&
			!containsRune(r, start+1, *hi-1, '{', '}') {
			*hi = start
			changed = true
		}
	}

	return changed
}

// isEdgeBlank classifica o rune como espaçamento descartável de borda.
func isEdgeBlank(r rune) bool {
	switch r {
	case ' ', '\n', '\r', '\t':
		return true
	}
	return false
}

// indexRune devolve a primeira posição de q em r[from:to), ou -1.
func indexRune(r []rune, from, to int, q rune) int {
	for i := from; i < to; i++ {
		if r[i] == q {
			return i
		}
	}
	return -1
}

// lastIndexRune devolve a última posição de q em r[from:to), ou -1.
func lastIndexRune(r []rune, from, to int, q rune) int {
	for i := to - 1; i >= from; i-- {
		if r[i] == q {
			return i
		}
	}
	return -1
}

// containsRune informa se algum dos runes qs aparece em r[from:to).
func containsRune(r []rune, from, to int, qs ...rune) bool {
	for i := from; i < to; i++ {
		for _, q := range qs {
			if r[i] == q {
				return true
			}
		}
	}
	return false
}
