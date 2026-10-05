package services_test

import (
	"path/filepath"
	"strings"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/services"
	testcommon "ffxresources/testData"
)

// TestTagCatalogRoundTrip garante que cada tag sugerida pelo catálogo
// sobrevive ao round-trip texto → bytes → texto idêntico — o contrato que o
// editor pressupõe ao inserir sugestões como chip atômico.
func TestTagCatalogRoundTrip(t *testing.T) {
	prevRoot := common.GameFilesRoot
	t.Cleanup(func() { common.GameFilesRoot = prevRoot })

	common.SetGameFilesRoot(filepath.Join(testcommon.GetTestDataRootDirectory(), "FFX", "binary"))

	// ensureVersionReady memoiza por processo; prepara direto para garantir os
	// macros do testData mesmo que outro teste já tenha marcado a versão.
	if err := reader.PrepareVersion(common.GameVersionFFX); err != nil {
		t.Fatalf("PrepareVersion: %v", err)
	}

	catalog, err := services.BuildTagCatalog(common.GameVersionFFX)
	if err != nil {
		t.Fatalf("BuildTagCatalog: %v", err)
	}
	if catalog.Version != "ffx" {
		t.Fatalf("versão do catálogo = %q, quero %q", catalog.Version, "ffx")
	}

	// As 4 tags nomeadas do estágio 1, sempre presentes e não vazias.
	byTag := map[string]services.TagCatalogEntry{}
	for _, entry := range catalog.Tags {
		if want := "{" + entry.Tag + ":"; !strings.HasPrefix(entry.Prefix, want) {
			t.Errorf("prefixo de %s = %q, quero que comece com %q", entry.Tag, entry.Prefix, want)
		}
		if len(entry.Values) == 0 {
			t.Errorf("entrada %s sem valores", entry.Tag)
		}
		byTag[entry.Tag] = entry
	}
	for _, name := range []string{"PC", "MCR", "BUTTON", "ICON"} {
		if _, ok := byTag[name]; !ok {
			t.Errorf("catálogo sem a entrada %s", name)
		}
	}

	// Valores canônicos esperados (ordem fixa: índice crescente quando não há
	// ranking de frequência no diretório de teste).
	pc := byTag["PC"].Values
	if len(pc) == 0 || pc[0].Tag != "{PC:00:TIDUS}" {
		t.Errorf("primeiro valor de PC = %+v, quero {PC:00:TIDUS}", pc)
	}
	if byTag["MCR"].Values[0].Key == "s00:l00" {
		t.Error("MCR não pode sugerir a sessão 0 (byte 0x13 é PC)")
	}

	// Round-trip byte-idêntico de tudo que o catálogo sugere.
	for _, entry := range catalog.Tags {
		for _, v := range entry.Values {
			// Regressão: a tag tem que fechar no "}" — uma aspa extra depois do
			// fecha-chave sobrevivia ao round-trip (virava texto literal) e o
			// chip formado no editor acabava com o label contaminado.
			if !strings.HasSuffix(v.Tag, "}") {
				t.Errorf("%s: tag não fecha no }: %s", entry.Tag, v.Tag)
			}
			if idx := strings.LastIndex(v.Tag, "}"); idx+1 < len(v.Tag) {
				t.Errorf("%s: sobra %q depois do fecha-chave em %s", entry.Tag, v.Tag[idx+1:], v.Tag)
			}
			if strings.Count(v.Tag, "{") != 1 {
				t.Errorf("%s: tag fora do formato {…}: %s", entry.Tag, v.Tag)
			}
			if entry.Tag == "MCR" && strings.Contains(v.Tag, "<Missing>") &&
				!strings.HasSuffix(v.Tag, ":<Missing>}") {
				t.Errorf("MCR sem macro deve emitir <Missing> sem aspas: %s", v.Tag)
			}
			b, convErr := converter.StringToBytes(v.Tag, "us", common.GameVersionFFX)
			if convErr != nil {
				t.Errorf("%s: StringToBytes: %v", v.Tag, convErr)
				continue
			}
			if back := converter.BytesToString(b, "us", common.GameVersionFFX); back != v.Tag {
				t.Errorf("round-trip quebrou:\n  entrou: %s\n  voltou: %s", v.Tag, back)
			}
		}
	}
}

// TestTagCatalogMCRCoverage garante que o MCR cobre TODAS as seções usadas
// pelo jogo: o chunk 7 (personagens de blitzball) e o chunk 13 (cidades/
// mapas — Zanarkand, Spira, Luca) fazem parte do mesmo espaço s01–s0F do
// decoder. Regressão do sintoma de "só aparecem NPCs": sem filtro, a ordem
// por índice lista primeiro o chunk 6 (NPCs), mas s07/s0D têm de estar
// presentes, e nenhum valor pode carregar hash não resolvido.
func TestTagCatalogMCRCoverage(t *testing.T) {
	prevRoot := common.GameFilesRoot
	t.Cleanup(func() { common.GameFilesRoot = prevRoot })

	common.SetGameFilesRoot(filepath.Join(testcommon.GetTestDataRootDirectory(), "FFX", "binary"))
	if err := reader.PrepareVersion(common.GameVersionFFX); err != nil {
		t.Fatalf("PrepareVersion: %v", err)
	}

	catalog, err := services.BuildTagCatalog(common.GameVersionFFX)
	if err != nil {
		t.Fatalf("BuildTagCatalog: %v", err)
	}

	var mcr services.TagCatalogEntry
	found := false
	for _, entry := range catalog.Tags {
		if entry.Tag == "MCR" {
			mcr, found = entry, true
			break
		}
	}
	if !found {
		t.Fatal("catálogo sem a entrada MCR")
	}

	byKey := map[string]string{} // key -> label
	for _, v := range mcr.Values {
		byKey[v.Key] = v.Label
		if strings.Contains(v.Tag, "$") {
			t.Errorf("MCR sugere hash não resolvido: %s", v.Tag)
		}
		if strings.HasPrefix(v.Key, "s10") || strings.HasPrefix(v.Key, "s11") {
			t.Errorf("MCR sugere seção inalcançável pelo decoder: %s", v.Tag)
		}
	}

	// s06 = NPC, s07 = personagens de blitzball, s0D = locais/cidades.
	for key, want := range map[string]string{
		"s06:l01": "Gatta",
		"s07:l02": "Datto",
		"s0D:l01": "Zanarkand",
		"s0D:l03": "Spira",
	} {
		if label, ok := byKey[key]; !ok {
			t.Errorf("MCR sem a chave %s (valores no catálogo: %d)", key, len(mcr.Values))
		} else if label != want {
			t.Errorf("MCR %s = %q, quero %q", key, label, want)
		}
	}
}
