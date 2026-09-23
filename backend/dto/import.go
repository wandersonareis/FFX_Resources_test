package dto

// Este arquivo define o resumo da importação devolvido ao modal de
// confirmação. Erros de arquivo (ilegível, kind/versão divergentes etc.)
// voltam como error (toast no frontend); erros por entrada ou de capacidade
// ficam nestes slices — Errors não vazio desabilita o botão Importar.
// Todos os slices são inicializados não-nil no backend (nil → JSON null).

// ImportEntryInfo compara uma entrada do arquivo importado com a store.
type ImportEntryInfo struct {
	ID              string `json:"id"`
	Key             string `json:"key"`
	IndexCount      int    `json:"index_count"`
	StoreIndexCount int    `json:"store_index_count"`
	ChangedTexts    int    `json:"changed_texts"`
	Error           string `json:"error,omitempty"`
}

// ImportUsage reporta o uso pós-rebuild do limite uint16 de um binário
// (Used em bytes; Over bloqueia a importação).
type ImportUsage struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Used  int    `json:"used"`
	Limit int    `json:"limit"`
	Over  bool   `json:"over"`
}

// ImportSummary é o resumo exibido no modal antes de aplicar a importação.
type ImportSummary struct {
	Path         string            `json:"path"`
	Format       string            `json:"format"` // "json" | "strings"
	Kind         string            `json:"kind"`
	Version      string            `json:"version"`
	Languages    []string          `json:"languages"`
	EntryCount   int               `json:"entry_count"`
	TotalIndices int               `json:"total_indices"`
	ChangedTexts int               `json:"changed_texts"`
	Entries      []ImportEntryInfo `json:"entries"`
	Usages       []ImportUsage     `json:"usages"`
	Errors       []string          `json:"errors"`
}
