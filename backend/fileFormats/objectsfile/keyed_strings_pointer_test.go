package objectsfile_test

import (
	"bytes"
	"testing"

	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/models"
)

/*
A regra central da guarda, testada isoladamente da leitura de arquivo.

O ref é DIREÇÃO, não fato: quem decide se há texto é a FRONTEIRA do bloco
(o byte anterior ser terminador), nunca o valor do offset nem a ordem dos
ponteiros. Um offset no meio de um texto vira o sufixo de outro — e sufixo
de outro texto jamais é frase.
*/

// TestPointerStatusAtBoundary fixa a classificação byte a byte contra uma
// string table real:
//
//	"Hi\0Bye\0\0Go\0"
//	 0  1  2  3  4  5  6  7  8  9 10
//
// Nada aqui é decodificado: só a posição.
func TestPointerStatusAtBoundary(t *testing.T) {
	table := []byte("Hi\x00Bye\x00\x00Go\x00")

	cases := []struct {
		name   string
		offset int
		want   objectsfile.PointerStatus
	}{
		{"fronteira com texto", 0, objectsfile.PointerOK},
		{"meio do texto anterior", 1, objectsfile.PointerMid},
		{"terminador", 2, objectsfile.PointerEmpty},
		{"fronteira do próximo texto", 3, objectsfile.PointerOK},
		{"meio do texto (y)", 4, objectsfile.PointerMid},
		{"meio do texto (e)", 5, objectsfile.PointerMid},
		{"terminador", 6, objectsfile.PointerEmpty},
		{"terminador duplo aponta para terminador", 7, objectsfile.PointerEmpty},
		{"fronteira depois do terminador duplo", 8, objectsfile.PointerOK},
		{"meio do texto curto", 9, objectsfile.PointerMid},
		{"último byte é terminador", 10, objectsfile.PointerEmpty},
		{"um além do fim", 11, objectsfile.PointerOOB},
		{"muito além do fim", 1000, objectsfile.PointerOOB},
		{"offset negativo", -1, objectsfile.PointerOOB},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := objectsfile.PointerStatusAt(table, c.offset)
			if got != c.want {
				t.Errorf("PointerStatusAt(table, %d) = %s, quero %s", c.offset, got, c.want)
			}
		})
	}

	// Tabela vazia: não existe fronteira alguma — tudo é OOB, inclusive 0.
	for _, off := range []int{-1, 0, 1} {
		if got := objectsfile.PointerStatusAt(nil, off); got != objectsfile.PointerOOB {
			t.Errorf("PointerStatusAt(nil, %d) = %s, quero oob", off, got)
		}
	}
}

// TestPointerStatusMonotonicoNaoBasta é o caso que derrubaria uma guarda
// feita por ordem dos ponteiros: 0,6,10,13 é estritamente crescente e ainda
// assim 6 e 10 caem no meio de um bloco (é a row real de name_txt, onde 10
// cai dentro de "-----"). O teste falha em dobro se alguém trocar a regra
// de fronteira pela monotonicidade.
func TestPointerStatusMonotonicoNaoBasta(t *testing.T) {
	table := []byte("AAAAAAAAAAAAA") // 13 bytes, nenhum terminador
	offsets := []int{0, 6, 10, 13}

	want := []objectsfile.PointerStatus{
		objectsfile.PointerOK,
		objectsfile.PointerMid,
		objectsfile.PointerMid,
		objectsfile.PointerOOB,
	}

	for i := 1; i < len(offsets); i++ {
		if offsets[i] <= offsets[i-1] {
			t.Fatalf("o caso só prova alguma coisa se for estritamente crescente: %v", offsets)
		}
	}

	broken := 0
	for i, off := range offsets {
		got := objectsfile.PointerStatusAt(table, off)
		if got != want[i] {
			t.Errorf("offset %d = %s, quero %s", off, got, want[i])
		}
		if got.IsBroken() {
			broken++
		}
	}
	if broken != 3 {
		t.Errorf("refs quebrados = %d, quero 3 (6, 10 e 13)", broken)
	}
}

// TestPointerStatusIsBrokenSomenteMidEOOB: EMPTY é ponteiro legítimo para
// um terminador ("não tem texto" de direção certa), não é ref quebrado.
// QUEBRAR é só cair no meio de um bloco ou fora da tabela.
func TestPointerStatusIsBrokenSomenteMidEOOB(t *testing.T) {
	cases := []struct {
		status objectsfile.PointerStatus
		want   bool
	}{
		{objectsfile.PointerCreated, false},
		{objectsfile.PointerOK, false},
		{objectsfile.PointerEmpty, false},
		{objectsfile.PointerMid, true},
		{objectsfile.PointerOOB, true},
	}

	for _, c := range cases {
		if got := c.status.IsBroken(); got != c.want {
			t.Errorf("IsBroken(%s) = %v, quero %v", c.status, got, c.want)
		}
	}
}

// TestUsableStringTableBaseValidaReproducao fixa o contrato da base do
// rebuild: a string table original só é aceita quando ela reproduz TODOS
// os refs lidos — mesmo offset, mesmos bytes e mesma classificação,
// inclusive os MID/OOB, que têm que continuar quebrados contra ela.
//
// NÃO depende de haver ref quebrado: o splice usa a base no arquivo limpo
// também (sem edição o delta é zero em todo ponto e a saída é o próprio
// base byte a byte). O que rejeita a base é ela NÃO reproduzir os refs.
func TestUsableStringTableBaseValidaReproducao(t *testing.T) {
	if err := ffxencoding.PrepareVersionCharsets(common.GameVersionFFX); err != nil {
		t.Fatalf("prepare charsets: %v", err)
	}

	base := []byte("Hi\x00Bye\x00\x00Go\x00")
	seg := func(off int) models.Segment {
		return models.Segment{Offset: models.Offset(off), Key: models.Key(off + 1)}
	}
	read := func(table []byte, offsets ...int) []datastore.IGlobalKeyedString {
		out := make([]datastore.IGlobalKeyedString, 0, len(offsets))
		for _, off := range offsets {
			out = append(out, objectsfile.NewKeyedString(
				common.DefaultLocalization, seg(off), table, common.GameVersionFFX))
		}
		return out
	}

	// Arquivo com ref quebrado: 0 OK, 1 MID e 13 OOB.
	broken := read(base, 0, 1, 13)
	if got := objectsfile.UsableStringTableBase(broken, base); !bytes.Equal(got, base) {
		t.Errorf("a tabela do próprio arquivo tem que ser aceita quando há ref quebrado (got %d bytes)", len(got))
	}

	// Arquivo limpo: a base também é aceita — é ela que garante o
	// deslocamento global e a saída byte idêntica sem edição.
	clean := read(base, 0, 2, 3)
	if got := objectsfile.UsableStringTableBase(clean, base); !bytes.Equal(got, base) {
		t.Errorf("a tabela do próprio arquivo foi rejeitada num arquivo limpo (got %d bytes)", len(got))
	}

	// Sem candidato não há base.
	if got := objectsfile.UsableStringTableBase(broken, nil); got != nil {
		t.Errorf("base nula foi aceita")
	}

	// Candidato que não reproduz o ref OK (mesmo offset, outro texto).
	tampered := []byte("XX\x00Bye\x00\x00Go\x00")
	if got := objectsfile.UsableStringTableBase(broken, tampered); got != nil {
		t.Errorf("base que não reproduz os bytes do ref OK foi aceita")
	}

	// Candidato que não reproduz a fronteira de um ref EMPTY: o offset 6
	// deixou de cair em terminador (o byte virou 'X'). O ref OK continua
	// idêntico — só a classificação muda, e a base cai mesmo assim.
	shifted := []byte("Hi\x00ByeX\x00Go\x00")
	emptyRef := read(base, 0, 6)
	if got := objectsfile.UsableStringTableBase(emptyRef, shifted); got != nil {
		t.Errorf("base que não reproduz a fronteira do ref EMPTY foi aceita")
	}

	// Tabela de outro arquivo, só que maior: o OOB deixa de ser OOB e a
	// base não reproduz mais a classificação.
	foreign := append(append([]byte{}, base...), "extra\x00"...)
	if got := objectsfile.UsableStringTableBase(broken, foreign); got != nil {
		t.Errorf("base estranha (OOB deixou de ser OOB) foi aceita")
	}
}
