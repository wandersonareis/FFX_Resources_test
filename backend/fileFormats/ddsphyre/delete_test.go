package ddsphyre

import (
	"os"
	"path/filepath"
	"testing"

	"ffxresources/backend/common"
)

// Delete precisa cobrir os três escopos E os artefatos derivados: Resolve
// prefere mods/images, então apagar só o container deixaria um .dds
// órfão servindo imagem fantasma no lugar da textura apagada.
func TestDeleteRemovesContainerAndDerivedArtifacts(t *testing.T) {
	root, a, _, _, _ := withDupTree(t)
	dataPath := filepath.Join(root, RelPath(common.GameVersionFFX, a))
	modsPath := filepath.Join(root, common.ModsFolder, RelPath(common.GameVersionFFX, a))
	ddsPath, pngPath := ExportPaths(common.GameVersionFFX, a)

	seed := func(path string) {
		t.Helper()
		if err := common.EnsurePathExists(path); err != nil {
			t.Fatalf("criando %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte("conteudo"), 0o644); err != nil {
			t.Fatalf("gravando %s: %v", path, err)
		}
	}
	gone := func(path string) bool {
		_, err := os.Stat(path)
		return os.IsNotExist(err)
	}

	seed(modsPath)
	seed(ddsPath)
	seed(pngPath)

	// mods: some o overlay e os artefatos; o pristine fica intacto.
	removed, err := Delete(common.GameVersionFFX, a, DeleteMods)
	if err != nil {
		t.Fatalf("Delete(mods): %v", err)
	}
	if !removed {
		t.Error("removed = false com mods/ existente")
	}
	if !gone(modsPath) {
		t.Errorf("container de mods/ sobreviveu: %s", modsPath)
	}
	if !gone(ddsPath) || !gone(pngPath) {
		t.Error("artefato derivado ficou órfão depois do delete")
	}
	if _, err := os.Stat(dataPath); err != nil {
		t.Errorf("data/ foi junto sem pedir: %v", err)
	}

	// Segunda passada: nada a apagar NÃO é erro (idempotente).
	removed, err = Delete(common.GameVersionFFX, a, DeleteMods)
	if err != nil || removed {
		t.Errorf("idempotência: removed=%v err=%v (esperado false, nil)", removed, err)
	}

	// Escopo inválido: recusado antes de tocar em qualquer coisa.
	if _, err = Delete(common.GameVersionFFX, a, "qualquer"); err == nil {
		t.Error("escopo inválido aceito")
	}
	if _, err := os.Stat(dataPath); err != nil {
		t.Errorf("escopo inválido apagou data/: %v", err)
	}

	// data: sai o pristine (a textura some da árvore — regra 4).
	removed, err = Delete(common.GameVersionFFX, a, DeleteData)
	if err != nil {
		t.Fatalf("Delete(data): %v", err)
	}
	if !removed {
		t.Error("removed = false com data/ existente")
	}
	if !gone(dataPath) {
		t.Errorf("pristine sobreviveu: %s", dataPath)
	}
}

// both apaga nas duas árvores de uma vez (o escopo do diálogo "Ambos").
func TestDeleteBothClearsBothTrees(t *testing.T) {
	root, a, _, _, _ := withDupTree(t)
	dataPath := filepath.Join(root, RelPath(common.GameVersionFFX, a))
	modsPath := filepath.Join(root, common.ModsFolder, RelPath(common.GameVersionFFX, a))

	if err := common.EnsurePathExists(modsPath); err != nil {
		t.Fatalf("criando mods/: %v", err)
	}
	if err := os.WriteFile(modsPath, []byte("overlay"), 0o644); err != nil {
		t.Fatalf("gravando mods/: %v", err)
	}

	removed, err := Delete(common.GameVersionFFX, a, DeleteBoth)
	if err != nil {
		t.Fatalf("Delete(both): %v", err)
	}
	if !removed {
		t.Error("removed = false com arquivo em data/")
	}
	for _, path := range []string{dataPath, modsPath} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("sobrou %s (%v)", path, err)
		}
	}
}

// Id com traversal ou fora de gamedata é recusado antes de resolver caminho
// algum (é a trava dos bindings que recebem id da UI).
func TestDeleteRefusesInvalidID(t *testing.T) {
	withDupTree(t)
	if _, err := Delete(common.GameVersionFFX, "../../fora", DeleteBoth); err == nil {
		t.Error("id com traversal aceito")
	}
	if _, err := Delete(common.GameVersionFFX, "outros/dados/x", DeleteBoth); err == nil {
		t.Error("id fora de gamedata/ps3data aceito")
	}
}
