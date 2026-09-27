package builders

import (
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/hash"
)

// refBare é a guarda anti-falso-positivo das refs "$hash": só é ref quando
// o valor casa com o hash PRÓPRIO da row (literal que comece com `$` passa
// reto — o texto vence o hash).
func TestRefBareOnlyMatchesOwnHash(t *testing.T) {
	text := map[string]string{common.DefaultLocalization: "Texto compartilhado entre entradas"}
	row := dto.TextRow{Index: 0, Hash: hash.Texts(text), Text: text}

	if bare, ok := refBare(row, "$"+row.Hash[common.DefaultLocalization]); !ok || bare == "" {
		t.Fatal("ref válida deveria resolver")
	}
	if _, ok := refBare(row, "$deadbeefdeadbeef"); ok {
		t.Fatal("valor $ que não casa com o hash próprio não é ref")
	}
	if _, ok := refBare(row, "Texto compartilhado entre entradas"); ok {
		t.Fatal("literal pura não deveria resolver como ref")
	}
}
