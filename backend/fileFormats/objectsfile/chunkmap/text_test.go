package chunkmap_test

import (
	"bytes"
	"reflect"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/fileFormats/objectsfile/chunkmap"
	"ffxresources/backend/fileFormats/objectsfile/schema/ffx"
	"ffxresources/backend/fileFormats/objectsfile/schema/ffx2"
)

/*
A struct é quem declara seus campos de texto (TextRef/TextPair): as chaves
saem do nome do campo e ninguém de fora passa nomes como "name"/"help".

checkTextRoundTrip confere, sobre o primeiro chunk de um binário real:

 1. Export -> Import -> Export devolve os MESMOS campos, na mesma ordem;
 2. objectsfile.ExportFieldTexts (o caminho dos builders, compartilhado com a
    camada Wails) devolve exatamente o mesmo resultado;
 3. ToBytes continua byte-exato depois do round-trip (importar o que foi
    exportado não muda nada);
 4. chave desconhecida é rejeitada e não corrompe o objeto;
 5. o caminho dos builders (ApplyFieldTexts) também rejeita/ignora a mesma
    chave sem corromper o objeto.
*/
func checkTextRoundTrip[T any](t *testing.T, patternPath string, version common.GameVersion) {
	t.Helper()

	relPath, data := readBinary(t, version, patternPath)
	h, chunks, strtab, err := chunkmap.Split(data, version)
	if err != nil {
		t.Fatalf("%s: %v", relPath, err)
	}
	raw := h.Chunk(chunks, 0)

	obj, err := chunkmap.NewMappedTextObject[T](raw, strtab, h.IndividualLength, lang, version, patternPath)
	if err != nil {
		t.Fatalf("NewMappedTextObject: %v", err)
	}

	first, err := obj.ExportText()
	if err != nil {
		t.Fatalf("ExportText: %v", err)
	}
	if len(first) == 0 {
		t.Fatalf("ExportText devolveu 0 campos")
	}
	if got := objectsfile.ExportFieldTexts(obj); !reflect.DeepEqual(got, first) {
		t.Errorf("ExportFieldTexts = %v, a struct exporta %v", got, first)
	}
	if err := obj.ImportText(first); err != nil {
		t.Fatalf("ImportText: %v", err)
	}
	second, err := obj.ExportText()
	if err != nil {
		t.Fatalf("ExportText apos import: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Errorf("export apos import difere:\n  antes %v\n  depois %v", first, second)
	}
	out, err := obj.ToBytes(lang)
	if err != nil {
		t.Fatalf("ToBytes: %v", err)
	}
	if !bytes.Equal(out, raw) {
		t.Errorf("ToBytes nao reproduz os bytes originais apos round-trip")
	}

	unknown := []objectsfile.FieldText{{Key: "no_such_field", Texts: map[string]string{lang: "x"}}}
	if err := obj.ImportText(unknown); err == nil {
		t.Errorf("ImportText aceitou chave desconhecida")
	}
	// o caminho dos builders ignora a chave desconhecida (loga e segue); o
	// objeto não pode ter sido corrompido no caminho
	objectsfile.ApplyFieldTexts(obj, first, version)
	out, err = obj.ToBytes(lang)
	if err != nil {
		t.Fatalf("ToBytes apos import via builder: %v", err)
	}
	if !bytes.Equal(out, raw) {
		t.Errorf("ToBytes mudou apos import via builder")
	}
}

func TestFFXTextExportImport(t *testing.T) {
	for _, c := range []struct {
		file string
		size int
	}{
		{"command", 96},
		{"item", 96},
		{"important", 20},
		{"panel", 24},
		{"monster1", 128},
		{"btl_txt", 8},
		{"ply_rom", 44},
	} {
		c := c
		t.Run(c.file, func(t *testing.T) {
			path := ffxKernelPath + c.file + ".bin"
			switch c.file {
			case "command", "item":
				checkTextRoundTrip[ffx.Command](t, path, common.GameVersionFFX)
			case "important":
				checkTextRoundTrip[ffx.KeyItem](t, path, common.GameVersionFFX)
			case "panel":
				checkTextRoundTrip[ffx.SphereGridNodeType](t, path, common.GameVersionFFX)
			case "monster1":
				checkTextRoundTrip[ffx.MonStats](t, path, common.GameVersionFFX)
			case "btl_txt":
				checkTextRoundTrip[ffx.HelpText](t, path, common.GameVersionFFX)
			case "ply_rom":
				checkTextRoundTrip[ffx.PlyRom](t, path, common.GameVersionFFX)
			}
		})
	}
}

func TestFFX2TextExportImport(t *testing.T) {
	for _, c := range []struct {
		file string
	}{
		{"command"}, {"item"}, {"monmagic"}, {"plate"}, {"job"},
		{"accessory"}, {"monster"}, {"menu_txt"}, {"btlend_txt"},
	} {
		c := c
		t.Run(c.file, func(t *testing.T) {
			path := ffxKernelPath + c.file + ".bin"
			switch c.file {
			case "command":
				checkTextRoundTrip[ffx2.PCommand](t, path, common.GameVersionFFX2)
			case "item":
				checkTextRoundTrip[ffx2.Item](t, path, common.GameVersionFFX2)
			case "monmagic":
				checkTextRoundTrip[ffx2.MCommand](t, path, common.GameVersionFFX2)
			case "plate":
				checkTextRoundTrip[ffx2.Plate](t, path, common.GameVersionFFX2)
			case "job":
				checkTextRoundTrip[ffx2.Job](t, path, common.GameVersionFFX2)
			case "accessory":
				checkTextRoundTrip[ffx2.Accessory](t, path, common.GameVersionFFX2)
			case "monster":
				checkTextRoundTrip[ffx2.Monster](t, path, common.GameVersionFFX2)
			case "menu_txt":
				checkTextRoundTrip[ffx2.MenuTxt](t, path, common.GameVersionFFX2)
			case "btlend_txt":
				checkTextRoundTrip[ffx2.BtlEndTxt](t, path, common.GameVersionFFX2)
			}
		})
	}
}

func TestLastMissTextExportImport(t *testing.T) {
	for _, c := range []struct {
		file string
	}{
		{"lm_accesary"}, {"lm_command"}, {"lm_monmagic"}, {"lm_monster"},
		{"lm_item"}, {"lm_dress"}, {"lm_trap"}, {"lm_mes"}, {"lm_player"},
		{"lm_warehouse"}, {"lm_floorname"},
	} {
		c := c
		t.Run(c.file, func(t *testing.T) {
			path := lastMissPath + c.file + ".bin"
			switch c.file {
			case "lm_accesary":
				checkTextRoundTrip[ffx2.LmAccesary](t, path, common.GameVersionLastMiss)
			case "lm_command":
				checkTextRoundTrip[ffx2.LmCommand](t, path, common.GameVersionLastMiss)
			case "lm_monmagic":
				checkTextRoundTrip[ffx2.LmMonMagic](t, path, common.GameVersionLastMiss)
			case "lm_monster":
				checkTextRoundTrip[ffx2.LmMonster](t, path, common.GameVersionLastMiss)
			case "lm_item":
				checkTextRoundTrip[ffx2.LmItem](t, path, common.GameVersionLastMiss)
			case "lm_dress":
				checkTextRoundTrip[ffx2.LmDress](t, path, common.GameVersionLastMiss)
			case "lm_trap":
				checkTextRoundTrip[ffx2.LmTrap](t, path, common.GameVersionLastMiss)
			case "lm_mes":
				checkTextRoundTrip[ffx2.LmMes](t, path, common.GameVersionLastMiss)
			case "lm_player":
				checkTextRoundTrip[ffx2.LmPlayer](t, path, common.GameVersionLastMiss)
			case "lm_warehouse":
				checkTextRoundTrip[ffx2.LmWarehouse](t, path, common.GameVersionLastMiss)
			case "lm_floorname":
				checkTextRoundTrip[ffx2.LmFloorName](t, path, common.GameVersionLastMiss)
			}
		})
	}
}

// TestFieldsECamposDeTexto garante que SegmentFields é exatamente o filtro
// KindText de Fields, na mesma ordem, e que os campos de dado existem com
// chave própria (é a base para exportar dados no futuro).
func TestFieldsECamposDeTexto(t *testing.T) {
	all, err := chunkmap.Fields[ffx.Command]()
	if err != nil {
		t.Fatalf("Fields: %v", err)
	}
	text, err := chunkmap.SegmentFields[ffx.Command]()
	if err != nil {
		t.Fatalf("SegmentFields: %v", err)
	}
	var want []chunkmap.Field
	for _, f := range all {
		if f.Kind == chunkmap.KindText {
			want = append(want, f)
		}
	}
	if !reflect.DeepEqual(text, want) {
		t.Errorf("SegmentFields != Fields|KindText:\n  %v\n  %v", text, want)
	}
	if len(text) == 0 || len(all) <= len(text) {
		t.Fatalf("Fields devolveu %d campos, SegmentFields %d — deveria haver dados",
			len(all), len(text))
	}
	if got, wantKeys := keysOf(text), []string{"name", "name_simplified", "desc", "desc_simplified"}; !reflect.DeepEqual(got, wantKeys) {
		t.Errorf("chaves de texto = %v, esperado %v", got, wantKeys)
	}
	// dados na mesma ordem do arquivo: os primeiros vêm logo após os textos
	byKey := map[string]chunkmap.Field{}
	for _, f := range all {
		byKey[f.Key] = f
	}
	for _, k := range []string{"anim1", "anim2", "sub_menu_cat2", "mp_cost", "dmg_formula"} {
		f, ok := byKey[k]
		if !ok {
			t.Errorf("campo de dado %q ausente de Fields", k)
			continue
		}
		if f.Kind != chunkmap.KindData {
			t.Errorf("campo %q deveria ser KindData", k)
		}
	}
	// array de dado vira um campo por elemento
	if _, ok := byKey["status_map_death"]; !ok {
		t.Errorf("sub-struct de dado sem prefixo: esperado status_map_death em %v", keysOf(all))
	}
}

// keysOf devolve só as chaves, para mensagens de erro legíveis.
func keysOf(fields []chunkmap.Field) []string {
	out := make([]string, len(fields))
	for i, f := range fields {
		out[i] = f.Key
	}
	return out
}
