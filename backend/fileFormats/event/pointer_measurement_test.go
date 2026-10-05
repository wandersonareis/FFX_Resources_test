package event_test

import (
	"encoding/binary"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ffxresources/backend/fileFormats/objectsfile"
)

/*
Fase 0 — medir antes de decidir.

O formato de event tem o MESMO ponto cego do objectsfile: o ref (uint16 no
cabeçalho de 8 bytes por campo) vira consulta direta em
converter.GetStringBytesAtLookupOffset, que só trata FORA da tabela — nunca
testa se o offset cai na fronteira de um bloco. Um offset no meio de outro
texto devolve o SUFIXO dele.

Este teste não guarda nada: só conta, em todos os binários reais, quantos
refs caem no meio de um bloco ou fora do arquivo. É a medição que decide se
a guarda entra aqui.
*/

// raizDados sobe do diretório do pacote até achar build/bin/data (ou, na
// falta dele, testData) — o mesmo achado que os testes de chunkmap usam.
func raizDados(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("wd: %v", err)
	}
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "build", "bin", "data")); err == nil {
			return filepath.Join(dir, "build", "bin", "data")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip("build/bin/data não encontrado — sem dado real para medir")
	return ""
}

// medirEvent classifica TODOS os refs (regular e simplified) de cada
// binário de strings contra o próprio arquivo, que é a string table do
// formato.
func medirEvent(root string) (arquivos, refs, mid, oob int) {
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.EqualFold(filepath.Ext(d.Name()), ".bin") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil || len(data) < 2 {
			return nil
		}
		first := binary.LittleEndian.Uint16(data)
		// Mesmos testes de sanidade que FromFieldStringData faz: fora
		// disso não é arquivo deste formato.
		if first%8 != 0 || first < 8 || int(first) > len(data) {
			return nil
		}
		arquivos++
		for i := 0; i < int(first)/8; i++ {
			base := i * 8
			for _, pos := range []int{base, base + 4} {
				refs++
				switch objectsfile.PointerStatusAt(data, int(binary.LittleEndian.Uint16(data[pos:pos+2]))) {
				case objectsfile.PointerMid:
					mid++
				case objectsfile.PointerOOB:
					oob++
				}
			}
		}
		return nil
	})
	return
}

func TestMedeRefsQuebradosDeEvent(t *testing.T) {
	data := raizDados(t)

	for _, c := range []struct {
		jogo string
		rel  string
	}{
		{"FFX", filepath.Join("ffx_ps2", "ffx", "master", "new_uspc", "event")},
		{"FFX-2", filepath.Join("ffx_ps2", "ffx2", "master", "new_uspc", "event")},
	} {
		arquivos, refs, mid, oob := medirEvent(filepath.Join(data, c.rel))
		fmt.Printf("EVENT   %-6s arquivos=%-5d refs=%-6d mid=%-5d oob=%-5d\n",
			c.jogo, arquivos, refs, mid, oob)

		// A medição só significa alguma coisa se varreu dado de verdade:
		// sem arquivos ou sem refs, ou o formato mudou ou sumiu o dado e
		// a decisão "a guarda entra (ou não) aqui" foi tomada no vazio.
		if arquivos == 0 || refs == 0 {
			t.Errorf("%s: medição vazia (%d arquivos, %d refs)", c.jogo, arquivos, refs)
		}
	}
}

/*
Resultado da medição (build/bin/data, new_uspc):

	FFX    arquivos=318   refs=35738  mid=168  oob=0
	FFX-2  arquivos=1060  refs=149278 mid=4970 oob=0

Há ref quebrado (mid > 0) e a forma bate com a do objectsfile: este leitor
é candidato a guarda. O que existe hoje é só o bounds-check (FORA da tabela
→ nil); o de FRONTEIRA (meio de um bloco → sufixo de outro texto) não, e é
por aí que ainda entra frase errada.
*/
