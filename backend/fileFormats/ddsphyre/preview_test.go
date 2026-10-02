package ddsphyre

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDumpPNGPreview extrai pré-visualizações PNG (com o mesmo caminho que o
// frontend recebe) para conferir a orientação. Guardada por DDS_PHYRE_DUMP.
func TestDumpPNGPreview(t *testing.T) {
	out := os.Getenv("DDS_PHYRE_DUMP")
	if out == "" {
		t.Skip("defina DDS_PHYRE_DUMP para extrair previews")
	}
	root := filepath.Join("..", "..", "..", "build", "bin", "data")
	count := 0
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() || count >= 6 {
			return nil
		}
		if !strings.HasSuffix(d.Name(), Suffix) {
			return nil
		}
		// Prefere UI/texto (menu/help): orientação errada é óbvia neles.
		if !strings.Contains(strings.ToLower(path), "menu") &&
			!strings.Contains(strings.ToLower(path), "help") {
			return nil
		}
		tex, perr := ParseFile(path)
		if perr != nil {
			return nil
		}
		png, perr := tex.ToPNG()
		if perr != nil {
			t.Logf("png %s: %v", path, perr)
			return nil
		}
		dest := filepath.Join(out, "preview_"+filepath.Base(path)+".png")
		if e := os.WriteFile(dest, png, 0o644); e != nil {
			return nil
		}
		t.Logf("%s (%s %dx%d) -> %s", tex.Format, filepath.Base(path), tex.Width, tex.Height, dest)
		count++
		return nil
	})
	if count == 0 {
		t.Log("nenhuma textura de menu/help encontrada")
	}
}
