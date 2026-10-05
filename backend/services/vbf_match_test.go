package services

import (
	"strings"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/ddsphyre"
	"ffxresources/backend/fileFormats/lockit"
)

// Caminhos reais tirados do índice de FFX_Data.vbf e FFX2_Data.vbf — se o
// formato ou o matcher mudarem, estes casos param de bater antes de o
// navegador mostrar a árvore errada.

func TestMatchVbfPath_CaminhosReais(t *testing.T) {
	cases := []struct {
		name    string
		vbf     string
		inner   string
		wantOk  bool
		kind    string
		id      string
		version common.GameVersion
	}{
		{
			name: "evento em new_uspc", vbf: "FFX_Data.vbf",
			inner:  "ffx_ps2/ffx/master/new_uspc/event/obj_ps3/zn/znkd1400/znkd1400.bin",
			wantOk: true, kind: KindEvents, id: "znkd1400", version: common.GameVersionFFX,
		},
		{
			name: "mesmo evento em outra localização (new_sppc)", vbf: "FFX_Data.vbf",
			inner:  "ffx_ps2/ffx/master/new_sppc/event/obj_ps3/zn/znkd1400/znkd1400.bin",
			wantOk: true, kind: KindEvents, id: "znkd1400", version: common.GameVersionFFX,
		},
		{
			name: "lockit por idioma", vbf: "FFX_Data.vbf",
			inner:  "ffx_data/gamedata/ps3data/lockit/ffx_loc_kit_ps3_ch.bin",
			wantOk: true, kind: KindLockit, id: "ffx_loc_kit_ps3", version: common.GameVersionFFX,
		},
		{
			name: "macrodic da localização padrão", vbf: "FFX_Data.vbf",
			inner:  "ffx_ps2/ffx/master/new_uspc/menu/macrodic.dcp",
			wantOk: true, kind: KindMacro, id: "", version: common.GameVersionFFX,
		},
		{
			name: "macrodic de outra localização", vbf: "FFX_Data.vbf",
			inner:  "ffx_ps2/ffx/master/new_sppc/menu/macrodic.dcp",
			wantOk: true, kind: KindMacro, id: "", version: common.GameVersionFFX,
		},
		{
			name: "macrodic legado sem o prefixo new_", vbf: "FFX_Data.vbf",
			inner:  "ffx_ps2/ffx/master/uspc/menu/macrodic.dcp",
			wantOk: false,
		},
		{
			name: "painel de help", vbf: "FFX_Data.vbf",
			inner:  "ffx_ps2/ffx/master/new_uspc/help/s_monitor/s_monitor.sps2",
			wantOk: true, kind: KindHelp, id: "s_monitor", version: common.GameVersionFFX,
		},
		{
			name: "subpágina de help não é entrada", vbf: "FFX_Data.vbf",
			inner:  "ffx_ps2/ffx/master/new_uspc/help/s_monitor/gxm/s_monitor_page.sps2",
			wantOk: false,
		},
		{
			name: "binário de objects", vbf: "FFX_Data.vbf",
			inner:  "ffx_ps2/ffx/master/new_uspc/battle/kernel/monmagic1.bin",
			wantOk: true, kind: KindObjects, id: "monmagic1", version: common.GameVersionFFX,
		},
		{
			name: "textura", vbf: "FFX_Data.vbf",
			inner: "ffx_data/gamedata/ps3data/yonishi_data/dat_et/et_ffx/tex/d3d11/" +
				"15040_19_0_0_128_128.dds.phyre",
			wantOk: true, kind: KindImages, version: common.GameVersionFFX,
		},
		{
			name: "fora do escopo do app", vbf: "FFX_Data.vbf",
			inner:  "version_config/region.cfg",
			wantOk: false,
		},
		{
			name: "árvore do ffx2 num .vbf do ffx", vbf: "FFX_Data.vbf",
			inner:  "ffx-2_data/gamedata/ps3data/lockit/ffx2_loc_kit_ps3_ch.bin",
			wantOk: true, kind: KindLockit, id: "ffx2_loc_kit_ps3", version: common.GameVersionFFX2,
		},
		{
			name: "evento do ffx2", vbf: "FFX2_Data.vbf",
			inner:  "ffx_ps2/ffx2/master/new_uspc/event/obj_ps3/zn/znkd1500/znkd1500.bin",
			wantOk: true, kind: KindEvents, id: "znkd1500", version: common.GameVersionFFX2,
		},
		{
			name: "textura do ffx2", vbf: "FFX2_Data.vbf",
			inner: "ffx-2_data/gamedata/ps3data/yonishi_data/dat_et/et_ffx/tex/d3d11/" +
				"15040_19_0_0_128_128.dds.phyre",
			wantOk: true, kind: KindImages, version: common.GameVersionFFX2,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := matchVbfPath(c.vbf, c.inner)
			if ok != c.wantOk {
				t.Fatalf("ok=%v (queria %v), alvo=%+v", ok, c.wantOk, got)
			}
			if !c.wantOk {
				return
			}
			if got.Kind != c.kind {
				t.Errorf("kind = %q, queria %q", got.Kind, c.kind)
			}
			if c.id != "" && got.ID != c.id {
				t.Errorf("id = %q, queria %q", got.ID, c.id)
			}
			if got.Version != c.version {
				t.Errorf("version = %s, queria %s", got.Version, c.version)
			}
		})
	}
}

// A textura clicada tem de virar um id que o próprio ddsphyre aceita — é o
// mesmo id usado em data/ e em mods/.
func TestMatchVbfPath_TexturaViraIDValido(t *testing.T) {
	inner := "ffx_data/gamedata/ps3data/yonishi_data/dat_et/et_ffx/tex/d3d11/" +
		"15040_19_0_0_128_128.dds.phyre"
	got, ok := matchVbfPath("FFX_Data.vbf", inner)
	if !ok {
		t.Fatal("textura não casou")
	}
	if !ddsphyre.ValidID(got.ID) {
		t.Fatalf("id devolvido inválido: %q", got.ID)
	}
	rel, okRel := originalRelPath(KindImages, got.ID, got.Version)
	if !okRel {
		t.Fatal("originalRelPath não resolveu o id da textura")
	}
	if !eqPath(rel, inner) {
		t.Fatalf("round-trip divergiu: %q vs %q", rel, inner)
	}
}

func TestVbfVersionOf(t *testing.T) {
	cases := []struct {
		name, inner string
		want        common.GameVersion
	}{
		{"FFX_Data.vbf", "", common.GameVersionFFX},
		{"FFX2_Data.vbf", "", common.GameVersionFFX2},
		{"ffx_data.vbf", "ffx_ps2/ffx/master/new_uspc/menu/macrodic.dcp", common.GameVersionFFX},
		{"qualquer.vbf", "ffx_ps2/ffx2/master/new_uspc/menu/macrodic.dcp", common.GameVersionFFX2},
		{"qualquer.vbf", "ffx-2_data/gamedata/x.dds.phyre", common.GameVersionFFX2},
	}
	for _, c := range cases {
		if got := vbfVersionOf(c.name, c.inner); got != c.want {
			t.Errorf("vbfVersionOf(%q, %q) = %s, queria %s", c.name, c.inner, got, c.want)
		}
	}
}

// Overlay cobre TODAS as localizações da entrada — sem isso a coluna
// Original sairia em um único idioma, porque os readers varrem as pastas.
func TestVbfOverlayPaths_TodasAsLocalizacoes(t *testing.T) {
	events := vbfTarget{Kind: KindEvents, ID: "znkd1400", Version: common.GameVersionFFX}
	clicked := "ffx_ps2/ffx/master/new_uspc/event/obj_ps3/zn/znkd1400/znkd1400.bin"
	paths := vbfOverlayPaths(events, clicked)

	langs := common.SupportedLanguageCodes()
	if len(paths) != len(langs) {
		// O clicado é uma das variantes e cai na deduplicação: o total é
		// um caminho por idioma.
		t.Fatalf("overlay com %d caminhos, esperava %d (um por idioma)", len(paths), len(langs))
	}
	seen := map[string]bool{}
	for _, p := range paths {
		if seen[normPath(p)] {
			t.Fatalf("caminho repetido no overlay: %s", p)
		}
		seen[normPath(p)] = true
	}
	if !seen[normPath(clicked)] {
		t.Fatal("o arquivo clicado não entrou no overlay")
	}
	for _, loc := range langs {
		want, ok := originalRelPathLoc(KindEvents, events.ID, events.Version, loc)
		if !ok {
			t.Fatalf("originalRelPathLoc falhou para %s", loc)
		}
		if !seen[normPath(want)] {
			t.Errorf("overlay sem a variante de %s (%s)", loc, want)
		}
	}
}

// Textura é um arquivo único: o overlay não deve varrer localização nenhuma.
func TestVbfOverlayPaths_TexturaUnica(t *testing.T) {
	id := "gamedata/ps3data/yonishi_data/dat_et/et_ffx/tex/d3d11/15040_19_0_0_128_128"
	clicked := ddsphyre.RelPath(common.GameVersionFFX, id)
	paths := vbfOverlayPaths(vbfTarget{Kind: KindImages, ID: id, Version: common.GameVersionFFX}, clicked)
	if len(paths) != 1 {
		t.Fatalf("overlay de textura com %d caminhos, queria 1: %v", len(paths), paths)
	}
	if !eqPath(paths[0], clicked) {
		t.Fatalf("overlay = %s, queria %s", paths[0], clicked)
	}
}

// lockit: um arquivo por idioma, todos do layout — nenhuma pasta de
// localização no caminho.
func TestVbfOverlayPaths_LockitPorIdioma(t *testing.T) {
	target := vbfTarget{Kind: KindLockit, ID: "ffx_loc_kit_ps3", Version: common.GameVersionFFX}
	clicked := "ffx_data/gamedata/ps3data/lockit/ffx_loc_kit_ps3_ch.bin"
	paths := vbfOverlayPaths(target, clicked)

	l, ok := lockit.LayoutForID(common.GameVersionFFX, "ffx_loc_kit_ps3")
	if !ok {
		t.Fatal("layout de lockit do ffx não encontrado")
	}
	langs := lockitLangs(l)
	if len(paths) != len(langs) {
		t.Fatalf("overlay de lockit com %d caminhos, esperava %d (um por idioma)",
			len(paths), len(langs))
	}
	if !eqPath(paths[0], clicked) && !containsPath(paths, clicked) {
		t.Errorf("o arquivo clicado não entrou no overlay de lockit")
	}
	for _, p := range paths {
		if !strings.Contains(normPath(p), "/lockit/") || !strings.HasSuffix(normPath(p), ".bin") {
			t.Errorf("overlay de lockit com caminho fora do formato: %s", p)
		}
	}
}

func containsPath(paths []string, want string) bool {
	for _, p := range paths {
		if eqPath(p, want) {
			return true
		}
	}
	return false
}
