package services

import (
	"os"
	"path/filepath"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/helpfile"
	"ffxresources/backend/formatters/hash"
)

// kindFromKey detecta help pela key (help/…/*.sps2) e mantém os outros kinds.
func TestKindFromKeyHelp(t *testing.T) {
	cases := []struct {
		key  string
		want string
	}{
		{"ffx/help/now_help/now_help.sps2", KindHelp},
		{"ffx/help/now_help/now_help_page.sps2", KindHelp},
		{"ffx/help/mon_boku/mon_boku.sps2", KindHelp},
		{"ffx/help/dvdcopy/dvdcopy_page.sps2", KindHelp},
		{"ffx/event/obj_ps3/az/azit0000/azit0000.bin", KindEvents},
		{"ffx/menu/macrodic.dcp", KindMacro},
		{"ffx/battle/btl/btl_txt.bin", KindObjects},
	}
	for _, tc := range cases {
		if got := kindFromKey(tc.key); got != tc.want {
			t.Fatalf("kindFromKey(%q) = %q, want %q", tc.key, got, tc.want)
		}
	}
}

// detectImportKind valida versão e kind: artefato help só importa na aba ffx.
func TestDetectImportKindHelpVersionGuard(t *testing.T) {
	imported := dto.Collection{
		"now_help": dto.FileEntry{
			Metadata: dto.NewHelpMetadata("now_help", "help/now_help", common.GameVersionFFX),
		},
	}
	kind, err := detectImportKind(common.GameVersionFFX, imported)
	if err != nil {
		t.Fatalf("ffx: %v", err)
	}
	if kind != KindHelp {
		t.Fatalf("kind: %q", kind)
	}
	if _, err := detectImportKind(common.GameVersionFFX2, imported); err == nil {
		t.Fatal("ffx2 deveria rejeitar artefato help de ffx")
	}
}

// seedFFXHelpRoot copia os fixtures FFX para um root temporário e aponta o
// GameFilesRoot para lá (espelho do seedTempHelpRoot dos builders).
func seedFFXHelpRoot(t *testing.T) {
	t.Helper()
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "..", "testData", "FFX", "binary"))
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	tempRoot := t.TempDir()
	if err := os.CopyFS(tempRoot, os.DirFS(fixtureRoot)); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	prevRoot, prevMods := common.GameFilesRoot, common.DisableMods
	common.GameFilesRoot = tempRoot
	common.DisableMods = true
	t.Cleanup(func() {
		common.GameFilesRoot = prevRoot
		common.DisableMods = prevMods
		helpfile.ClearHelp(common.GameVersionFFX)
	})
}

// helpRowIsRef espelha a regra do backend: ref = "$" + hash PRÓPRIO da row.
func helpRowIsRef(r dto.TextRow) bool {
	bare, ok := hash.Strip(r.Text[common.DefaultLocalization])
	return ok && bare == r.Hash[common.DefaultLocalization]
}

// GetEntry de help entrega o DTO DEDUPLICADO (refs p/ ocultar na tabela);
// GetCollection continua RAW (export/preview/validação de import).
func TestGetEntryHelpDedupsDTO(t *testing.T) {
	seedFFXHelpRoot(t)
	svc := NewMetadataService(nil)

	raw, err := svc.GetCollection(KindHelp, common.GameVersionFFX, []string{"now_help_page"})
	if err != nil {
		t.Fatalf("collection: %v", err)
	}
	entryRaw, ok := raw["now_help_page"]
	if !ok {
		t.Fatalf("now_help_page ausente: %v", raw.SortedKeys())
	}
	for _, r := range entryRaw.Rows {
		if helpRowIsRef(r) {
			t.Fatalf("collection raw com ref em [%d]", r.Index)
		}
	}

	// now_help_page repete textos com now_help (defs em now_help, que vem
	// antes na ordem de chave) → as cópias viram refs no DTO entregue.
	entry, err := svc.GetEntry(KindHelp, "now_help_page", common.GameVersionFFX)
	if err != nil {
		t.Fatalf("get entry: %v", err)
	}
	refs := 0
	for _, r := range entry.Rows {
		if helpRowIsRef(r) {
			refs++
		}
	}
	if refs == 0 {
		t.Fatal("GetEntry deveria entregar refs de dedup em now_help_page")
	}

	// Toda ref tem a def literal correspondente em algum dos 6 painéis.
	full, err := svc.GetCollection(KindHelp, common.GameVersionFFX, nil)
	if err != nil {
		t.Fatalf("full: %v", err)
	}
	defs := make(map[string]bool)
	for _, k := range full.SortedKeys() {
		for _, r := range full[k].Rows {
			if h := r.Hash[common.DefaultLocalization]; h != "" && r.Text[common.DefaultLocalization] != "" {
				defs[h] = true
			}
		}
	}
	for _, r := range entry.Rows {
		if !helpRowIsRef(r) {
			continue
		}
		bare, _ := hash.Strip(r.Text[common.DefaultLocalization])
		if !defs[bare] {
			t.Fatalf("ref %s sem def literal na collection completa", bare)
		}
	}
}

// Compat: artefatos help antigos carregavam name (text_%04d); sem a
// normalização, o rowKey (index\x00name) não casa com a store nova e o
// import seria bloqueado.
func TestStripHelpRowNamesImportCompat(t *testing.T) {
	text := map[string]string{common.DefaultLocalization: "Directional Button"}
	h := hash.Texts(text)
	metadata := func() dto.Metadata {
		return dto.NewHelpMetadata("now_help", "help/now_help", common.GameVersionFFX)
	}
	store := dto.Collection{
		"now_help": dto.FileEntry{
			Metadata: metadata(),
			Rows:     []dto.TextRow{{Index: 3, Hash: h, Text: text}},
		},
	}
	legacy := dto.Collection{
		"now_help": dto.FileEntry{
			Metadata: metadata(),
			Rows:     []dto.TextRow{{Index: 3, Name: "text_0003", Hash: h, Text: text}},
		},
	}

	if _, errs := validateImportEntry(legacy["now_help"], store["now_help"]); len(errs) == 0 {
		t.Fatal("artefato legado sem normalização deveria falhar no rowKey")
	}

	stripHelpRowNames(legacy)
	if legacy["now_help"].Rows[0].Name != "" {
		t.Fatalf("name legado não removido: %q", legacy["now_help"].Rows[0].Name)
	}
	changed, errs := validateImportEntry(legacy["now_help"], store["now_help"])
	if len(errs) > 0 {
		t.Fatalf("após normalizar: %v", errs)
	}
	if changed != 0 {
		t.Fatalf("changed = %d, queria 0 (mesmo texto)", changed)
	}
}
