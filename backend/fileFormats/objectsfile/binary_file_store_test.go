package objectsfile

import (
	"os"
	"path/filepath"
	"testing"

	"ffxresources/backend/common"
)

// TestReadFileFromBaseModsFirst valida a preferência mods-first do store de
// objetos: com mods habilitado, um binário presente em <base>/mods/<rel>
// vence o de <base>/<rel>; ausente (ou mods desabilitado) cai para o
// original.
func TestReadFileFromBaseModsFirst(t *testing.T) {
	prevMods := common.DisableMods
	defer func() { common.DisableMods = prevMods }()

	store := NewObjectBinaryFileStore("battle/kernel/test.bin", nil, "us", common.GameVersionFFX2)
	rel := store.resolveFilePath()

	base := t.TempDir()
	dataPath := filepath.Join(base, filepath.FromSlash(rel))
	modPath := filepath.Join(base, "mods", filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(dataPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(modPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dataPath, []byte("DATA"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(modPath, []byte("MODS"), 0644); err != nil {
		t.Fatal(err)
	}

	common.SetModsEnabled(true)
	got, err := store.readFileFromBase(base)
	if err != nil {
		t.Fatalf("readFileFromBase com mods: %v", err)
	}
	if string(got) != "MODS" {
		t.Fatalf("mods habilitado deve ler mods/<rel>: got %q", got)
	}

	// Ausente em mods → fallback para o original.
	if err := os.Remove(modPath); err != nil {
		t.Fatal(err)
	}
	got, err = store.readFileFromBase(base)
	if err != nil {
		t.Fatalf("readFileFromBase fallback: %v", err)
	}
	if string(got) != "DATA" {
		t.Fatalf("fallback deve ler <base>/<rel>: got %q", got)
	}

	// Mods desabilitado → sempre o original, mesmo com mods/<rel> presente.
	if err := os.WriteFile(modPath, []byte("MODS"), 0644); err != nil {
		t.Fatal(err)
	}
	common.SetModsEnabled(false)
	got, err = store.readFileFromBase(base)
	if err != nil {
		t.Fatalf("readFileFromBase com mods off: %v", err)
	}
	if string(got) != "DATA" {
		t.Fatalf("mods desabilitado deve ler <base>/<rel>: got %q", got)
	}
}
