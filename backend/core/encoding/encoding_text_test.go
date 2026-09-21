package ffxencoding

import (
	"errors"
	"testing"

	"ffxresources/backend/common"
)

// testCharset é exclusivo destes testes (aditivo e inofensivo aos demais:
// nenhum teste enumera charsets de uma versão).
const testCharset = "testzz"

func seedTestCharset(t *testing.T) {
	t.Helper()
	b2c := map[uint]rune{0x50: 'A'}
	c2b := map[rune]uint{'A': 0x50}
	SetCharMap(common.GameVersionFFX, testCharset, b2c, c2b)
	SetCharMap(common.GameVersionFFX2, testCharset, b2c, c2b)
}

func TestCharToByteSentinels(t *testing.T) {
	seedTestCharset(t)

	if _, err := CharToByte('A', testCharset, common.GameVersion{}); !errors.Is(err, ErrVersionNotFound) {
		t.Fatalf("missing bucket must wrap ErrVersionNotFound: %v", err)
	}
	if _, err := CharToByte('A', "xx", common.GameVersionFFX); !errors.Is(err, ErrCharsetNotFound) {
		t.Fatalf("missing charset must wrap ErrCharsetNotFound: %v", err)
	}
	if _, err := CharToByte('€', testCharset, common.GameVersionFFX); !errors.Is(err, ErrCharNotFound) {
		t.Fatalf("missing char must wrap ErrCharNotFound: %v", err)
	}
	if b, err := CharToByte('A', testCharset, common.GameVersionFFX); err != nil || b != 0x50 {
		t.Fatalf("hit must resolve: %v %v", b, err)
	}
}

func TestByteToCharSentinels(t *testing.T) {
	seedTestCharset(t)

	if _, err := ByteToChar(0x50, testCharset, common.GameVersion{}); !errors.Is(err, ErrVersionNotFound) {
		t.Fatalf("missing bucket must wrap ErrVersionNotFound: %v", err)
	}
	if _, err := ByteToChar(0x50, "xx", common.GameVersionFFX); !errors.Is(err, ErrCharsetNotFound) {
		t.Fatalf("missing charset must wrap ErrCharsetNotFound: %v", err)
	}
	if _, err := ByteToChar(0x51, testCharset, common.GameVersionFFX); !errors.Is(err, ErrCharNotFound) {
		t.Fatalf("missing byte must wrap ErrCharNotFound: %v", err)
	}
	if c, err := ByteToChar(0x50, testCharset, common.GameVersionFFX); err != nil || c != 'A' {
		t.Fatalf("hit must resolve: %v %v", c, err)
	}
}

func TestEnsureCharsetLoaded(t *testing.T) {
	seedTestCharset(t)

	if err := EnsureCharsetLoaded(common.GameVersionFFX, testCharset); err != nil {
		t.Fatalf("seeded charset must pass: %v", err)
	}
	if err := EnsureCharsetLoaded(common.GameVersionLastMiss, testCharset); err != nil {
		t.Fatalf("lastmiss must resolve ffx2-seeded charset: %v", err)
	}
	if err := EnsureCharsetLoaded(common.GameVersionFFX, "xx"); !errors.Is(err, ErrCharsetNotFound) {
		t.Fatalf("missing charset must wrap ErrCharsetNotFound: %v", err)
	}
}

func TestEnsureCharsetLoadedPanicsOnUnknownVersion(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic for zero version")
		}
	}()
	_ = EnsureCharsetLoaded(common.GameVersion{}, testCharset)
}
