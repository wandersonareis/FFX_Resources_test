package macrodic_test

import (
	"errors"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/macrodic"
)

// snapshotCharMaps guarda os mapas atuais para restaurar após o teste
// (ClearAllCharMaps seria destrutivo para as demais specs do pacote).
func snapshotCharMaps() map[common.GameVersion]map[string][2]any {
	snap := make(map[common.GameVersion]map[string][2]any)
	for _, v := range []common.GameVersion{common.GameVersionFFX, common.GameVersionFFX2, common.GameVersionLastMiss} {
		snap[v] = make(map[string][2]any)
		for loc := range common.SupportedLanguages {
			charset := ffxencoding.GetCharsetForLanguage(loc)
			snap[v][charset] = [2]any{
				ffxencoding.GetByteToCharMap(v, charset),
				ffxencoding.GetCharToByteMap(v, charset),
			}
		}
	}
	return snap
}

func restoreCharMaps(snap map[common.GameVersion]map[string][2]any) {
	for v, byCharset := range snap {
		for charset, maps := range byCharset {
			b2c, _ := maps[0].(map[uint]rune)
			c2b, _ := maps[1].(map[rune]uint)
			if b2c == nil && c2b == nil {
				continue
			}
			ffxencoding.SetCharMap(v, charset, b2c, c2b)
		}
	}
}

func TestImportFromJsonAbortsWithoutCharsetMaps(t *testing.T) {
	snap := snapshotCharMaps()
	defer restoreCharMaps(snap)
	ffxencoding.ClearAllCharMaps()

	data := &macrodic.MacroDictionaryJsonImport{
		Chunks: []macrodic.MacroChunkJsonImport{
			{ChunkIndex: 0, Strings: []macrodic.MacroStringJsonImport{
				{Index: 0, Name: map[string]string{"us": "A"}},
			}},
		},
	}
	if _, err := macrodic.ImportFromJson(data, common.GameVersionFFX); !errors.Is(err, ffxencoding.ErrVersionNotFound) {
		t.Fatalf("missing maps must abort with ErrVersionNotFound: %v", err)
	}
}
