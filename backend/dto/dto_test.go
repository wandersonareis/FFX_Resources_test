package dto_test

import (
	"testing"

	"ffxresources/backend/dto"
)

// TestSortRowsKeepsFieldOrderForEqualIndex garante que rows com o mesmo Index
// (campos de um mesmo objeto) mantêm a ordem de emissão do domínio, em vez de
// serem reordenadas alfabeticamente por Name.
func TestSortRowsKeepsFieldOrderForEqualIndex(t *testing.T) {
	rows := []dto.TextRow{
		{Index: 0, Name: "name"},
		{Index: 0, Name: "simplifiedName"},
		{Index: 0, Name: "description"},
		{Index: 0, Name: "simplifiedDescription"},
	}
	dto.SortRows(rows)

	want := []string{"name", "simplifiedName", "description", "simplifiedDescription"}
	for i, w := range want {
		if rows[i].Name != w {
			t.Fatalf("position %d: got %q, want %q (%+v)", i, rows[i].Name, w, rows)
		}
	}
}

// TestSortRowsOrdersByIndex garante a ordenação primária por Index.
func TestSortRowsOrdersByIndex(t *testing.T) {
	rows := []dto.TextRow{
		{Index: 2, Name: "name"},
		{Index: 1, Name: "name"},
		{Index: 1, Name: "description"},
		{Index: 0, Name: "name"},
	}
	dto.SortRows(rows)

	wantIndexes := []int{0, 1, 1, 2}
	wantNames := []string{"name", "name", "description", "name"}
	for i := range rows {
		if rows[i].Index != wantIndexes[i] || rows[i].Name != wantNames[i] {
			t.Fatalf("position %d: got (%d,%q), want (%d,%q)",
				i, rows[i].Index, rows[i].Name, wantIndexes[i], wantNames[i])
		}
	}
}
