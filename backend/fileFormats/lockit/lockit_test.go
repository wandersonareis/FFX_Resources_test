package lockit_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/lockit"
	jsonfmt "ffxresources/backend/formatters/json"
	testcommon "ffxresources/testData"
)

func projectRoot() string {
	return filepath.Dir(testcommon.GetTestDataRootDirectory())
}

func srcDataRoot() string {
	return filepath.Join(projectRoot(), "build", "bin", "data")
}

// setupTempData copia as árvores de lockit para um diretório temporário e
// aponta o GameFilesRoot para lá (sem tocar nos originais).
func setupTempData(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	for _, rel := range []string{
		"ffx_data/gamedata/ps3data/lockit",
		"ffx-2_data/gamedata/ps3data/lockit",
	} {
		copyTree(t, filepath.Join(srcDataRoot(), filepath.FromSlash(rel)), filepath.Join(tmp, filepath.FromSlash(rel)))
	}
	prev := common.GameFilesRoot
	common.SetGameFilesRoot(tmp)
	common.SetVerboseMode(false)
	lockit.DataStore.Clear()
	t.Cleanup(func() {
		common.SetGameFilesRoot(prev)
		lockit.DataStore.Clear()
	})
	return tmp
}

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("read dir %s: %v", src, err)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dst, err)
	}
	for _, e := range entries {
		s := filepath.Join(src, e.Name())
		d := filepath.Join(dst, e.Name())
		if e.IsDir() {
			copyTree(t, s, d)
			continue
		}
		data, err := os.ReadFile(s)
		if err != nil {
			t.Fatalf("read %s: %v", s, err)
		}
		if err := os.WriteFile(d, data, 0o644); err != nil {
			t.Fatalf("write %s: %v", d, err)
		}
	}
}

func runCount(f *lockit.LockitFile) int {
	runs := 0
	var prev lockit.Kind
	for i, r := range f.Records() {
		if i == 0 || r.Kind() != prev {
			runs++
		}
		prev = r.Kind()
	}
	return runs
}

func TestLockitLayoutRegistry(t *testing.T) {
	if _, ok := lockit.LayoutForID(common.GameVersionFFX, "ffx_loc_kit_ps3"); !ok {
		t.Fatal("expected ffx lockit layout")
	}
	if _, ok := lockit.LayoutForID(common.GameVersionFFX2, "ffx2_loc_kit_ps3"); !ok {
		t.Fatal("expected ffx2 lockit layout")
	}
	if lockit.IsLockitKey("ffx2/gamedata/ps3data/lockit/ffx2_loc_kit_ps3.bin") != true {
		t.Fatal("expected lockit key detection")
	}
	if lockit.IsLockitKey("ffx/battle/kernel/command.bin") != false {
		t.Fatal("did not expect lockit key")
	}
}

func TestLockitRoundTripBytes(t *testing.T) {
	tmp := setupTempData(t)

	cases := []struct {
		version common.GameVersion
		stem    string
		game    int
		utf8    int
		runs    int
	}{
		{common.GameVersionFFX, "ffx_loc_kit_ps3", 526, 388, 12},
		{common.GameVersionFFX2, "ffx2_loc_kit_ps3", 1220, 476, 16},
	}

	for _, tc := range cases {
		l, ok := lockit.LayoutForID(tc.version, tc.stem)
		if !ok {
			t.Fatalf("layout %s not found", tc.stem)
		}
		f, err := lockit.Load(l)
		if err != nil {
			t.Fatalf("load %s: %v", tc.stem, err)
		}
		game, utf8 := f.IndexesByKind()
		if len(game) != tc.game || len(utf8) != tc.utf8 {
			t.Fatalf("%s: game=%d utf8=%d, want game=%d utf8=%d", tc.stem, len(game), len(utf8), tc.game, tc.utf8)
		}
		if got := runCount(f); got != tc.runs {
			t.Fatalf("%s: runs=%d, want %d", tc.stem, got, tc.runs)
		}

		encoded, err := f.Bytes()
		if err != nil {
			t.Fatalf("encode %s: %v", tc.stem, err)
		}
		for _, lang := range f.Languages() {
			orig, err := os.ReadFile(filepath.Join(tmp, filepath.FromSlash(l.RelPath(lang))))
			if err != nil {
				t.Fatalf("read orig %s/%s: %v", tc.stem, lang, err)
			}
			if !bytes.Equal(encoded[lang], orig) {
				t.Fatalf("%s/%s: round-trip diverge (orig=%d novos=%d)", tc.stem, lang, len(orig), len(encoded[lang]))
			}
		}
	}
}

func TestLockitDTOExportGroupingAndImportUsOnly(t *testing.T) {
	tmp := setupTempData(t)

	l, _ := lockit.LayoutForID(common.GameVersionFFX, "ffx_loc_kit_ps3")

	c, err := builders.BuildLockitDTO(common.GameVersionFFX, nil)
	if err != nil {
		t.Fatalf("build dto: %v", err)
	}
	entry, ok := c["ffx_loc_kit_ps3"]
	if !ok {
		t.Fatalf("entry missing: %v", c.SortedKeys())
	}

	// game ocupa 0..G-1 e utf8 G..G+U-1, nessa ordem após SortRows.
	game, utf8 := 0, 0
	for i, row := range entry.Rows {
		switch row.Name {
		case "game":
			if row.Index != game {
				t.Fatalf("row %d: game index=%d, want %d", i, row.Index, game)
			}
			game++
		case "utf8":
			if row.Index != 526+utf8 {
				t.Fatalf("row %d: utf8 index=%d, want %d", i, row.Index, 526+utf8)
			}
			utf8++
		default:
			t.Fatalf("row %d: name inválido %q", i, row.Name)
		}
		if len(row.Text) < 2 {
			t.Fatalf("row %d: esperado múltiplos idiomas de referência, got %v", i, row.Text)
		}
	}
	if game != 526 || utf8 != 388 {
		t.Fatalf("grouping game=%d utf8=%d", game, utf8)
	}

	// Altera o us de uma row com texto conhecido e aplica.
	target := entry.Rows[2] // deve ser game index 2 (FINAL FANTASY X HD Remaster)
	if target.Name != "game" || target.Index != 2 {
		t.Fatalf("row[2] inesperada: name=%q index=%d", target.Name, target.Index)
	}
	entry.Rows[2].Text["us"] = "LOCKIT EDITADO"
	if err := builders.ApplyLockitDTO(common.GameVersionFFX, dto.Collection{"ffx_loc_kit_ps3": entry}); err != nil {
		t.Fatalf("apply: %v", err)
	}

	// Reload: us vem do mods (editado); de continua o original.
	lockit.DataStore.Clear()
	f2, err := lockit.Load(l)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	game2, _ := f2.IndexesByKind()
	rec := f2.Records()[game2[2]]
	if got := rec.Text("us"); got != "LOCKIT EDITADO" {
		t.Fatalf("us não aplicado: %q", got)
	}
	if rec.Text("de") == "LOCKIT EDITADO" {
		t.Fatalf("de não deveria ser alterado")
	}

	// O arquivo us foi gravado em mods; de permanece intacto no original.
	modUs := filepath.Join(tmp, "mods", filepath.FromSlash(l.RelPath("us")))
	if _, err := os.Stat(modUs); err != nil {
		t.Fatalf("us mod não gravado: %v", err)
	}
	modDe := filepath.Join(tmp, "mods", filepath.FromSlash(l.RelPath("de")))
	if _, err := os.Stat(modDe); err == nil {
		t.Fatalf("de não deveria ter sido gravado no import (us-only)")
	}
}

func TestLockitJSONExportShape(t *testing.T) {
	setupTempData(t)

	c, err := builders.BuildLockitDTO(common.GameVersionFFX, nil)
	if err != nil {
		t.Fatalf("build dto: %v", err)
	}
	raw, err := jsonfmt.NewJSONObjectFormatter().Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(raw)

	// Separação por tipo com o Name do objectsfile.
	if !strings.Contains(body, `"name": "game"`) || !strings.Contains(body, `"name": "utf8"`) {
		t.Fatalf("json não separa game/utf8 por name")
	}
	// Todos os idiomas presentes como referência de tradução.
	for _, lang := range []string{"us", "de", "fr", "it", "sp", "jp", "kr", "ch"} {
		if !strings.Contains(body, `"`+lang+`":`) {
			t.Fatalf("idioma %s ausente no json exportado", lang)
		}
	}
	// Dedup: o texto repetido entre game e utf8 vira referência $hash.
	if !strings.Contains(body, `"$`) {
		t.Fatalf("esperado referência de dedup ($hash) no json")
	}
	// game antes de utf8 (índice crescente).
	if strings.Index(body, `"name": "game"`) > strings.Index(body, `"name": "utf8"`) {
		t.Fatalf("game deveria vir antes de utf8")
	}
}
