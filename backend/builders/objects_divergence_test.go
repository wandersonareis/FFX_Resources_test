package builders_test

import (
	"testing"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/converter"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/datastore"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/formatters/hash"
)

func TestApplyObjectsDivergentCopyStaysIndependent(t *testing.T) {
	version := common.GameVersionFFX2
	if err := ffxencoding.PrepareVersionCharsets(version); err != nil {
		t.Skipf("charsets: %v", err)
	}
	charset := ffxencoding.GetCharsetForLanguage(common.DefaultLocalization)
	const original = "Same object field text across records"
	const definition = "Object definition translation A"
	const divergent = "Object copy-specific translation B"
	bare := hash.Sum64Hex(original)

	object := func(text string) datastore.IGlobalLocalizedTextObject {
		bytes, err := converter.StringToBytes(text, charset, version)
		if err != nil {
			t.Fatalf("encode fixture: %v", err)
		}
		name := objectsfile.NewLocalizedKeyedStringObjectWithContent(common.DefaultLocalization, &objectsfile.KeyedString{
			Charset: charset,
			Version: version,
			Bytes:   bytes,
		})
		return &objectsfile.NameDescriptionEffectAbilityTextObject{
			Name:        name,
			Description: objectsfile.NewLocalizedKeyedStringObject(),
			Effect:      objectsfile.NewLocalizedKeyedStringObject(),
			Version:     version,
		}
	}
	objects := components.NewList[datastore.IGlobalLocalizedTextObject](3)
	objects.Add(object(original))
	objects.Add(object(original))
	objects.Add(object(original))

	entry := dto.FileEntry{Rows: []dto.TextRow{
		{Index: 0, Name: "name", Hash: map[string]string{"us": bare}, Text: map[string]string{"us": definition}},
		{Index: 1, Name: "name", Hash: map[string]string{"us": bare}, Text: map[string]string{"us": divergent}, Divergent: true},
		{Index: 2, Name: "name", Hash: map[string]string{"us": bare}, Text: map[string]string{"us": hash.Prefix(bare)}},
	}}
	if err := builders.ApplyObjectsEntry(objects, version, "", entry); err != nil {
		t.Fatalf("apply: %v", err)
	}
	want := []string{definition, divergent, definition}
	for i, obj := range objects.Items() {
		if got := obj.GetKeyedString("name").GetLocalizedString(common.DefaultLocalization); got != want[i] {
			t.Errorf("object %d name = %q, esperado %q", i, got, want[i])
		}
	}
}
