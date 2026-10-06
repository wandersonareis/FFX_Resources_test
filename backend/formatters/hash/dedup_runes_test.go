package hash

import (
	"testing"
	"unicode/utf8"
)

// Partículas em inglês que se repetem muito nos eventos mas podem ter
// tradução diferente a cada contexto — NUNCA podem virar ref de dedup.
// Regressão: "Whoa!" (5 runes) era dedupado em genk0100/genk1000.
func TestIsDedupEligibleExcludesParticles(t *testing.T) {
	particles := []string{
		"ok", "OK", "Hey!", "Huh?", "Wow!", "Yeah", "Whoa!",
		"What!", "What?!", "Yeah!", "Yeah!!", "Whoa!!", "Whoa!!!",
		"No way!", "Oh no!", "Ahh!", "Oww!", "Tidus!", "Seymour!",
		"Let's go!", "Thank you.", "Thank you!", "I'm sorry.",
	}

	for _, p := range particles {
		if rc := utf8.RuneCountInString(p); rc > MinDedupRunes {
			t.Fatalf("fixture desatualizada: %q tem %d runes > MinDedupRunes (%d) — mova-a para a lista de frases dedupáveis", p, rc, MinDedupRunes)
		}
		if IsDedupEligible(p) {
			t.Errorf("partícula %q (%d runes) não deve ser elegível para dedup", p, utf8.RuneCountInString(p))
		}
	}
}

// Frases completas (11+ runes) seguem dedupáveis: repetição com tradução
// tipicamente única é a carga real do dedup.
func TestIsDedupEligibleAcceptsPhrases(t *testing.T) {
	phrases := []string{
		"Are you sure?",
		"Thank you very much.",
		"We're counting on you.",
		"Texto compartilhado entre entradas",
	}

	for _, s := range phrases {
		if rc := utf8.RuneCountInString(s); rc <= MinDedupRunes {
			t.Fatalf("fixture desatualizada: %q tem %d runes ≤ MinDedupRunes (%d) — mova-a para a lista de partículas", s, rc, MinDedupRunes)
		}
		if !IsDedupEligible(s) {
			t.Errorf("frase %q (%d runes) deveria ser elegível para dedup", s, utf8.RuneCountInString(s))
		}
	}
}

// O texto de exatamente MinDedupRunes runes é o limite: inelegível.
// A comparação é estrita (>) — nunca >=.
func TestIsDedupEligibleBoundary(t *testing.T) {
	if got := utf8.RuneCountInString("Thank you."); got != MinDedupRunes {
		t.Fatalf("fixture de boundary desatualizada: %q tem %d runes, esperado %d", "Thank you.", got, MinDedupRunes)
	}
	if IsDedupEligible("Thank you.") {
		t.Errorf("texto de exatamente %d runes deve ser INELEGÍVEL (regra estrita >)", MinDedupRunes)
	}
}

// Partículas embrulhadas em tags de borda (falante/newline/formatação) não
// podem ser elegíveis: as tags das pontas são ignoradas e a régua cai sobre
// o miolo visível. Regressão do genk0100 do FFX: "{PC:00:Tidus}{TEXT_NEWLINE}
// Whoa!" (~30 runes crus) era dedupado como se fosse texto longo.
func TestIsDedupEligibleIgnoresEdgeTags(t *testing.T) {
	cases := []struct {
		name string
		text string
		core string
		want bool
	}{
		{
			name: "falante + newline + partícula (genk0100)",
			text: "{PC:00:Tidus}{TEXT_NEWLINE}Whoa!",
			core: "Whoa!",
			want: false,
		},
		{
			name: "qualquer número de tags na frente, intercaladas com espaços",
			text: "{TEXT_ITALIC} \n {TEXT_NEWLINE} {PC:01:Yuna}Whoa!",
			core: "Whoa!",
			want: false,
		},
		{
			name: "tag no fim também é ignorada",
			text: "Thanks!{PAUSE}",
			core: "Thanks!",
			want: false,
		},
		{
			name: "só tags e espaços: miolo vazio",
			text: "{TEXT_NEWLINE}",
			core: "",
			want: false,
		},
		{
			name: "frase longa com falante continua elegível",
			text: "{PC:00:Yuna}Are you sure about this?",
			core: "Are you sure about this?",
			want: true,
		},
		{
			name: "tag no meio do texto permanece no miolo (frase de 2 linhas)",
			text: "Foo{TEXT_NEWLINE}Bar",
			core: "Foo{TEXT_NEWLINE}Bar",
			want: true,
		},
		{
			// Espelha tagCounts: o varredor casa '{' com o primeiro '}'
			// seguinte — mesmo que o '{' seja literal sem fechamento próprio.
			// Caso patológico, documentado para fixar a semântica.
			name: "'{' literal seguido de tag: o '}' da tag fecha o '{' literal",
			text: "{PC:00:Tidus{TEXT_NEWLINE}Whoa!",
			core: "Whoa!",
			want: false,
		},
		{
			name: "'}' literal no fim não é tag",
			text: "Very long phrase here}}",
			core: "Very long phrase here}}",
			want: true,
		},
	}

	for _, c := range cases {
		if got := VisibleCore(c.text); got != c.core {
			t.Errorf("%s: VisibleCore(%q) = %q, esperado %q", c.name, c.text, got, c.core)
			continue
		}
		if got := IsDedupEligible(c.text); got != c.want {
			t.Errorf("%s: IsDedupEligible(%q) = %v, esperado %v (miolo %q)", c.name, c.text, got, c.want, c.core)
		}
	}
}
