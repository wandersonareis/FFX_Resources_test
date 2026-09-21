package converter_test

import (
	"errors"
	"reflect"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
	"ffxresources/backend/core/encoding"
)

// seedTestMaps publica mapas mínimos us/ffx+ffx2 e devolve restore.
func seedTestMaps(t *testing.T) func() {
	t.Helper()
	prevFFX := ffxencoding.GetByteToCharMap(common.GameVersionFFX, "us")
	prevFFXRev := ffxencoding.GetCharToByteMap(common.GameVersionFFX, "us")
	prevFFX2 := ffxencoding.GetByteToCharMap(common.GameVersionFFX2, "us")
	prevFFX2Rev := ffxencoding.GetCharToByteMap(common.GameVersionFFX2, "us")
	b2c := map[uint]rune{0x50: 'A', 0x51: 'B'}
	c2b := map[rune]uint{'A': 0x50, 'B': 0x51}
	ffxencoding.SetCharMap(common.GameVersionFFX, "us", b2c, c2b)
	ffxencoding.SetCharMap(common.GameVersionFFX2, "us", b2c, c2b)
	return func() {
		ffxencoding.SetCharMap(common.GameVersionFFX, "us", prevFFX, prevFFXRev)
		ffxencoding.SetCharMap(common.GameVersionFFX2, "us", prevFFX2, prevFFX2Rev)
	}
}

func TestLastMissUsesFFX2Maps(t *testing.T) {
	restore := seedTestMaps(t)
	defer restore()

	want, _ := converter.StringToBytes("AB", "us", common.GameVersionFFX2)
	got, _ := converter.StringToBytes("AB", "us", common.GameVersionLastMiss)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("lastmiss encode != ffx2 encode: %v vs %v", got, want)
	}
	back := converter.BytesToString([]byte{0x50, 0x51, 0x00}, "us", common.GameVersionLastMiss)
	if back != "AB" {
		t.Fatalf("lastmiss decode: got %q", back)
	}
}

func TestUnknownVersionPanics(t *testing.T) {
	restore := seedTestMaps(t)
	defer restore()

	for name, fn := range map[string]func(){
		"StringToBytes":    func() { converter.StringToBytes("A", "us", common.GameVersion{}) },
		"BytesToString":    func() { converter.BytesToString([]byte{0x50, 0x00}, "us", common.GameVersion{}) },
		"StringToByteList": func() { converter.StringToByteList([]rune("A"), "us", common.GameVersion{}) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic for %s with zero version", name)
				}
			}()
			fn()
		}()
	}
}

func TestCharToBytesDistinguishesErrors(t *testing.T) {
	restore := seedTestMaps(t)
	defer restore()

	if _, err := converter.CharToBytes('€', "us", common.GameVersionFFX); !errors.Is(err, ffxencoding.ErrCharNotFound) {
		t.Fatalf("unknown char must wrap ErrCharNotFound: %v", err)
	}
	if _, err := converter.CharToBytes('A', "xx", common.GameVersionFFX); !errors.Is(err, ffxencoding.ErrCharsetNotFound) {
		t.Fatalf("unknown charset must wrap ErrCharsetNotFound: %v", err)
	}
	if _, err := converter.CharToBytes('A', "us", common.GameVersion{}); !errors.Is(err, ffxencoding.ErrVersionNotFound) {
		t.Fatalf("zero version must wrap ErrVersionNotFound (strict, no panic here): %v", err)
	}
}

func TestStringToByteListAbortsOnlyOnConfig(t *testing.T) {
	restore := seedTestMaps(t)
	defer restore()

	if _, err := converter.StringToByteList([]rune("A"), "xx", common.GameVersionFFX); !errors.Is(err, ffxencoding.ErrCharsetNotFound) {
		t.Fatalf("unknown charset must abort: %v", err)
	}
	got, err := converter.StringToByteList([]rune("A€B"), "us", common.GameVersionFFX)
	if err != nil {
		t.Fatalf("unknown char must not abort: %v", err)
	}
	if !reflect.DeepEqual(got, []byte{0x50, 0x51}) {
		t.Fatalf("unknown char must be skipped, rest kept: %v", got)
	}
}

func TestPUARunesRoundTrip(t *testing.T) {
	restore := seedTestMaps(t)
	defer restore()

	pua := rune(0xE0B5)
	// Fora dos mapas: bijeção PUA→código, sem erro.
	got, err := converter.CharToBytes(pua, "us", common.GameVersionFFX)
	if err != nil {
		t.Fatalf("PUA rune must not error: %v", err)
	}
	if !reflect.DeepEqual(got, []uint{0xB5}) {
		t.Fatalf("PUA fallback must emit slot byte: %v", got)
	}
	// Mapa tem prioridade sobre a bijeção.
	ffxencoding.SetCharMap(common.GameVersionFFX, "us",
		map[uint]rune{0x50: 'A', 0x99: pua},
		map[rune]uint{'A': 0x50, pua: 0x99})
	got, err = converter.CharToBytes(pua, "us", common.GameVersionFFX)
	if err != nil {
		t.Fatalf("mapped PUA must not error: %v", err)
	}
	if !reflect.DeepEqual(got, []uint{0x99}) {
		t.Fatalf("mapped PUA must win over fallback: %v", got)
	}
	// Round-trip completo pelo mapa.
	raw, err := converter.StringToByteList([]rune{pua}, "us", common.GameVersionFFX)
	if err != nil {
		t.Fatalf("PUA list must not error: %v", err)
	}
	back := converter.BytesToString(append(raw, 0x00), "us", common.GameVersionFFX)
	if back != string([]rune{pua}) {
		t.Fatalf("PUA round-trip mismatch: %q", back)
	}
}
