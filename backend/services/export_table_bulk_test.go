package services

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/eventtable"
	jsonfmt "ffxresources/backend/formatters/json"
)

// repoDataRootLocal sobe do arquivo de teste até o go.mod e devolve
// build/bin/data. Inline (não usa testData): o mock notifier daquele pacote
// importa services, e teste interno + testData = ciclo de import.
func repoDataRootLocal(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Skip("caminho do fonte indisponível")
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "build", "bin", "data")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("go.mod não encontrado na hierarquia")
		}
		dir = parent
	}
}

// tableBulkFixtures copia os binários 'us' mínimos (codec events) de
// build/bin/data para um GameFilesRoot temporário: o export lê de data/ e
// grava o artefato em mods/edits — tudo contido no temp, sem tocar a árvore
// real. Fixture ausente pula o teste (ambiente sem o jogo extraído).
func tableBulkFixtures(t *testing.T, version common.GameVersion, rels []string) {
	t.Helper()
	dataRoot := repoDataRootLocal(t)
	if _, err := os.Stat(dataRoot); err != nil {
		t.Skip("build/bin/data indisponível")
	}
	root := t.TempDir()
	prevRoot, prevMods := common.GameFilesRoot, common.DisableMods
	common.GameFilesRoot = root
	common.DisableMods = true
	t.Cleanup(func() {
		common.GameFilesRoot = prevRoot
		common.DisableMods = prevMods
		eventtable.ClearStore()
	})
	eventtable.ClearStore()

	locRoot := filepath.FromSlash(common.GetLocalizationRootForVersion(version, common.DefaultLocalization))
	for _, rel := range rels {
		data, err := os.ReadFile(filepath.Join(dataRoot, locRoot, filepath.FromSlash(rel)))
		if err != nil {
			t.Skipf("fixture ausente: %s (%v)", rel, err)
		}
		dst := filepath.Join(root, locRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatalf("criar diretório de %s: %v", rel, err)
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			t.Fatalf("copiar %s: %v", rel, err)
		}
	}
}

// TestTableBulkPathScopes: o nome do artefato bulk por kind segue o padrão
// <família>_all_localizations_<versão>, cria edits/ e NUNCA colide com o
// artefato por-entrada (régua de export_paths.go: seleção não pode
// sobrescrever o bulk do tradutor).
func TestTableBulkPathScopes(t *testing.T) {
	seedEventsTempRoot(t)

	cases := []struct {
		kind string
		want string
	}{
		{KindBattleText, "battletext_all_localizations_ffx.json"},
		{KindCloud, "cloudsave_all_localizations_ffx.json"},
		{KindMenuMain, "mainmenu_all_localizations_ffx.json"},
		{KindTutorial, "tuto_all_localizations_ffx.json"},
	}
	bulks := make(map[string]bool, len(cases))
	for _, tc := range cases {
		p, err := tableBulkPath(tc.kind, common.GameVersionFFX)
		if err != nil {
			t.Fatalf("%s: %v", tc.kind, err)
		}
		if base := filepath.Base(p); base != tc.want {
			t.Fatalf("%s: bulk %q, queria %q", tc.kind, base, tc.want)
		}
		if bulks[p] {
			t.Fatalf("%s: bulk colide com outro kind: %s", tc.kind, p)
		}
		bulks[p] = true
		if _, err := os.Stat(filepath.Dir(p)); err != nil {
			t.Fatalf("%s: diretório edits não criado: %v", tc.kind, err)
		}
	}

	// O artefato por-entrada é <id>_all_localizations_<v>: nenhum id servido
	// pelos kinds de tabela pode ter o mesmo nome de um bulk (cloud e
	// menumain têm entrada com o mesmo nome do kind — o bulk usa a família
	// física: cloudsave/mainmenu).
	for _, id := range []string{
		"bika00_10", "zzzz03_19", // battletext
		"cloud", "cloudv", // cloud
		"menumain", "menumain_btl", // menumain
		"tutorial", "tuto0000", // tutorial
	} {
		perEntry, err := jsonfmt.ObjectsJSONPath(id, common.GameVersionFFX)
		if err != nil {
			t.Fatalf("%s: artefato por-entrada: %v", id, err)
		}
		if bulks[perEntry] {
			t.Fatalf("bulk colide com o artefato por-entrada de %q: %s", id, perEntry)
		}
	}

	if _, err := tableBulkPath(KindEvents, common.GameVersionFFX); err == nil {
		t.Fatal("kind sem nome bulk registrado deve falhar (events tem path próprio)")
	}
}

// TestExportTableKindsBulkSingleFile: "exportar tudo" (ids vazio) dos kinds
// de tabela escreve UM artefato por kind — nunca um arquivo por entrada —
// e o arquivo volta a importar com o kind detectado pela metadata.key
// (regressão da ordem de kindFromKey: menumain_btl mora em /battle/btl/).
func TestExportTableKindsBulkSingleFile(t *testing.T) {
	tableBulkFixtures(t, common.GameVersionFFX, []string{
		"battle/btl/bika00_10/bika00_10.bin",
		"battle/btl/zzzz03_19/zzzz03_19.bin",
		"battle/btl/menumain/menumain.bin",
		"cloudsave/cloud.bin",
		"cloudsave/cloudv.bin",
		"menu/menumain.bin",
	})
	if err := ensureVersionReady(common.GameVersionFFX); err != nil {
		t.Fatalf("prepare version: %v", err)
	}
	svc := NewMetadataService(nil)

	// battletext: bulk único com as entradas da árvore (bika00_10 e
	// zzzz03_19; menumain/ é excluído do battletext pela régua do eventtable).
	paths, err := svc.ExportJSON(KindBattleText, common.GameVersionFFX, nil, nil)
	if err != nil {
		t.Fatalf("export battletext: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("exportar tudo de battletext: %d arquivo(s) (%v), queria 1", len(paths), paths)
	}
	if base := filepath.Base(paths[0]); base != "battletext_all_localizations_ffx.json" {
		t.Fatalf("bulk battletext: %q", base)
	}
	read, err := jsonfmt.NewJSONObjectFormatter().ReadObjects(paths[0])
	if err != nil {
		t.Fatalf("releitura do bulk: %v", err)
	}
	for _, id := range []string{"bika00_10", "zzzz03_19"} {
		if _, ok := read[id]; !ok {
			t.Fatalf("entrada %s ausente no bulk battletext (%v)", id, read.SortedKeys())
		}
	}
	plan, err := svc.prepareImport(paths[0], common.GameVersionFFX)
	if err != nil {
		t.Fatalf("prepare import do bulk battletext: %v", err)
	}
	if plan.kind != KindBattleText {
		t.Fatalf("import detectou kind %q, queria %q", plan.kind, KindBattleText)
	}
	if len(plan.summary.Errors) > 0 {
		t.Fatalf("bulk battletext deveria importar limpo: %v", plan.summary.Errors)
	}

	// Espelho .strings: mesmo basename, extensão trocada, também único.
	spaths, err := svc.ExportStrings(KindBattleText, common.GameVersionFFX, nil, nil)
	if err != nil {
		t.Fatalf("export .strings battletext: %v", err)
	}
	if len(spaths) != 1 || filepath.Base(spaths[0]) != "battletext_all_localizations_ffx.strings" {
		t.Fatalf("bulk .strings battletext: %v", spaths)
	}

	// Seleção mantém o artefato por-entrada (dedup intra-arquivo) e não pode
	// sobrescrever o bulk.
	sel, err := svc.ExportJSON(KindBattleText, common.GameVersionFFX, []string{"bika00_10"}, nil)
	if err != nil {
		t.Fatalf("export seleção battletext: %v", err)
	}
	if len(sel) != 1 || filepath.Base(sel[0]) != "bika00_10_all_localizations_ffx.json" {
		t.Fatalf("seleção battletext: %v", sel)
	}
	if sel[0] == paths[0] {
		t.Fatal("seleção sobrescreveu o bulk (colisão de nomes)")
	}

	// menumain: bulk único com menumain + menumain_btl; o import detecta o
	// kind certo mesmo com a key da entrada btl dentro de /battle/btl/.
	mpaths, err := svc.ExportJSON(KindMenuMain, common.GameVersionFFX, nil, nil)
	if err != nil {
		t.Fatalf("export menumain: %v", err)
	}
	if len(mpaths) != 1 || filepath.Base(mpaths[0]) != "mainmenu_all_localizations_ffx.json" {
		t.Fatalf("bulk menumain: %v", mpaths)
	}
	mread, err := jsonfmt.NewJSONObjectFormatter().ReadObjects(mpaths[0])
	if err != nil {
		t.Fatalf("releitura do bulk menumain: %v", err)
	}
	for _, id := range []string{"menumain", "menumain_btl"} {
		if _, ok := mread[id]; !ok {
			t.Fatalf("entrada %s ausente no bulk menumain (%v)", id, mread.SortedKeys())
		}
	}
	mplan, err := svc.prepareImport(mpaths[0], common.GameVersionFFX)
	if err != nil {
		t.Fatalf("prepare import do bulk menumain: %v", err)
	}
	if mplan.kind != KindMenuMain {
		t.Fatalf("import detectou kind %q, queria %q (ordem de kindFromKey?)", mplan.kind, KindMenuMain)
	}

	// cloud: bulk cloudsave_* não pode colidir com o artefato por-entrada de
	// "cloud" (mesmo nome do kind).
	cpaths, err := svc.ExportJSON(KindCloud, common.GameVersionFFX, nil, nil)
	if err != nil {
		t.Fatalf("export cloud: %v", err)
	}
	if len(cpaths) != 1 || filepath.Base(cpaths[0]) != "cloudsave_all_localizations_ffx.json" {
		t.Fatalf("bulk cloud: %v", cpaths)
	}
	perCloud, err := jsonfmt.ObjectsJSONPath("cloud", common.GameVersionFFX)
	if err != nil {
		t.Fatalf("artefato por-entrada de cloud: %v", err)
	}
	if cpaths[0] == perCloud {
		t.Fatal("bulk cloud colide com o artefato por-entrada de \"cloud\"")
	}
	cread, err := jsonfmt.NewJSONObjectFormatter().ReadObjects(cpaths[0])
	if err != nil {
		t.Fatalf("releitura do bulk cloud: %v", err)
	}
	for _, id := range []string{"cloud", "cloudv"} {
		if _, ok := cread[id]; !ok {
			t.Fatalf("entrada %s ausente no bulk cloud (%v)", id, cread.SortedKeys())
		}
	}
}
