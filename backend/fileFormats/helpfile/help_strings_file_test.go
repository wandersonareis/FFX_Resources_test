package helpfile

import (
	"os"
	"path/filepath"
	"testing"

	"ffxresources/backend/common"
)

// Absolutiza a raiz de gamefiles dos fixtures e devolve restaurador.
func useTestDataRoot(t *testing.T) {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", "..", "testData", "FFX", "binary"))
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	prevRoot, prevMods := common.GameFilesRoot, common.DisableMods
	common.GameFilesRoot = abs
	t.Cleanup(func() {
		common.GameFilesRoot = prevRoot
		common.DisableMods = prevMods
	})
}

func useFFX2Root(t *testing.T) {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", "..", "testData", "FFX-2", "binary"))
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	prevRoot, prevMods := common.GameFilesRoot, common.DisableMods
	common.GameFilesRoot = abs
	t.Cleanup(func() {
		common.GameFilesRoot = prevRoot
		common.DisableMods = prevMods
	})
}

// LoadFromBinary carrega os 6 painéis (FFX) fundindo as localizações que têm
// o arquivo (us e jp nas fixtures).
func TestLoadFromBinaryLoadsAllPanels(t *testing.T) {
	useTestDataRoot(t)
	binFile := NewHelpStringsBinaryFile(common.GameVersionFFX)
	if err := binFile.LoadFromBinary(); err != nil {
		t.Fatalf("load: %v", err)
	}
	if binFile.GetObjects().Len() != len(HelpEntries) {
		t.Fatalf("painéis: got %d want %d", binFile.GetObjects().Len(), len(HelpEntries))
	}

	// Ordem determinística = registro.
	for i, entry := range HelpEntries {
		obj, ok := binFile.GetObjects().Get(i).(*HelpKeyedStringFile)
		if !ok || obj.Name != entry.Name {
			t.Fatalf("objeto %d: esperado %s", i, entry.Name)
		}
		if obj.Files[common.DefaultLocalization] == nil {
			t.Fatalf("painel %s sem conteúdo 'us'", entry.Name)
		}
	}

	// Segmentos fundidos: now_help (1208) e mon_boku (1578) com conteúdo us+jp.
	nowHelp := binFile.GetObjects().Get(0).(*HelpKeyedStringFile)
	if nowHelp.SegmentCount() != 1208 {
		t.Fatalf("now_help segmentos: %d", nowHelp.SegmentCount())
	}
	monBoku := binFile.GetObjects().Get(2).(*HelpKeyedStringFile)
	if monBoku.SegmentCount() != 1578 {
		t.Fatalf("mon_boku segmentos: %d", monBoku.SegmentCount())
	}
	strs := monBoku.LocalizedStrings()
	if len(strs) != 1578 {
		t.Fatalf("localized strings: %d", len(strs))
	}
	if got := strs[0].GetLocalizedString(common.DefaultLocalization); got == "" {
		// seg 0 do mon_boku é espaço + comandos; o texto tem ao menos um espaço.
		t.Logf("seg 0 us = %q", got)
	}
}

// ffx2/lastmiss não têm a pasta help/: carga vazia, sem erro.
func TestLoadFromBinaryFFX2Empty(t *testing.T) {
	useFFX2Root(t)
	binFile := NewHelpStringsBinaryFile(common.GameVersionFFX2)
	if err := binFile.LoadFromBinary(); err != nil {
		t.Fatalf("load ffx2: %v", err)
	}
	if binFile.GetObjects().Len() != 0 {
		t.Fatalf("ffx2 deveria carregar 0 painéis, carregou %d", binFile.GetObjects().Len())
	}
}

// ToBytes reconstrói o binário da localização (byte-exato sem edição).
func TestToBytesRoundTrip(t *testing.T) {
	useTestDataRoot(t)
	binFile := NewHelpStringsBinaryFile(common.GameVersionFFX)
	if err := binFile.LoadFromBinary(); err != nil {
		t.Fatalf("load: %v", err)
	}
	obj := binFile.GetObjects().Get(0).(*HelpKeyedStringFile)
	rebuilt, err := obj.ToBytes(common.DefaultLocalization)
	if err != nil {
		t.Fatalf("ToBytes: %v", err)
	}
	original, err := os.ReadFile(testNowHelpPath)
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	if string(rebuilt) != string(original) {
		t.Fatal("ToBytes divergiu do original")
	}
}

// EnsureHelpLoaded popula o store da versão; GetHelp devolve os painéis.
func TestEnsureHelpLoadedStore(t *testing.T) {
	useTestDataRoot(t)
	if err := EnsureHelpLoaded(common.GameVersionFFX); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	for _, name := range HelpEntryNames() {
		if GetHelp(common.GameVersionFFX, name) == nil {
			t.Fatalf("painel %s ausente no store", name)
		}
	}
	if !HasHelp(common.GameVersionFFX) {
		t.Fatal("HasHelp deveria ser true para ffx")
	}
	if GetHelp(common.GameVersionFFX2, NowHelpName) != nil {
		t.Fatal("ffx2 não deveria ter painéis")
	}
}
