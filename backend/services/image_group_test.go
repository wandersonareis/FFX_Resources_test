package services

import (
	"reflect"
	"strings"
	"testing"
)

// O lote de import é validado ANTES de gravar: o grupo vem do índice de
// duplicatas (payload original igual) e qualquer alvo fora dele derruba a
// operação inteira — nunca se escreve meio lote.
func TestGroupTargetsValidatesBeforeWriting(t *testing.T) {
	group := []string{"gamedata/ps3data/dup/a", "gamedata/ps3data/dup/b", "gamedata/ps3data/dup/c"}

	got, err := groupTargets("gamedata/ps3data/dup/a", group, []string{
		"gamedata/ps3data/dup/c",
		"gamedata/ps3data/dup/b",
		"gamedata/ps3data/dup/c", // repetido
		"  ",                     // vazio
		"gamedata/ps3data/dup/a", // o próprio id de novo
	})
	if err != nil {
		t.Fatalf("groupTargets: %v", err)
	}
	want := []string{
		"gamedata/ps3data/dup/a",
		"gamedata/ps3data/dup/c",
		"gamedata/ps3data/dup/b",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("lista = %v, esperado %v (id primeiro, sem repetições)", got, want)
	}

	// Sem alvos explícitos o lote é só a própria textura.
	only, err := groupTargets("gamedata/ps3data/dup/a", group, nil)
	if err != nil {
		t.Fatalf("groupTargets(sem alvos): %v", err)
	}
	if !reflect.DeepEqual(only, []string{"gamedata/ps3data/dup/a"}) {
		t.Errorf("lista = %v, esperado só o id", only)
	}
}

func TestGroupTargetsRefusesForeignCopy(t *testing.T) {
	group := []string{"gamedata/ps3data/dup/a", "gamedata/ps3data/dup/b"}

	// Alvo de OUTRO grupo: recusado, com o motivo no erro.
	if _, err := groupTargets("gamedata/ps3data/dup/a", group,
		[]string{"gamedata/ps3data/other/c"}); err == nil {
		t.Fatal("esperava recusa para alvo fora do grupo")
	} else if !strings.Contains(err.Error(), "não é cópia idêntica") {
		t.Errorf("erro sem o motivo: %v", err)
	}

	// id nem no próprio grupo = índice desatualizado: recusado também.
	if _, err := groupTargets("gamedata/ps3data/other/c", group, nil); err == nil {
		t.Fatal("esperava recusa para id fora do índice")
	}
}
