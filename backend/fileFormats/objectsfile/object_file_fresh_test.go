package objectsfile_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/interactions"
	testcommon "ffxresources/testData"
)

// TestObjectFileStoreReloadOnExternalChange cobre o vigia de frescura do
// store de objetos: reuso por identidade enquanto o binário não muda e
// recarga quando o arquivo em mods/ é substituído externamente.
func TestObjectFileStoreReloadOnExternalChange(t *testing.T) {
	prevRoot, prevMods := common.GameFilesRoot, common.DisableMods
	t.Cleanup(func() {
		common.SetGameFilesRoot(prevRoot)
		common.SetModsEnabled(prevMods)
		objectsfile.ObjectFileDataStore.Clear()
	})

	layout, ok := objectsfile.FileLayoutFor(common.GameVersionFFX, "battle/kernel/command.bin")
	if !ok {
		t.Fatal("layout command.bin ausente")
	}

	srcTree := filepath.Join(testcommon.GetTestDataRootDirectory(), "FFX", "binary")
	tmp := t.TempDir()
	if err := os.CopyFS(tmp, os.DirFS(srcTree)); err != nil {
		t.Fatal(err)
	}
	common.SetGameFilesRoot(tmp)
	common.SetModsEnabled(true)
	objectsfile.ObjectFileDataStore.Clear()
	// Semeia o serviço de interação com ESTA árvore: a carga consulta o
	// singleton (gamefiles dir) e a criação com default resetaria o root.
	config := interactions.NewAppConfig()
	config.SetGameVersion(common.GameVersionFFX)
	config.SetLocation("GameFilesLocation", tmp)
	interactions.NewInteractionServiceWithConfig(config)
	if err := ffxencoding.PrepareVersionCharsets(common.CharsetVersion(common.GameVersionFFX)); err != nil {
		t.Fatalf("preparar charsets: %v", err)
	}

	f1, err := objectsfile.LoadObjectFileFromStoreByLayout(layout)
	if err != nil {
		t.Fatalf("load 1: %v", err)
	}
	f2, err := objectsfile.LoadObjectFileFromStoreByLayout(layout)
	if err != nil {
		t.Fatalf("load 2 (sem mudança): %v", err)
	}
	if f1 != f2 {
		t.Fatal("sem mudança no disco: LoadObjectFileFromStoreByLayout devia reutilizar a instância")
	}

	// Substitui o binário em mods/ externamente (mtime/size novos).
	rel := objectsfile.ObjectBinaryRelPath(layout)
	data, err := os.ReadFile(filepath.Join(tmp, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("ler original: %v", err)
	}
	modPath := filepath.Join(tmp, "mods", filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(modPath), 0o755); err != nil {
		t.Fatal(err)
	}
	altered := append([]byte{0x00}, data...)
	if err := os.WriteFile(modPath, altered, 0o644); err != nil {
		t.Fatalf("gravar mods: %v", err)
	}
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(modPath, future, future); err != nil {
		t.Fatal(err)
	}

	f3, err := objectsfile.LoadObjectFileFromStoreByLayout(layout)
	if err != nil {
		t.Fatalf("load 3 (mudou): %v", err)
	}
	if f3 == f1 {
		t.Fatal("binário mudou no disco: devia recarregar")
	}
	f4, err := objectsfile.LoadObjectFileFromStoreByLayout(layout)
	if err != nil {
		t.Fatalf("load 4: %v", err)
	}
	if f4 != f3 {
		t.Fatal("instância recarregada devia ser reutilizada")
	}
}
