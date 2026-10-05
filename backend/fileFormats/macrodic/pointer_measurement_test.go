package macrodic_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/macrodic"
	"ffxresources/backend/fileFormats/objectsfile"
)

/*
Fase 0 — medir antes de decidir.

macrodic tem a MESMA forma do objectsfile: a tabela de offsets (pares uint16
por chunk) vira consulta direta em common.GetStringBytesAtLookupOffset, que
só trata FORA do arquivo — nunca testa se o offset cai na fronteira de um
bloco.

Só contagem, aqui: quantos refs caem no meio de um bloco ou fora do chunk.
É a medição que decide se a guarda entra neste leitor.
*/

// raizDados sobe do diretório do pacote até achar build/bin/data (ou
// testData, na falta dele).
func raizDados(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("wd: %v", err)
	}
	for i := 0; i < 8; i++ {
		for _, cand := range []string{
			filepath.Join(dir, "build", "bin", "data"),
			filepath.Join(dir, "testData"),
		} {
			if _, err := os.Stat(cand); err == nil {
				return cand
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip("build/bin/data/testData não encontrado — sem dado real para medir")
	return ""
}

// medirMacrodic abre um macrodic.dcp, fatia os chunks pela tabela de offsets
// e classifica cada Name/SimplifiedName contra o chunk de onde veio.
func medirMacrodic(path string, version common.GameVersion) (chunks, refs, mid, oob int) {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return
	}
	container, err := macrodic.NewMacroDictionaryBinaryFileFromBytes(
		data, common.DefaultLocalization, version)
	if err != nil {
		fmt.Printf("MACRODIC %s: container ilegível: %v\n", filepath.Base(path), err)
		return
	}

	for i, start := range container.ChunkOffsets {
		end := uint32(len(container.Bytes))
		if i+1 < len(container.ChunkOffsets) {
			end = container.ChunkOffsets[i+1]
		}
		if start == 0 || uint64(end) > uint64(len(container.Bytes)) || start >= end {
			continue
		}
		textFile, err := macrodic.NewMacroDictionaryTextFile(
			container.Bytes[start:end], common.DefaultLocalization, version)
		if err != nil {
			// Chunk que não é arquivo de texto: não é ref algum.
			continue
		}
		chunks++
		for _, seg := range textFile.Segments {
			refs++
			switch objectsfile.PointerStatusAt(textFile.Bytes, int(seg.NameOffset)) {
			case objectsfile.PointerMid:
				mid++
			case objectsfile.PointerOOB:
				oob++
			}
			if seg.SimplifiedNameOffset == seg.NameOffset {
				continue
			}
			refs++
			switch objectsfile.PointerStatusAt(textFile.Bytes, int(seg.SimplifiedNameOffset)) {
			case objectsfile.PointerMid:
				mid++
			case objectsfile.PointerOOB:
				oob++
			}
		}
	}
	return
}

func TestMedeRefsQuebradosDeMacrodic(t *testing.T) {
	data := raizDados(t)

	for _, c := range []struct {
		jogo    string
		rel     string
		version common.GameVersion
	}{
		{"FFX", filepath.Join("ffx_ps2", "ffx", "master", "new_uspc", "menu", "macrodic.dcp"), common.GameVersionFFX},
		{"FFX-2", filepath.Join("ffx_ps2", "ffx2", "master", "new_uspc", "menu", "macrodic.dcp"), common.GameVersionFFX2},
	} {
		chunks, refs, mid, oob := medirMacrodic(filepath.Join(data, c.rel), c.version)
		fmt.Printf("MACRODIC %-6s chunks=%-3d refs=%-6d mid=%-5d oob=%-5d\n",
			c.jogo, chunks, refs, mid, oob)

		// Medição sem chunk ou sem ref não decide nada: ou o formato
		// mudou, ou sumiu o dado.
		if chunks == 0 || refs == 0 {
			t.Errorf("%s: medição vazia (%d chunks, %d refs)", c.jogo, chunks, refs)
		}
	}
}

/*
Resultado da medição (build/bin/data, new_uspc):

	FFX    chunks=2  refs=210 mid=2 oob=0
	FFX-2  chunks=4  refs=539 mid=4 oob=0

Há ref quebrado e a forma é a mesma do objectsfile (offset -> lookup sem
teste de fronteira), então este leitor também é candidato a guarda.
*/
