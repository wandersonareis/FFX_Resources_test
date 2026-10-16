package services

import (
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
)

// TestObjectsWeaponNamesCarregaNoFluxoChunkmap: w_name.bin é registrado
// pelos weaponnames (Register via init de services) e carrega pelo MESMO
// caminho do frontend — buildObjectsCollection. Sem o registro, o arquivo
// era pulado com "skip objects" e o painel vinha vazio.
func TestObjectsWeaponNamesCarregaNoFluxoChunkmap(t *testing.T) {
	seedFFXHelpRoot(t)
	_ = ensureVersionReady(common.GameVersionFFX)

	const id = "w_name"
	s := NewMetadataService(nil)
	c, err := s.buildObjectsCollection(common.GameVersionFFX, []string{id}, nil)
	if err != nil {
		t.Fatalf("buildObjectsCollection: %v", err)
	}
	entry, ok := c[id]
	if !ok {
		t.Fatal("w_name ausente na coleção — schema chunkmap não registrado?")
	}
	if len(entry.Rows) == 0 {
		t.Fatalf("w_name sem rows (metadata=%+v)", entry.Metadata)
	}
	// Rows rotuladas pelo dono: "Tidus.Standard", "Yuna.Standard", ...
	var names []string
	for _, r := range entry.Rows {
		names = append(names, r.Name)
	}
	t.Logf("w_name: %d rows, ex: %v", len(entry.Rows), names[:min(4, len(names))])
	if entry.Metadata.RowCount != len(entry.Rows) {
		t.Fatalf("RowCount=%d, esperado %d", entry.Metadata.RowCount, len(entry.Rows))
	}
	_ = dto.RowKey(entry.Rows[0])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
