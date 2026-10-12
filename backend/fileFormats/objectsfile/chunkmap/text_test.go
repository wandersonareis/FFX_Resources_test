package chunkmap_test

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
	"ffxresources/backend/fileFormats/objectsfile/chunkmap"
	"ffxresources/backend/fileFormats/objectsfile/schema/ffx"
	"ffxresources/backend/fileFormats/objectsfile/schema/ffx2"
	"ffxresources/backend/models"
)

type pairTextChunk struct {
	Pair models.TextPair
}

type sequentialTextChunk struct {
	First  models.TextRef
	Second models.TextRef
	Third  models.TextRef
}

func makePairTextBinary(t *testing.T, standard, simplified string) []byte {
	t.Helper()
	charset := common.LanguageCodeToCharset(lang)
	standardBytes, err := converter.StringToBytes(standard, charset, common.GameVersionFFX)
	if err != nil {
		t.Fatal(err)
	}
	simplifiedBytes, err := converter.StringToBytes(simplified, charset, common.GameVersionFFX)
	if err != nil {
		t.Fatal(err)
	}
	stringTable := append(append([]byte(nil), standardBytes...), 0)
	simplifiedOffset := uint16(len(stringTable))
	stringTable = append(stringTable, simplifiedBytes...)
	stringTable = append(stringTable, 0)

	chunk := make([]byte, 8)
	_ = models.WriteSegmentAt(chunk, 0, models.Segment{Offset: 0, Key: 1})
	_ = models.WriteSegmentAt(chunk, 4, models.Segment{Offset: models.Offset(simplifiedOffset), Key: 1})
	data := make([]byte, 20+len(chunk)+len(stringTable))
	binary.LittleEndian.PutUint16(data[10:12], 0)
	binary.LittleEndian.PutUint16(data[12:14], uint16(len(chunk)))
	binary.LittleEndian.PutUint16(data[14:16], uint16(len(chunk)))
	copy(data[20:], chunk)
	copy(data[20+len(chunk):], stringTable)
	return data
}

func makeSequentialTextBinary(t *testing.T, values ...string) []byte {
	t.Helper()
	if len(values) != 3 {
		t.Fatalf("expected 3 texts, got %d", len(values))
	}
	charset := common.LanguageCodeToCharset(lang)
	stringTable := make([]byte, 0)
	segments := make([]models.Segment, len(values))
	for i, text := range values {
		encoded, err := converter.StringToBytes(text, charset, common.GameVersionFFX)
		if err != nil {
			t.Fatal(err)
		}
		segments[i] = models.Segment{Offset: models.Offset(len(stringTable)), Key: 1}
		stringTable = append(stringTable, encoded...)
		stringTable = append(stringTable, 0)
	}
	chunk := make([]byte, 12)
	for i, segment := range segments {
		if err := models.WriteSegmentAt(chunk, i*4, segment); err != nil {
			t.Fatal(err)
		}
	}
	data := make([]byte, 20+len(chunk)+len(stringTable))
	binary.LittleEndian.PutUint16(data[10:12], 0)
	binary.LittleEndian.PutUint16(data[12:14], uint16(len(chunk)))
	binary.LittleEndian.PutUint16(data[14:16], uint16(len(chunk)))
	copy(data[20:], chunk)
	copy(data[20+len(chunk):], stringTable)
	return data
}

func TestEditingOneCharacterMovesFollowingTextPointersByOne(t *testing.T) {
	original := makeSequentialTextBinary(t, "abcd", "X", "yz")
	file, err := chunkmap.LoadFileFromBytes[sequentialTextChunk](
		common.GameVersionFFX, "sequential.bin", original,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, originalChunks, _, err := chunkmap.Split(original, common.GameVersionFFX)
	if err != nil {
		t.Fatal(err)
	}
	originalChunk, err := chunkmap.Decode[sequentialTextChunk](originalChunks)
	if err != nil {
		t.Fatal(err)
	}
	if got := []models.Offset{originalChunk.Value.First.Offset, originalChunk.Value.Second.Offset, originalChunk.Value.Third.Offset}; !reflect.DeepEqual(got, []models.Offset{0, 5, 7}) {
		t.Fatalf("fixture offsets = %v, want [0 5 7]", got)
	}

	if err := file.Objects[0].ImportText([]chunkmap.TextField{{
		Index: 0,
		Name:  "First.Standard",
		Texts: map[string]string{lang: "abcde"},
	}}); err != nil {
		t.Fatal(err)
	}
	updated, err := file.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	_, chunks, stringTable, err := chunkmap.Split(updated, common.GameVersionFFX)
	if err != nil {
		t.Fatal(err)
	}
	updatedChunk, err := chunkmap.Decode[sequentialTextChunk](chunks)
	if err != nil {
		t.Fatal(err)
	}
	gotOffsets := []models.Offset{updatedChunk.Value.First.Offset, updatedChunk.Value.Second.Offset, updatedChunk.Value.Third.Offset}
	if want := []models.Offset{0, 6, 8}; !reflect.DeepEqual(gotOffsets, want) {
		t.Fatalf("after adding one character offsets = %v, want %v", gotOffsets, want)
	}
	wantTable := make([]byte, 0)
	for _, text := range []string{"abcde", "X", "yz"} {
		encoded, err := converter.StringToBytes(text, common.LanguageCodeToCharset(lang), common.GameVersionFFX)
		if err != nil {
			t.Fatal(err)
		}
		wantTable = append(wantTable, encoded...)
		wantTable = append(wantTable, 0)
	}
	if got, want := stringTable, wantTable; !bytes.Equal(got, want) {
		t.Fatalf("rebuilt string table = %v, want %v", got, want)
	}
}

func TestTextPairExportsOnlyDistinctSimplifiedAndImportsItFromStandard(t *testing.T) {
	t.Run("equal texts export once and standard import updates both refs", func(t *testing.T) {
		data := makePairTextBinary(t, "same", "same")
		file, err := chunkmap.LoadFileFromBytes[pairTextChunk](common.GameVersionFFX, "pair.bin", data)
		if err != nil {
			t.Fatal(err)
		}
		fields, err := file.ExportText(lang)
		if err != nil {
			t.Fatal(err)
		}
		if len(fields) != 1 || len(fields[0].Fields) != 1 {
			t.Fatalf("equal pair exported as separate translations: %#v", fields)
		}
		standard := fields[0].Fields[0]
		if standard.Name != "Pair.Standard" || standard.Texts[lang] != "same" {
			t.Fatalf("unexpected standard export: %#v", standard)
		}
		standard.Texts[lang] = "traduzido"
		if err := file.ImportText(fields); err != nil {
			t.Fatal(err)
		}
		out, err := file.ToBytes()
		if err != nil {
			t.Fatal(err)
		}
		reloaded, err := chunkmap.LoadFileFromBytes[pairTextChunk](common.GameVersionFFX, "pair.bin", out)
		if err != nil {
			t.Fatal(err)
		}
		got, err := reloaded.ExportText(lang)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0].Fields[0].Texts[lang] != "traduzido" {
			t.Fatalf("standard translation was not applied to the pair: %#v", got)
		}
		_, chunks, stringTable, err := chunkmap.Split(out, common.GameVersionFFX)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := chunkmap.Decode[pairTextChunk](chunks)
		if err != nil {
			t.Fatal(err)
		}
		segments := decoded.Value.Pair.ExportTextSegments()
		charset := common.LanguageCodeToCharset(lang)
		for i, segment := range segments {
			got := converter.BytesToString(converter.GetStringBytesAtLookupOffset(stringTable, int(segment.Segment.Offset)), charset, common.GameVersionFFX)
			if got != "traduzido" {
				t.Fatalf("pair part %d = %q, want both refs to resolve to the standard translation", i, got)
			}
		}
	})

	t.Run("export is scoped to the requested localization", func(t *testing.T) {
		otherLanguage := "fr"
		file, err := chunkmap.LoadFileFromBytes[pairTextChunk](common.GameVersionFFX, "pair.bin", makePairTextBinary(t, "same", "same"))
		if err != nil {
			t.Fatal(err)
		}
		if err := file.Objects[0].ImportText([]chunkmap.TextField{{
			Index: 0,
			Name:  "Pair.Standard",
			Texts: map[string]string{lang: "same", otherLanguage: "A"},
		}}); err != nil {
			t.Fatal(err)
		}
		texts, err := file.ExportText(otherLanguage)
		if err != nil {
			t.Fatal(err)
		}
		if len(texts) != 1 || len(texts[0].Fields) != 1 || texts[0].Fields[0].Texts[otherLanguage] != "A" || len(texts[0].Fields[0].Texts) != 1 {
			t.Fatalf("export returned other localizations: %#v", texts)
		}
	})

	t.Run("different texts export simplified separately", func(t *testing.T) {
		data := makePairTextBinary(t, "standard", "simplified")
		file, err := chunkmap.LoadFileFromBytes[pairTextChunk](common.GameVersionFFX, "pair.bin", data)
		if err != nil {
			t.Fatal(err)
		}
		fields, err := file.ExportText(lang)
		if err != nil {
			t.Fatal(err)
		}
		if len(fields) != 1 || len(fields[0].Fields) != 2 {
			t.Fatalf("distinct pair should export both texts: %#v", fields)
		}
		if fields[0].Fields[0].Name != "Pair.Standard" || fields[0].Fields[1].Name != "Pair.Simplified" {
			t.Fatalf("unexpected pair field names: %#v", fields[0].Fields)
		}
	})
}

func checkTextRoundTrip[T any](t *testing.T, patternPath string, version common.GameVersion) {
	t.Helper()
	relPath, data := readBinary(t, version, patternPath)
	h, chunks, strtab, err := chunkmap.Split(data, version)
	if err != nil {
		t.Fatalf("%s: %v", relPath, err)
	}
	obj, err := chunkmap.NewMappedTextObject[T](h.Chunk(chunks, 0), strtab, h.IndividualLength, lang, version, patternPath)
	if err != nil {
		t.Fatalf("NewMappedTextObject: %v", err)
	}

	first, err := obj.ExportText(lang)
	if err != nil {
		t.Fatalf("ExportText: %v", err)
	}
	if len(first) == 0 {
		t.Fatal("ExportText devolveu 0 campos")
	}
	if err := obj.ImportText(first); err != nil {
		t.Fatalf("ImportText: %v", err)
	}
	second, err := obj.ExportText(lang)
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
	if !bytes.Equal(out, h.Chunk(chunks, 0)) {
		t.Errorf("ToBytes nao reproduz o chunk original apos import identidade")
	}

	fieldNames := obj.OrderedFieldNames()
	unknown := []chunkmap.TextField{
		{Index: 0, Name: fieldNames[0], Texts: map[string]string{lang: "probe"}},
		{Index: len(fieldNames), Name: "unknown", Texts: map[string]string{lang: "x"}},
	}
	if err := obj.ImportText(unknown); err == nil {
		t.Fatal("ImportText aceitou indice desconhecido junto com campo válido")
	}
	out, err = obj.ToBytes(lang)
	if err != nil {
		t.Fatalf("ToBytes apos campo desconhecido: %v", err)
	}
	if !bytes.Equal(out, h.Chunk(chunks, 0)) {
		t.Errorf("import invalido alterou o chunk")
	}
}

func TestFFXTextExportImport(t *testing.T) {
	for _, file := range []string{"command", "item", "important", "panel", "monster1", "btl_txt", "ply_rom"} {
		t.Run(file, func(t *testing.T) {
			path := ffxKernelPath + file + ".bin"
			switch file {
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
	for _, file := range []string{"command", "item", "monmagic", "plate", "job", "accessory", "monster", "menu_txt", "btlend_txt"} {
		t.Run(file, func(t *testing.T) {
			path := ffxKernelPath + file + ".bin"
			switch file {
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
	for _, file := range []string{"lm_accesary", "lm_command", "lm_monmagic", "lm_monster", "lm_item", "lm_dress", "lm_trap", "lm_mes", "lm_player", "lm_warehouse", "lm_floorname"} {
		t.Run(file, func(t *testing.T) {
			path := lastMissPath + file + ".bin"
			switch file {
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

func TestFieldsExposeStructNamesAndTextPairParts(t *testing.T) {
	all, err := chunkmap.Fields[ffx.Command]()
	if err != nil {
		t.Fatal(err)
	}
	text, err := chunkmap.SegmentFields[ffx.Command]()
	if err != nil {
		t.Fatal(err)
	}
	var wantText []string = []string{"Name.Standard", "Name.Simplified", "Desc.Standard", "Desc.Simplified"}
	if got := namesOf(text); !reflect.DeepEqual(got, wantText) {
		t.Fatalf("text field names = %v, want %v", got, wantText)
	}
	if len(text) == 0 || len(all) <= len(text) {
		t.Fatalf("Fields devolveu %d campos, SegmentFields %d — deveria haver dados", len(all), len(text))
	}

	fieldsByPath := make(map[string]chunkmap.Field, len(all))
	for _, field := range all {
		fieldsByPath[field.Name] = field
	}
	for _, name := range []string{"Anim1", "Anim2", "SubMenuCat2", "MpCost", "DmgFormula", "StatusMap.Death"} {
		field, ok := fieldsByPath[name]
		if !ok {
			t.Errorf("campo de dado %q ausente", name)
			continue
		}
		if field.Kind != chunkmap.KindData {
			t.Errorf("campo %q deveria ser KindData", name)
		}
	}
}

func namesOf(fields []chunkmap.Field) []string {
	names := make([]string, len(fields))
	for i, field := range fields {
		names[i] = field.Name
	}
	return names
}
