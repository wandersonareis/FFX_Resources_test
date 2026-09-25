package lockit

import "unicode/utf8"

// splitRecords separa os bytes de um arquivo lockit em registros.
//
// O separador é CRLF (LF também é aceito). O terminador final é registrado em
// terminated: true quando o arquivo termina com quebra (caso dos arquivos US),
// false quando o último registro não tem quebra.
func splitRecords(data []byte) (records [][]byte, terminated bool) {
	if len(data) == 0 {
		return nil, false
	}
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] != '\n' {
			continue
		}
		end := i
		if end > start && data[end-1] == '\r' {
			end--
		}
		records = append(records, data[start:end])
		start = i + 1
	}
	if start < len(data) {
		// Último registro sem quebra final.
		records = append(records, data[start:])
		return records, false
	}
	// Arquivo termina com quebra: o pedaço após o último LF é vazio.
	return records, true
}

// joinRecords monta o arquivo de volta a partir dos registros codificados.
func joinRecords(encoded [][]byte, terminated bool) []byte {
	total := 0
	for _, r := range encoded {
		total += len(r) + 2 // CRLF
	}
	if !terminated && len(encoded) > 0 {
		total -= 2
	}
	out := make([]byte, 0, total)
	for i, r := range encoded {
		out = append(out, r...)
		if i < len(encoded)-1 || terminated {
			out = append(out, '\r', '\n')
		}
	}
	return out
}

// vote classifica um registro bruto em um voto de codificação:
//
//	+1 = game, -1 = utf8, 0 = indefinido.
//
// O byte 0x20 (espaço ASCII) nunca aparece no charset do jogo (o espaço do
// jogo é 0x3A), então sua presença é sinal forte de UTF-8. Fora isso, 0x3A ou
// não ser UTF-8 válido indica game.
func vote(raw []byte) int {
	if len(raw) == 0 {
		return 0
	}
	hasSpace := false
	hasGameSpace := false
	for _, b := range raw {
		if b == 0x20 {
			hasSpace = true
		}
		if b == 0x3A {
			hasGameSpace = true
		}
	}
	if hasSpace {
		return -1
	}
	if hasGameSpace || !utf8.Valid(raw) {
		return 1
	}
	return 0
}

// detectKinds determina a codificação de cada registro combinando os votos de
// todos os idiomas disponíveis (o layout é idêntico entre eles) e preenchendo
// as lacunas por vizinhança (as seções são runs contíguos).
//
// n é o número de registros; rawByLang mapeia idioma -> registros brutos.
func detectKinds(rawByLang map[string][][]byte, n int) []Kind {
	votes := make([]int8, n) // +1 game, -1 utf8, 0 indefinido
	for i := 0; i < n; i++ {
		game, utf8Votes := 0, 0
		for _, recs := range rawByLang {
			if i >= len(recs) {
				continue
			}
			switch vote(recs[i]) {
			case 1:
				game++
			case -1:
				utf8Votes++
			}
		}
		switch {
		case game > utf8Votes:
			votes[i] = 1
		case utf8Votes > game:
			votes[i] = -1
		}
	}
	// Preenchimento para frente e para trás.
	last := int8(0)
	for i := range votes {
		if votes[i] != 0 {
			last = votes[i]
		} else if last != 0 {
			votes[i] = last
		}
	}
	next := int8(0)
	for i := len(votes) - 1; i >= 0; i-- {
		if votes[i] != 0 {
			next = votes[i]
		} else if next != 0 {
			votes[i] = next
		}
	}
	kinds := make([]Kind, n)
	for i, v := range votes {
		if v == 1 {
			kinds[i] = KindGame
		} else {
			// Indefinido (arquivo todo sem sinal) cai para UTF-8: ASCII puro
			// é seguro nos dois sentidos.
			kinds[i] = KindUTF8
		}
	}
	return kinds
}
