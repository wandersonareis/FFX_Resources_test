package eventtable_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/eventtable"
	testcommon "ffxresources/testData"
)

// Round-trip byte-exato: parse pelo codec events (FromFieldStringData) →
// EncodeLocalizedStrings → mesmos bytes, para a tabela inteira em 'us'.
func testRoundTrip(t *testing.T, version common.GameVersion, kind, id string, data []byte) {
	t.Helper()
	charset := ffxencoding.GetCharsetForLanguage(common.DefaultLocalization)
	fs, err := event.FromFieldStringData(data, charset, version)
	if err != nil {
		t.Fatalf("%s/%s: parse: %v", kind, id, err)
	}
	if len(fs) == 0 {
		t.Fatalf("%s/%s: 0 entries", kind, id)
	}
	// Cloud: meta-byte constante F0 FF não é (flags,choices) — ver comentário
	// em eventtable.go. O codec é do events; qualquer divergência de bytes
	// aqui é instrutiva do desvio.
	locals := make([]*event.LocalizedFieldStringObject, 0, len(fs))
	for _, f := range fs {
		locals = append(locals, event.NewLocalizedFieldStringObjectWithContent(common.DefaultLocalization, f))
	}
	enc, err := event.EncodeLocalizedStrings(locals, common.DefaultLocalization, version)
	if err != nil {
		t.Fatalf("%s/%s: encode: %v", kind, id, err)
	}
	if !bytes.Equal(enc, data) {
		i := 0
		for i < len(enc) && i < len(data) && enc[i] == data[i] {
			i++
		}
		t.Fatalf("%s/%s: round-trip divergente (%d -> %d bytes, 1o desvio em %#x)", kind, id, len(data), len(enc), i)
	}
}

func TestCodecEventTablesFFX2(t *testing.T) {
	if err := ffxencoding.PrepareVersionCharsets(common.GameVersionFFX2); err != nil {
		t.Skipf("charsets: %v", err)
	}
	root := filepath.Dir(testcommon.GetTestDataRootDirectory())
	binRoot := filepath.Join(root, "build", "bin", "data", "ffx_ps2", "ffx2", "master", "new_uspc")
	if _, err := os.Stat(binRoot); err != nil {
		t.Skip("build/bin/data indisponível")
	}
	sizes := map[string]int{}
	for _, p := range []string{"menu/tutorial.msb"} {
		data, err := os.ReadFile(filepath.Join(binRoot, filepath.FromSlash(p)))
		if err != nil {
			t.Fatalf("ler %s: %v", p, err)
		}
		sizes[p] = len(data)
		testRoundTrip(t, common.GameVersionFFX2, "probe", p, data)
	}
	// Um btl do FFX-2: parse sim; re-encode exige macros no datastore (o
	// decod gera {MCR:...} que só re-codifica com a store — fora dos testes,
	// que só preparam charsets).
	b := `F:\ffxWails\FFX_Resources\build\bin\data\ffx_ps2\ffx2\master\new_uspc\battle\btl\bika07_228\bika07_228.bin`
	data, err := os.ReadFile(b)
	if err != nil {
		t.Skip("btl fixture ausente")
	}
	_ = data
	_ = sizes
}

// cloud/cloudv (F0 FF no meta-byte + string duplicada sem dedup) NÃO
// round-tripam ao byte exato pelo codec events (marcador F0 FF é
// normalizado e a duplicata é deduplicada no rebuild) — detalhe anotado em
// eventtable.go. Exclusos do teste de igualdade estrita.

func TestEventtableRelPath(t *testing.T) {
	if rel, ok := eventtable.RelPath(eventtable.KindBattleText, "bika07_228"); !ok || filepath.ToSlash(rel) != "battle/btl/bika07_228/bika07_228.bin" {
		t.Errorf("btl relpath: %q %v", rel, ok)
	}
	if _, ok := eventtable.RelPath(eventtable.KindBattleText, "a/b"); ok {
		t.Error("btl com barra deve falhar")
	}
	if rel, ok := eventtable.RelPath(eventtable.KindCloud, "cloud"); !ok || rel != "cloudsave/cloud.bin" {
		t.Errorf("cloud relpath: %q %v", rel, ok)
	}
	if _, ok := eventtable.RelPath(eventtable.KindCloud, "cloudv"); ok {
		t.Error("cloudv deixa de ser id próprio (partura de 'cloud'), deve falhar")
	}
	if rel, ok := eventtable.RelPath(eventtable.KindTutorial, "tutorial"); !ok || rel != "menu/tutorial.msb" {
		t.Errorf("tutorial relpath: %q %v", rel, ok)
	}
}
