package chunkmap_test

import (
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/fileFormats/objectsfile/chunkmap"
	"ffxresources/backend/fileFormats/objectsfile/schema/ffx"
	"ffxresources/backend/fileFormats/objectsfile/schema/ffx2"
)

/*
DIAGNÓSTICO DE MIGRAÇÃO — cruzamento do fluxo novo com o fluxo legado.

Este arquivo existe APENAS enquanto o fluxo legado (KeyedStringFile +
LayoutSet) não foi deletado. É deletável junto com ele: o fluxo novo é
auto-suficiente (chunkmap_test.go/text_test.go não o consultam).

A comparação é independente de rótulo: os nomes ("name", "help", ...) são
apenas rótulos para humanos e não participam. Para cada arquivo:

  1. QUANTIDADE de campos de texto: o layout legado e a struct T declaram o
     MESMO número de campos de texto — qualquer divergência falha e pede
     conferência no editor hex (qualquer dos lados pode estar errado);
  2. IGUALDADE dos valores extraídos: as sequências de textos (posiçional,
     ordem do arquivo, apenas conteúdos não-nil) são iguais nos dois lados.
*/

// crossCheckConf roda o diagnóstico de um arquivo (primeiro chunk).
func crossCheckConf[T any](t *testing.T, patternPath string, version common.GameVersion) {
	t.Helper()

	relPath, data := readBinary(t, version, patternPath)
	h, chunks, strtab, err := chunkmap.Split(data, version)
	if err != nil {
		t.Fatalf("%s: %v", relPath, err)
	}
	raw := h.Chunk(chunks, 0)

	mapped, err := chunkmap.NewMappedTextObject[T](raw, strtab, h.IndividualLength, lang, version, patternPath)
	if err != nil {
		t.Fatalf("%s: NewMappedTextObject: %v", relPath, err)
	}

	fl, ok := objectsfile.FileLayoutFor(version, patternPath)
	if !ok {
		t.Fatalf("%s: FileLayouts nao registra %q", relPath, patternPath)
	}
	legacy, err := objectsfile.NewKeyedStringFile(
		raw, strtab, h.IndividualLength, lang, version,
		objectsfile.LayoutSet{version: fl.Fields}, patternPath,
	)
	if err != nil {
		t.Fatalf("%s: KeyedStringFile legado: %v", relPath, err)
	}

	// (1) quantidade de campos de texto — sem assumir de qual lado é maior
	legacyTexts := legacyTextValues(legacy, lang)
	mappedTexts := mappedTextValuesOf(mapped, lang)
	if len(legacyTexts) != len(mappedTexts) {
		t.Errorf("%s: quantidade de campos de texto difere — legado %d, novo %d; conferir no editor hex",
			relPath, len(legacyTexts), len(mappedTexts))
	}

	// (2) igualdade dos valores extraídos, posiçional
	for i, want := range legacyTexts {
		if i >= len(mappedTexts) {
			break
		}
		if got := mappedTexts[i]; got != want {
			t.Errorf("%s: campo %d: texto legado=%q novo=%q", relPath, i, want, got)
		}
	}
}

// legacyTextValues devolve os textos não-vazios do objeto legado na ordem do
// arquivo (GetLocalizedKeyedStrings pula conteúdos nil).
func legacyTextValues(f *objectsfile.KeyedStringFile, languageCode string) []string {
	out := make([]string, 0, len(f.OrderedFieldKeys()))
	for _, ks := range f.GetLocalizedKeyedStrings(languageCode) {
		out = append(out, ks.GetString())
	}
	return out
}

// mappedTextValuesOf devolve os textos não-vazios do objeto novo na ordem do
// arquivo (GetLocalizedKeyedStrings pula conteúdos nil) — a mesma semântica
// do lado legado.
func mappedTextValuesOf[T any](obj *chunkmap.MappedTextObject[T], languageCode string) []string {
	out := make([]string, 0, len(obj.OrderedFieldKeys()))
	for _, ks := range obj.GetLocalizedKeyedStrings(languageCode) {
		out = append(out, ks.GetString())
	}
	return out
}

// ---- FFX (V1) ----

func TestLegacyCrossCheckFFX(t *testing.T) {
	for _, file := range []string{"command", "item", "a_ability"} {
		t.Run(file, func(t *testing.T) {
			crossCheckConf[ffx.Command](t, ffxKernelPath+file+".bin", common.GameVersionFFX)
		})
	}
	for _, file := range []string{"monmagic1", "monmagic2"} {
		t.Run(file, func(t *testing.T) {
			crossCheckConf[ffx.MonMagic](t, ffxKernelPath+file+".bin", common.GameVersionFFX)
		})
	}
	t.Run("important", func(t *testing.T) {
		crossCheckConf[ffx.KeyItem](t, ffxKernelPath+"important.bin", common.GameVersionFFX)
	})
	t.Run("panel", func(t *testing.T) {
		crossCheckConf[ffx.SphereGridNodeType](t, ffxKernelPath+"panel.bin", common.GameVersionFFX)
	})
	t.Run("sphere", func(t *testing.T) {
		crossCheckConf[ffx.Sphere](t, ffxKernelPath+"sphere.bin", common.GameVersionFFX)
	})
	t.Run("btl_txt", func(t *testing.T) {
		crossCheckConf[ffx.HelpText](t, ffxKernelPath+"btl_txt.bin", common.GameVersionFFX)
	})
	for _, name := range []string{
		"arms_txt", "btlend_txt", "build_txt", "config_txt", "item_txt",
		"menu_txt", "mmain_txt", "name_txt", "save_txt", "status_txt", "summon_txt",
	} {
		t.Run(name, func(t *testing.T) {
			crossCheckConf[ffx.NameHelpText](t, ffxKernelPath+name+".bin", common.GameVersionFFX)
		})
	}
	t.Run("monster1", func(t *testing.T) {
		crossCheckConf[ffx.MonStats](t, ffxKernelPath+"monster1.bin", common.GameVersionFFX)
	})
	t.Run("ply_rom", func(t *testing.T) {
		crossCheckConf[ffx.PlyRom](t, ffxKernelPath+"ply_rom.bin", common.GameVersionFFX)
	})
	t.Run("ply_save", func(t *testing.T) {
		crossCheckConf[ffx.PlySave](t, ffxKernelPath+"ply_save.bin", common.GameVersionFFX)
	})
}

// ---- FFX-2 (V2) ----

func TestLegacyCrossCheckFFX2(t *testing.T) {
	for _, c := range []struct {
		file string
		typ  string
	}{
		{"command", "command"},
		{"item", "item"},
		{"monmagic", "monmagic"},
		{"important", "important"},
		{"a_ability", "a_ability"},
		{"accessory", "accessory"},
		{"job", "job"},
		{"plate", "plate"},
		{"menu_txt", "menu_txt"},
		{"monster", "monster"},
		{"monster2", "monster2"},
		{"oversoul", "oversoul"},
		{"btl_txt", "btl_txt"},
		{"btlend_txt", "btlend_txt"},
	} {
		c := c
		t.Run(c.file, func(t *testing.T) {
			switch c.typ {
			case "command":
				crossCheckConf[ffx2.PCommand](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2)
			case "item":
				crossCheckConf[ffx2.Item](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2)
			case "monmagic":
				crossCheckConf[ffx2.MCommand](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2)
			case "important":
				crossCheckConf[ffx2.KeyItem](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2)
			case "a_ability":
				crossCheckConf[ffx2.AutoAbility](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2)
			case "accessory":
				crossCheckConf[ffx2.Accessory](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2)
			case "job":
				crossCheckConf[ffx2.Job](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2)
			case "plate":
				crossCheckConf[ffx2.Plate](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2)
			case "menu_txt":
				crossCheckConf[ffx2.MenuTxt](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2)
			case "monster":
				crossCheckConf[ffx2.Monster](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2)
			case "monster2":
				crossCheckConf[ffx2.Monster2](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2)
			case "oversoul":
				crossCheckConf[ffx2.Oversoul](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2)
			case "btl_txt":
				crossCheckConf[ffx2.BtlTxt](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2)
			case "btlend_txt":
				crossCheckConf[ffx2.BtlEndTxt](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2)
			}
		})
	}
}

// ---- Last Mission (V2) ----

func TestLegacyCrossCheckLastMiss(t *testing.T) {
	for _, c := range []struct {
		file string
		typ  string
	}{
		{"lm_accesary", "lm_accesary"},
		{"lm_command", "lm_command"},
		{"lm_monmagic", "lm_monmagic"},
		{"lm_monster", "lm_monster"},
		{"lm_item", "lm_item"},
		{"lm_dress", "lm_dress"},
		{"lm_trap", "lm_trap"},
		{"lm_mes", "lm_mes"},
		{"lm_player", "lm_player"},
		{"lm_warehouse", "lm_warehouse"},
		{"lm_floorname", "lm_floorname"},
	} {
		c := c
		t.Run(c.file, func(t *testing.T) {
			switch c.typ {
			case "lm_accesary":
				crossCheckConf[ffx2.LmAccesary](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss)
			case "lm_command":
				crossCheckConf[ffx2.LmCommand](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss)
			case "lm_monmagic":
				crossCheckConf[ffx2.LmMonMagic](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss)
			case "lm_monster":
				crossCheckConf[ffx2.LmMonster](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss)
			case "lm_item":
				crossCheckConf[ffx2.LmItem](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss)
			case "lm_dress":
				crossCheckConf[ffx2.LmDress](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss)
			case "lm_trap":
				crossCheckConf[ffx2.LmTrap](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss)
			case "lm_mes":
				crossCheckConf[ffx2.LmMes](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss)
			case "lm_player":
				crossCheckConf[ffx2.LmPlayer](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss)
			case "lm_warehouse":
				crossCheckConf[ffx2.LmWarehouse](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss)
			case "lm_floorname":
				crossCheckConf[ffx2.LmFloorName](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss)
			}
		})
	}
}
