package main

import (
	"fmt"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/services"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// TextService reúne tudo que mexe em texto: índice de entradas, coleções,
// busca, import/export JSON e .strings, catálogo de tags e idiomas.
type TextService struct {
	meta *services.MetadataService
}

// GetMetadata expõe a metadata nova (key, row_count, id, is_dir) ao frontend.
// Aceita metadata.key (ffx/...), id de collection (azit0000, command,
// chunk_00) ou caminho em disco. Gera models.Metadata no wailsjs.
func (s *TextService) GetMetadata(query string) (dto.Metadata, error) {
	if s.meta == nil {
		return dto.Metadata{}, fmt.Errorf("metadata service not initialized")
	}
	return s.meta.GetMetadata(query)
}

// ExportEntry monta o DTO da entrada e escreve os artefatos JSON e .strings
// em mods/edits. Devolve os caminhos escritos. langs nil/vazio = todos.
func (s *TextService) ExportEntry(kind, id string, version common.GameVersion, langs []string) ([]string, error) {
	if s.meta == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return s.meta.ExportEntry(kind, version, id, langs)
}

// ImportEntry lê o artefato JSON padrão da entrada em mods/edits e aplica o
// DTO de volta no binário. Devolve o caminho lido.
func (s *TextService) ImportEntry(kind, id string, version common.GameVersion) ([]string, error) {
	if s.meta == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return s.meta.ImportEntry(kind, version, id)
}

// ApplyTextCollection aplica um lote de entradas editadas (DTO) de volta no
// binário e persiste: é o "Salvar" do editor (todas as edições, não só a
// entrada aberta).
func (s *TextService) ApplyTextCollection(kind string, version common.GameVersion, c dto.Collection) error {
	if s.meta == nil {
		return fmt.Errorf("metadata service not initialized")
	}
	return s.meta.ApplyTextCollection(kind, version, c)
}

// ApplyVbfTextCollection aplica (no escopo da sessão do container) os edits
// vindos das tabelas abertas no navegador de .vbf: mesmo motor do "Salvar"
// de data/, com a propagação limitada às cópias abertas no clique. Grava em
// mods/ os binários tocados.
func (s *TextService) ApplyVbfTextCollection(vbfPath, kind string, version common.GameVersion, c dto.Collection) error {
	if s.meta == nil {
		return fmt.Errorf("metadata service not initialized")
	}
	return s.meta.ApplyVbfTextCollection(vbfPath, kind, version, c)
}

// ListTextEntries devolve o índice leve (id + key, sem rows) para montar
// sidebar/tree. kind: events, objects ou macro.
func (s *TextService) ListTextEntries(kind string, version common.GameVersion) ([]services.EntrySummary, error) {
	if s.meta == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return s.meta.ListEntries(kind, version)
}

// SearchTextEntries busca em data/ e mods/ no idioma us. O backend mantém o
// índice decodificado por versão; o resultado traz arquivos (com snippet
// das rows) e os totais da base para o "N rows em M arquivos" do modal.
func (s *TextService) SearchTextEntries(version common.GameVersion, query string) (services.TextSearchResponse, error) {
	if s.meta == nil {
		return services.TextSearchResponse{}, fmt.Errorf("metadata service not initialized")
	}
	return s.meta.SearchText(version, query)
}

// WarmTextSearch pré-constrói o índice de busca da versão em background
// (a aba chama depois que a árvore carrega): a primeira consulta do modal
// não paga a construção síncrona.
func (s *TextService) WarmTextSearch(version common.GameVersion) {
	if s.meta == nil {
		return
	}
	s.meta.WarmTextSearch(version)
}

// GetTextEntry devolve uma entrada completa (metadata + rows) por demanda,
// direto da memória — sem exportar para disco.
func (s *TextService) GetTextEntry(kind, id string, version common.GameVersion) (dto.FileEntry, error) {
	if s.meta == nil {
		return dto.FileEntry{}, fmt.Errorf("metadata service not initialized")
	}
	return s.meta.GetEntry(kind, id, version)
}

// GetTextCollection monta o DTO completo em memória (bulk; ids vazio = tudo).
// ids aceita ids ou keys (metadata.key/caminho resolvidos para id).
func (s *TextService) GetTextCollection(kind string, version common.GameVersion, ids []string) (dto.Collection, error) {
	if s.meta == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return s.meta.GetCollection(kind, version, ids)
}

// RevealEntryFile abre o explorador com o arquivo da entrada selecionado
// ("Abrir até o arquivo"), resolvido mods-first como o app lê.
func (s *TextService) RevealEntryFile(kind, id string, version common.GameVersion) error {
	if s.meta == nil {
		return fmt.Errorf("metadata service not initialized")
	}
	return s.meta.RevealEntryFile(kind, id, version)
}

// GetTagCatalog devolve o catálogo de tags nomeadas (PC/MCR/BUTTON/ICON) da
// versão para o autocomplete do editor. Os valores são ordenados pelo uso
// (mods/edits/tag_frequency_<v>.json), quando o ranking existe.
func (s *TextService) GetTagCatalog(version common.GameVersion) (services.TagCatalog, error) {
	if s.meta == nil {
		return services.TagCatalog{}, fmt.Errorf("metadata service not initialized")
	}
	return services.BuildTagCatalog(version)
}

// ListLanguages devolve os idiomas disponíveis em formato chave/valor
// (Code para arquivos, Name para exibição no picker).
func (s *TextService) ListLanguages() []common.Language {
	if s.meta == nil {
		return common.AvailableLanguages()
	}
	return s.meta.ListLanguages()
}

// ExportStrings escreve arquivos .strings ao lado dos .json.
// kind: events, objects ou macro. ids vazio = tudo (aceita ids ou keys);
// langs nil/vazio = todos os idiomas.
func (s *TextService) ExportStrings(kind string, version common.GameVersion, ids, langs []string) ([]string, error) {
	if s.meta == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return s.meta.ExportStrings(kind, version, ids, langs)
}

// ExportJSON escreve os artefatos JSON do lote (ids vazio = tudo).
// Caminho no padrão dos formatters (mods/edits) — o mesmo lido na importação.
func (s *TextService) ExportJSON(kind string, version common.GameVersion, ids, langs []string) ([]string, error) {
	if s.meta == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return s.meta.ExportJSON(kind, version, ids, langs)
}

// SelectImportFile abre o seletor nativo de arquivo para importar
// (.json / .strings). Devolve "" quando o usuário cancela.
func (s *TextService) SelectImportFile() string {
	return pickFile("Selecionar arquivo para importar", "",
		application.FileFilter{DisplayName: "Arquivos de importação (*.json; *.strings)", Pattern: "*.json;*.strings"},
		application.FileFilter{DisplayName: "JSON (*.json)", Pattern: "*.json"},
		application.FileFilter{DisplayName: "Strings (*.strings)", Pattern: "*.strings"},
		application.FileFilter{DisplayName: "Todos os arquivos (*.*)", Pattern: "*.*"},
	)
}

// PreviewImport parseia/valida o arquivo e devolve o resumo para o modal
// de confirmação (sem aplicar nada).
func (s *TextService) PreviewImport(path string, version common.GameVersion) (dto.ImportSummary, error) {
	if s.meta == nil {
		return dto.ImportSummary{}, fmt.Errorf("metadata service not initialized")
	}
	return s.meta.PreviewImport(path, version)
}

// ImportFile aplica o arquivo validado no binário (só o texto 'us').
// Devolve quantos textos em inglês mudaram em relação à store.
func (s *TextService) ImportFile(path string, version common.GameVersion) (int, error) {
	if s.meta == nil {
		return 0, fmt.Errorf("metadata service not initialized")
	}
	return s.meta.ImportFile(path, version)
}

// CountChangedTexts confere quantas rows têm 'us' diferente entre duas
// coleções (utilitário de conferência do frontend).
func (s *TextService) CountChangedTexts(imported, store dto.Collection) int {
	if s.meta == nil {
		return 0
	}
	return s.meta.CountChangedTexts(imported, store)
}

// PreviewImportTranslated varre a raiz de textos traduzidos
// (<TranslateLocation>/text) e devolve o resumo do lote SEM aplicar — é o
// que alimenta o diálogo de confirmação do Importar em lote.
func (s *TextService) PreviewImportTranslated(version common.GameVersion, kinds []string) (dto.ImportBatch, error) {
	if s.meta == nil {
		return dto.ImportBatch{}, fmt.Errorf("metadata service not initialized")
	}
	return s.meta.PreviewImportTranslated(version, kinds)
}

// ImportTranslated reexecuta o lote e aplica arquivo a arquivo, gravando os
// binários em mods/. Arquivos com erro são pulados um a um; as falhas voltam
// agregadas no resumo.
func (s *TextService) ImportTranslated(version common.GameVersion, kinds []string) (dto.ImportBatch, error) {
	if s.meta == nil {
		return dto.ImportBatch{}, fmt.Errorf("metadata service not initialized")
	}
	return s.meta.ImportTranslated(version, kinds)
}
