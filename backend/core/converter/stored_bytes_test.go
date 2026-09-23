package converter_test

import (
	"bytes"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
)

// StringToStoredBytes/StringToByteSize têm de bater byte a byte com
// FillByteList (incluindo o terminador 0x00): é o que garante poder medir/
// usar o texto direto — sem rebuild, arquivo ou instância de domínio — com o
// mesmo resultado do que o rebuild gravaria.
func TestStringToStoredBytesMatchesFillByteList(t *testing.T) {
	restore := seedTestMaps(t)
	defer restore()

	// O mapa seed só conhece 'A'/'B'; runes fora do mapa são ignorados nos
	// dois caminhos, e comandos convertem pelos dois via ParseCommand.
	for _, text := range []string{
		"",
		"A",
		"AB",
		"{PAUSE}A",
		"{SPACE:10}AB{BREAK}",
		"A{CHOICE:01}B",
		"{HEX:01:02}A",
		"{\\n}AB",
	} {
		var want bytes.Buffer
		converter.FillByteList(text, &want, "us", common.GameVersionFFX)

		got, err := converter.StringToStoredBytes(text, "us", common.GameVersionFFX)
		if err != nil {
			t.Fatalf("StringToStoredBytes(%q): %v", text, err)
		}
		if !bytes.Equal(got, want.Bytes()) {
			t.Fatalf("bytes divergem para %q: got %v want %v", text, got, want.Bytes())
		}

		size, err := converter.StringToByteSize(text, "us", common.GameVersionFFX)
		if err != nil {
			t.Fatalf("StringToByteSize(%q): %v", text, err)
		}
		if size != want.Len() {
			t.Fatalf("size de %q: got %d want %d", text, size, want.Len())
		}
	}
}

// Configuração inválida aborta (como StringToByteList); rune desconhecido
// não aborta (contagem segue igual a FillByteList).
func TestStringToByteSizeErrors(t *testing.T) {
	restore := seedTestMaps(t)
	defer restore()

	if _, err := converter.StringToByteSize("A", "xx", common.GameVersionFFX); err == nil {
		t.Fatal("charset desconhecido deve abortar")
	}
	size, err := converter.StringToByteSize("A€B", "us", common.GameVersionFFX)
	if err != nil {
		t.Fatalf("rune desconhecido não deve abortar: %v", err)
	}
	// 'A'→0x50, '€' ignorado, 'B'→0x51, + terminador.
	if size != 3 {
		t.Fatalf("size com rune ignorado: got %d want 3", size)
	}
}
