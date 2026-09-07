package objectsfile

import (
	"bytes"
	"fmt"
	"slices"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

type TextKey string

type PlayerChar int

const (
	Tidus PlayerChar = iota
	Yuna
	Auron
	Kimahri
	Wakka
	Lulu
	Rikku

	playerCharCount // 7
)

const WeaponsNameTextObjectLength = playerCharCount * 2 * 4 // 0x38

type WeaponsNameTextObject struct {
	Bytes           []byte
	Names           [playerCharCount]datastore.IGlobalLocalizedKeyedStringObject
	SimplifiedNames [playerCharCount]datastore.IGlobalLocalizedKeyedStringObject
	HeaderLength    int
}

var weaponRefs = []struct {
	key  string
	name string
}{
	{"T", "Tidus"},
	{"Y", "Yuna"},
	{"A", "Auron"},
	{"K", "Kimahri"},
	{"W", "Wakka"},
	{"L", "Lulu"},
	{"R", "Rikku"},
}

func NewWeaponsNameTextObject(bytes []byte, stringBytes []byte, headerLength int, languageCode string) (*WeaponsNameTextObject, error) {
	if common.GetGameVersionString() != "ffx" {
		return nil, fmt.Errorf("WeaponsNameTextObject is only compatible with FFX")
	}
	
	if len(bytes) < headerLength {
		return nil, fmt.Errorf("insufficient data: have %d bytes, need at least %d", len(bytes), headerLength)
	}

	w := &WeaponsNameTextObject{Bytes: bytes, HeaderLength: headerLength}
	for i := range w.Names {
		w.Names[i] = NewLocalizedKeyedStringObject()
		w.SimplifiedNames[i] = NewLocalizedKeyedStringObject()
	}

	if err := w.mapBytes(stringBytes, languageCode); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *WeaponsNameTextObject) mapBytes(stringBytes []byte, languageCode string) error {
	r := bytes.NewReader(w.Bytes)
	if err := readStringSegments(r, stringBytes, languageCode, w.Names[:]...); err != nil {
		return fmt.Errorf("reading names: %w", err)
	}
	return readStringSegments(r, stringBytes, languageCode, w.SimplifiedNames[:]...)
}

func (w *WeaponsNameTextObject) ToBytes(languageCode string) ([]byte, error) {
	data := slices.Clone(w.Bytes)

	segs := make([]datastore.IGlobalLocalizedKeyedStringObject, 0, len(w.Names)+len(w.SimplifiedNames))
	segs = append(segs, w.Names[:]...)
	segs = append(segs, w.SimplifiedNames[:]...)

	return data, writeStringSegments(data, 0, languageCode, segs...)
}

var weaponKeyIndex = map[string]PlayerChar{
	"T": Tidus, "Y": Yuna, "A": Auron, "K": Kimahri,
	"W": Wakka, "L": Lulu, "R": Rikku,
}

func (w *WeaponsNameTextObject) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	if i, ok := weaponKeyIndex[title]; ok {
		return w.Names[i]
	}
	if len(title) > 1 && title[0] == 's' {
		if i, ok := weaponKeyIndex[title[1:]]; ok {
			return w.SimplifiedNames[i]
		}
	}
	return nil
}

func (w *WeaponsNameTextObject) GetWeaponKeyedString(key TextKey) datastore.IGlobalLocalizedKeyedStringObject {
	return w.GetKeyedString(string(key))
}

func (w *WeaponsNameTextObject) GetName(languageCode string) string {
	return w.Names[Tidus].GetLocalizedString(languageCode)
}

func (w *WeaponsNameTextObject) GetHeaderLength() int {
	return w.HeaderLength
}

func (w *WeaponsNameTextObject) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return w
}

func (w *WeaponsNameTextObject) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if o, ok := other.(*WeaponsNameTextObject); ok {
		for i := range w.Names {
			o.Names[i].CopyInto(w.Names[i])
			o.SimplifiedNames[i].CopyInto(w.SimplifiedNames[i])
		}
	}
}

func (w *WeaponsNameTextObject) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	result := make([]datastore.IGlobalKeyedString, 0, len(w.Names)+len(w.SimplifiedNames))
	for i := range w.Names {
		result = append(result, w.Names[i].GetLocalizedContent(localization))
	}
	for i := range w.SimplifiedNames {
		result = append(result, w.SimplifiedNames[i].GetLocalizedContent(localization))
	}
	return result
}

func (w *WeaponsNameTextObject) ToString(languageCode string) string {
	return fmt.Sprintf("Weapons: %s", w.Names[Tidus].GetLocalizedString(languageCode))
}

func (w *WeaponsNameTextObject) String() string {
	return w.ToString(common.DefaultLocalization)
}
