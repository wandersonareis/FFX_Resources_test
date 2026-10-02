package ddsphyre

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExtractSamplesForValidation extrai amostras (uma por formato
// distinto) para validar o header gerado com o nvddsinfo.exe do repo.
func TestExtractSamplesForValidation(t *testing.T) {
	if os.Getenv("DDS_PHYRE_DUMP") == "" {
		t.Skip("defina DDS_PHYRE_DUMP para extrair amostras")
	}
	out := os.Getenv("DDS_PHYRE_DUMP")
	roots := []string{
		filepath.Join("..", "..", "..", "build", "bin", "data"),
	}
	seen := map[string]bool{}
	count := 0
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil || d == nil || d.IsDir() || count >= 8 {
				return nil
			}
			if !strings.HasSuffix(d.Name(), Suffix) {
				return nil
			}
			tex, perr := ParseFile(path)
			if perr != nil {
				t.Logf("parse %s: %v", path, perr)
				return nil
			}
			if seen[tex.Format] {
				return nil
			}
			seen[tex.Format] = true
			dds, e := tex.ExtractToDDS()
			if e != nil {
				t.Logf("extract %s: %v", path, e)
				return nil
			}
			dest := filepath.Join(out, filepath.Base(path)+".dds")
			if e := os.WriteFile(dest, dds, 0o644); e != nil {
				t.Logf("gravando %s: %v", dest, e)
				return nil
			}
			t.Logf("%s -> %s (%s %dx%d mips=%d)", tex.Format, dest, tex.Format, tex.Width, tex.Height, tex.MipmapCount)
			count++
			return nil
		})
	}
}
