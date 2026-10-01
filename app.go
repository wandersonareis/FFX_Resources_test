package main

import (
	"context"
	"ffxresources/backend/common"
	coreprogress "ffxresources/backend/core/progress"
	"ffxresources/backend/dto"
	"ffxresources/backend/interactions"
	"ffxresources/backend/loggingService"
	"ffxresources/backend/services"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	noticationService services.INotificationService

	progressService services.IProgressService

	statusBarService services.IStatusBarService

	MetadataService *services.MetadataService

	ctx          context.Context
	unsavedMu    sync.Mutex
	unsavedEdits bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	notifier := services.NewEventNotifier(context.Background())
	return &App{
		MetadataService: services.NewMetadataService(notifier),
	}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// O arquivo de log nasce aqui: um por início de app, ao lado do
	// executável (logs/ffx-<início>.log) — antes de qualquer mensagem.
	// A linha abaixo garante conteúdo desde o primeiro instante (e deixa
	// o caminho do log à vista no próprio arquivo).
	loggingService.Init()
	loggingService.Info("FFX Resources iniciado — log em %s", loggingService.LogDir())

	// Perform your setup here
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)

			l := loggingService.Get()
			l.Fatal().Caller(2).Err(err.(error)).Msg("panic occurred")
		}
	}()

	// Initialize services
	a.initServices(ctx)

	a.ctx = ctx

	interactions.NewInteractionWithCtx(ctx)
}

// domReady is called after front-end resources have been loaded
func (a App) domReady(ctx context.Context) {
	// Add your action here
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)

			l := loggingService.Get()
			l.Fatal().Caller(2).Err(err.(error)).Msg("panic occurred")
		}
	}()

	EventsOnStartup(ctx)

	EventsOnSaveConfig(ctx)
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	interactions.NewInteractionService().FFXAppConfig().ToJson()

	if !a.hasUnsavedEdits() {
		return false
	}

	// Há edições não salvas: o texto mora no frontend, então "Salvar" emite
	// o evento SaveRequested (o frontend salva e chama QuitApp para fechar).
	answer, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:    runtime.QuestionDialog,
		Title:   "Alterações não salvas",
		Message: "Há textos editados não salvos. Deseja salvar antes de fechar?",
		Buttons: []string{"&Salvar", "&Descartar", "&Cancelar"},
	})
	if err != nil {
		return true
	}

	fmt.Println("Close answer:", answer)
	switch answer {
	case "Salvar":
		runtime.EventsEmit(ctx, "SaveRequested")
		return true
	case "Descartar":
		a.SetUnsavedEdits(false)
		return false
	default:
		return true
	}
}

// hasUnsavedEdits consulta a flag de edições pendentes (thread-safe).
func (a *App) hasUnsavedEdits() bool {
	a.unsavedMu.Lock()
	defer a.unsavedMu.Unlock()
	return a.unsavedEdits
}

// SetUnsavedEdits recebe do frontend o estado de edições não salvas.
func (a *App) SetUnsavedEdits(dirty bool) {
	a.unsavedMu.Lock()
	a.unsavedEdits = dirty
	a.unsavedMu.Unlock()
}

// QuitApp encerra a aplicação (usado após salvar no fluxo SaveRequested).
func (a *App) QuitApp() {
	if a.ctx != nil {
		runtime.Quit(a.ctx)
		return
	}
	runtime.Quit(interactions.NewInteractionService().Ctx)
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform your teardown here
}

func (a *App) initServices(ctx context.Context) {
	notification := services.NewEventNotifier(ctx)

	a.noticationService = notification

	// Ponte de progresso: operações longas (carga de árvore, export) publicam
	// contadores pelo canal neutro core/progress e esta ponte emite os
	// eventos do wails (Progress/ShowProgress) para a barra do frontend.
	progressSvc := services.NewProgressService(ctx, notification)
	a.progressService = progressSvc
	coreprogress.Set(progressSvc)

	// Ponte da barra de status: desalinhamentos (original vs tradução)
	// publicados pelos services chegam à barra de rodapé do frontend.
	statusSvc := services.NewStatusBarService(ctx)
	a.statusBarService = statusSvc
	services.SetStatusBarService(statusSvc)

	// Initialize services
	a.MetadataService = services.NewMetadataService(notification)
}

func (a *App) ReadFileAsString(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	//fmt.Println(string(content))
	return string(content)
}

func (a *App) WriteTextFile(file string, text string) {
	err := os.WriteFile(file, []byte(text), 0644)
	if err != nil {
		a.noticationService.NotifyError(err)
	}
}

func (a *App) SelectDirectory(title string) string {
	selection, err := runtime.OpenDirectoryDialog(interactions.NewInteractionService().Ctx, runtime.OpenDialogOptions{
		Title:            title,
		DefaultDirectory: common.GetExecDir(),
	})

	if err != nil {
		return ""
	}
	return selection
}

// GetGameFilesLocation devolve o diretório dos arquivos do jogo
// (config ou default <exec>/data). Consulta sob demanda para o
// diálogo de config não depender do evento de startup.
func (a *App) GetGameFilesLocation() string {
	return interactions.NewInteractionService().GameLocation.GetTargetDirectory()
}

// GetTranslateLocation devolve o diretório de tradução usado no
// reimport (<game>/mods/translated por padrão). Consulta sob demanda.
func (a *App) GetTranslateLocation() string {
	return interactions.NewInteractionService().TranslateLocation.GetTargetDirectory()
}

// GetEnableMods devolve a preferência mods-first da carga de binários
// (config EnableMods). Consulta sob demanda para o checkbox de config.
func (a *App) GetEnableMods() bool {
	return interactions.NewInteractionService().FFXAppConfig().GetEnableMods()
}

// SetEnableMods persiste e aplica o toggle mods-first imediatamente.
func (a *App) SetEnableMods(enabled bool) error {
	interactions.NewInteractionService().FFXAppConfig().SetEnableMods(enabled)
	return nil
}

// GetMetadata expõe a metadata nova (key, row_count, id, is_dir) ao frontend.
// Aceita metadata.key (ffx/...), id de collection (azit0000, command,
// chunk_00) ou caminho em disco. Gera models.Metadata no wailsjs.
func (a *App) GetMetadata(query string) (dto.Metadata, error) {
	if a.MetadataService == nil {
		return dto.Metadata{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.GetMetadata(query)
}

// ExportEntry monta o DTO da entrada e escreve os artefatos JSON e .strings
// em mods/edits. Devolve os caminhos escritos. langs nil/vazio = todos.
func (a *App) ExportEntry(kind, id string, version common.GameVersion, langs []string) ([]string, error) {
	if a.MetadataService == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ExportEntry(kind, version, id, langs)
}

// ImportEntry lê o artefato JSON padrão da entrada em mods/edits e aplica o
// DTO de volta no binário. Devolve o caminho lido.
func (a *App) ImportEntry(kind, id string, version common.GameVersion) ([]string, error) {
	if a.MetadataService == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ImportEntry(kind, version, id)
}

// ApplyTextCollection aplica um lote de entradas editadas (DTO) de volta no
// binário e persiste: é o "Salvar" do editor (todas as edições, não só a
// entrada aberta).
func (a *App) ApplyTextCollection(kind string, version common.GameVersion, c dto.Collection) error {
	if a.MetadataService == nil {
		return fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ApplyTextCollection(kind, version, c)
}

// ListTextEntries devolve o índice leve (id + key, sem rows) para montar
// sidebar/tree. kind: events, objects ou macro.
func (a *App) ListTextEntries(kind string, version common.GameVersion) ([]services.EntrySummary, error) {
	if a.MetadataService == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ListEntries(kind, version)
}

// GetTextEntry devolve uma entrada completa (metadata + rows) por demanda,
// direto da memória — sem exportar para disco.
func (a *App) GetTextEntry(kind, id string, version common.GameVersion) (dto.FileEntry, error) {
	if a.MetadataService == nil {
		return dto.FileEntry{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.GetEntry(kind, id, version)
}

// GetTextCollection monta o DTO completo em memória (bulk; ids vazio = tudo).
// ids aceita ids ou keys (metadata.key/caminho resolvidos para id).
func (a *App) GetTextCollection(kind string, version common.GameVersion, ids []string) (dto.Collection, error) {
	if a.MetadataService == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.GetCollection(kind, version, ids)
}

// GetTagCatalog devolve o catálogo de tags nomeadas (PC/MCR/BUTTON/ICON) da
// versão para o autocomplete do editor. Os valores são ordenados pelo uso
// (mods/edits/tag_frequency_<v>.json), quando o ranking existe.
func (a *App) GetTagCatalog(version common.GameVersion) (services.TagCatalog, error) {
	if a.MetadataService == nil {
		return services.TagCatalog{}, fmt.Errorf("metadata service not initialized")
	}
	return services.BuildTagCatalog(version)
}

// ListLanguages devolve os idiomas disponíveis em formato chave/valor
// (Code para arquivos, Name para exibição no picker).
func (a *App) ListLanguages() []common.Language {
	if a.MetadataService == nil {
		return common.AvailableLanguages()
	}
	return a.MetadataService.ListLanguages()
}

// WriteLog persiste um log enviado pelo frontend no MESMO sistema do
// backend: console colorido pelo nível + arquivo em JSON
// (logs/ffx-<início>.log, ao lado do executável, com source=frontend).
// Nível: debug | info | warn | error.
func (a *App) WriteLog(level, message string, fields map[string]any) error {
	return loggingService.FromFrontend(level, message, fields)
}

// PreloadVersions aquece as versões indicadas em background (uma goroutine
// worker, versões uma a uma). Chamado pelo frontend após a carga da aba
// visualizada concluir — trocar de aba cai no caminho rápido.
// versões inválidas são ignoradas com aviso no log de arquivo.
func (a *App) PreloadVersions(versions []string) {
	if a.MetadataService == nil {
		return
	}
	valid := make([]common.GameVersion, 0, len(versions))
	for _, raw := range versions {
		v, err := common.ParseGameVersionStrict(raw)
		if err != nil {
			loggingService.Warn("pré-carga: versão inválida %q", raw)
			continue
		}
		valid = append(valid, v)
	}
	a.MetadataService.PreloadVersions(valid)
}

// ExportStrings escreve arquivos .strings ao lado dos .json.
// kind: events, objects ou macro. ids vazio = tudo (aceita ids ou keys);
// langs nil/vazio = todos os idiomas.
func (a *App) ExportStrings(kind string, version common.GameVersion, ids, langs []string) ([]string, error) {
	if a.MetadataService == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ExportStrings(kind, version, ids, langs)
}

// ExportJSON escreve os artefatos JSON do lote (ids vazio = tudo).
// Caminho no padrão dos formatters (mods/edits) — o mesmo lido na importação.
func (a *App) ExportJSON(kind string, version common.GameVersion, ids, langs []string) ([]string, error) {
	if a.MetadataService == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ExportJSON(kind, version, ids, langs)
}

// SelectImportFile abre o seletor nativo de arquivo para importar
// (.json / .strings). Devolve "" quando o usuário cancela.
func (a *App) SelectImportFile() string {
	selection, err := runtime.OpenFileDialog(interactions.NewInteractionService().Ctx, runtime.OpenDialogOptions{
		Title: "Selecionar arquivo para importar",
		Filters: []runtime.FileFilter{
			{DisplayName: "Arquivos de importação (*.json; *.strings)", Pattern: "*.json;*.strings"},
			{DisplayName: "JSON (*.json)", Pattern: "*.json"},
			{DisplayName: "Strings (*.strings)", Pattern: "*.strings"},
			{DisplayName: "Todos os arquivos (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return ""
	}
	return selection
}

// PreviewImport parseia/valida o arquivo e devolve o resumo para o modal
// de confirmação (sem aplicar nada).
func (a *App) PreviewImport(path string, version common.GameVersion) (dto.ImportSummary, error) {
	if a.MetadataService == nil {
		return dto.ImportSummary{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.PreviewImport(path, version)
}

// ImportFile aplica o arquivo validado no binário (só o texto 'us').
// Devolve quantos textos em inglês mudaram em relação à store.
func (a *App) ImportFile(path string, version common.GameVersion) (int, error) {
	if a.MetadataService == nil {
		return 0, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ImportFile(path, version)
}

// CountChangedTexts confere quantas rows têm 'us' diferente entre duas
// coleções (utilitário de conferência do frontend).
func (a *App) CountChangedTexts(imported, store dto.Collection) int {
	if a.MetadataService == nil {
		return 0
	}
	return a.MetadataService.CountChangedTexts(imported, store)
}
