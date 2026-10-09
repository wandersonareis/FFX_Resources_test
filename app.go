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
	"path/filepath"
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
	// Os .vbf abertos seguram handle de leitura: solta antes de sair.
	services.CloseVbfArchives()
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

// GetGameExeLocation devolve o executável do jogo configurado — é dele que
// sai a pasta onde os .vbf do container moram (FFX_Data.vbf, FFX2_Data.vbf).
// "" = o usuário ainda não escolheu (sem raízes VBF na sidebar).
func (a *App) GetGameExeLocation() string {
	return interactions.NewInteractionService().FFXAppConfig().GetGameExeLocation()
}

// SetGameExeLocation persiste o caminho do executável do jogo. Os containers
// abertos são fechados: a pasta de busca mudou e os handles antigos não são
// mais alcançáveis (a próxima ListVbfRoots reabre os da nova pasta).
func (a *App) SetGameExeLocation(path string) {
	interactions.NewInteractionService().FFXAppConfig().SetGameExeLocation(path)
	services.CloseVbfArchives()
}

// SelectGameExeFile abre o seletor nativo para escolher o executável do
// jogo. O filtro é LITERAL — só FFX.exe e FFX-2.exe, sem curinga *.exe —
// e a escolha é conferida de novo aqui: o diálogo do sistema aceita digitar
// qualquer nome no campo de arquivo. Devolve "" no cancelamento e numa
// escolha inválida.
func (a *App) SelectGameExeFile() string {
	selection, err := runtime.OpenFileDialog(interactions.NewInteractionService().Ctx, runtime.OpenDialogOptions{
		Title: "Selecionar o executável do jogo (FFX.exe ou FFX-2.exe)",
		Filters: []runtime.FileFilter{
			{DisplayName: "Executável do jogo (FFX.exe; FFX-2.exe)", Pattern: "FFX.exe;FFX-2.exe"},
		},
	})
	if err != nil {
		return ""
	}
	if selection == "" {
		return ""
	}
	normalized, ok := interactions.NormalizeGameExe(selection)
	if !ok {
		// Só o caminho vai para o diagnóstico — nunca conteúdo de arquivo.
		loggingService.DiagWarn("vbf",
			"executável fora de FFX.exe/FFX-2.exe recusado",
			map[string]any{"path": selection})
		if a.noticationService != nil {
			a.noticationService.NotifyWarn(
				"Escolha um dos executáveis do jogo: FFX.exe ou FFX-2.exe.")
		}
		return ""
	}
	return normalized
}

// ---- navegador de .vbf (SOMENTE LEITURA) ---------------------------------
//
// O container nunca é materializado inteiro em memória: a árvore sai do
// índice e a extração selecionada decodifica/grava um arquivo por vez. O
// .vbf nunca é escrito; export de texto vai para mods/edits, e extração de
// binários vai para data/ ou para a pasta escolhida pelo usuário.

// ListVbfRoots devolve os .vbf encontrados ao lado do executável do jogo
// (raízes da sidebar), com o total de arquivos do índice. Lista vazia =
// executável não configurado.
func (a *App) ListVbfRoots() ([]dto.VbfRoot, error) {
	if a.MetadataService == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ListVbfRoots()
}

// ListVbfDir devolve os filhos imediatos de um diretório do .vbf — diretórios
// primeiro, depois arquivos, cada um com kind/id quando o app sabe servi-lo.
func (a *App) ListVbfDir(vbfPath, dir string) ([]dto.VbfNode, error) {
	if a.MetadataService == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ListVbfDir(vbfPath, dir)
}

// ListVbfMacroChunks devolve os filhos virtuais de macrodic.dcp (chunk_00,
// chunk_01, …): o dicionário é um arquivo por localização e o app o trata
// como um grupo.
func (a *App) ListVbfMacroChunks(vbfPath, macroPath string) ([]dto.VbfNode, error) {
	if a.MetadataService == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ListVbfMacroChunks(vbfPath, macroPath)
}

// GetVbfTextEntry decodifica um arquivo de TEXTO do .vbf e devolve a tabela
// (rows + coluna Original). id é exigido só para macro (o chunk).
func (a *App) GetVbfTextEntry(vbfPath, innerPath, id string) (dto.FileEntry, error) {
	if a.MetadataService == nil {
		return dto.FileEntry{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.GetVbfTextEntry(vbfPath, innerPath, id)
}

// GetVbfImageEntry decodifica uma textura (.dds.phyre) do .vbf e devolve a
// pré-visualização PNG em data URL.
func (a *App) GetVbfImageEntry(vbfPath, innerPath string) (dto.ImageEntry, error) {
	if a.MetadataService == nil {
		return dto.ImageEntry{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.GetVbfImageEntry(vbfPath, innerPath)
}

func (a *App) ImportVbfImage(vbfPath, innerPath, ddsPath string) error {
	if a.MetadataService == nil {
		return fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ImportVbfImage(vbfPath, innerPath, ddsPath)
}

func (a *App) SaveVbfImage(vbfPath, innerPath, format, destPath string) error {
	if a.MetadataService == nil {
		return fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.SaveVbfImage(vbfPath, innerPath, format, destPath)
}

func (a *App) ReplicateVbfImage(vbfPath, sourcePath string, targets []string) (dto.BatchResult, error) {
	if a.MetadataService == nil {
		return dto.BatchResult{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ReplicateVbfImage(vbfPath, sourcePath, targets)
}

// ---- extração/export da SELEÇÃO da árvore do .vbf --------------------------
//
// A seleção é uma lista de caminhos internos (arquivos e diretórios);
// diretório = toda a subárvore, expandida contra o índice no backend.
// Extrair preserva a estrutura de caminhos no destino (default data/ do
// jogo); exportar grava JSON/.strings em mods/edits, como o export de
// data/. O container continua somente leitura.

// PreviewVbfExtraction devolve o resumo da seleção (arquivos, bytes e
// quantos já existem no destino) para o diálogo de confirmação — destRoot
// vazio significa data/ do jogo.
func (a *App) PreviewVbfExtraction(vbfPath string, paths []string, destRoot string) (dto.VbfExtractPreview, error) {
	if a.MetadataService == nil {
		return dto.VbfExtractPreview{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.PreviewVbfExtraction(vbfPath, paths, destRoot)
}

// ExtractVbfSelection extrai os binários da seleção para destRoot (vazio =
// data/ do jogo), preservando a estrutura interna de caminhos.
func (a *App) ExtractVbfSelection(vbfPath string, paths []string, destRoot string) (dto.BatchResult, error) {
	if a.MetadataService == nil {
		return dto.BatchResult{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ExtractVbfSelection(vbfPath, paths, destRoot)
}

// ExtractVbfImagesSelection extrai DDS/PNG das imagens marcadas, preservando
// os caminhos internos do .vbf sob destRoot (vazio = mods/images).
func (a *App) ExtractVbfImagesSelection(vbfPath string, paths []string, destRoot string) (dto.BatchResult, error) {
	if a.MetadataService == nil {
		return dto.BatchResult{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ExtractVbfImagesSelection(vbfPath, paths, destRoot)
}

// SelectVbfExtractDir abre o seletor nativo de pasta para o "Extrair
// para…", começando em data/ do jogo. Devolve "" no cancelamento.
func (a *App) SelectVbfExtractDir() string {
	selection, err := runtime.OpenDirectoryDialog(interactions.NewInteractionService().Ctx, runtime.OpenDialogOptions{
		Title:            "Extrair para…",
		DefaultDirectory: filepath.Join(common.GameFilesRoot, common.DirData),
	})
	if err != nil {
		return ""
	}
	return selection
}

// SelectVbfImageExtractDir escolhe a RAIZ onde DDS/PNG extraídos do .vbf
// serão gravados — mods/images por padrão; abaixo dela o caminho interno do
// container é somado na gravação (por isso o picker abre na raiz e não já
// com o caminho interno, que duplicaria a raiz).
func (a *App) SelectVbfImageExtractDir() string {
	selection, err := runtime.OpenDirectoryDialog(interactions.NewInteractionService().Ctx, runtime.OpenDialogOptions{
		Title:            "Escolher destino das imagens extraídas",
		DefaultDirectory: filepath.Join(common.GameFilesRoot, common.ModsFolder, common.ModsImagesDir),
	})
	if err != nil {
		return ""
	}
	return selection
}

// ExportVbfSelection decodifica os kinds de TEXTO da seleção e grava os
// artefatos em mods/edits no formato pedido ("json" | "strings").
func (a *App) ExportVbfSelection(vbfPath, format string, paths []string, langs []string) ([]string, error) {
	if a.MetadataService == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ExportVbfSelection(vbfPath, format, paths, langs)
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

// ApplyVbfTextCollection aplica (no escopo da sessão do container) os edits
// vindos das tabelas abertas no navegador de .vbf: mesmo motor do "Salvar"
// de data/, com a propagação limitada às cópias abertas no clique. Grava em
// mods/ os binários tocados.
func (a *App) ApplyVbfTextCollection(vbfPath, kind string, version common.GameVersion, c dto.Collection) error {
	if a.MetadataService == nil {
		return fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ApplyVbfTextCollection(vbfPath, kind, version, c)
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

// GetImageEntry carrega a textura kind=images e devolve a pré-visualização
// (data URL PNG) com o DDS da fonte escolhida — ordem .dds em disco →
// .png em disco → decode do .dds.phyre em memória.
func (a *App) GetImageEntry(kind, id string, version common.GameVersion) (dto.ImageEntry, error) {
	if a.MetadataService == nil {
		return dto.ImageEntry{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.GetImage(kind, id, version)
}

// ExtractImage grava .dds e .png em mods/images (cópia de trabalho) e
// devolve os caminhos escritos.
func (a *App) ExtractImage(kind, id string, version common.GameVersion) ([]string, error) {
	if a.MetadataService == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ExtractImage(kind, id, version)
}

// ImportImage reempacota o .dds escolhido sobre o container pristine e grava
// o resultado em mods/ (a árvore data/ nunca é alterada).
func (a *App) ImportImage(kind, id, ddsPath string, version common.GameVersion) error {
	if a.MetadataService == nil {
		return fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ImportImage(kind, id, ddsPath, version)
}

// ImportImageGroup reempacota o mesmo .dds sobre a textura e as cópias
// idênticas dela (payload original igual — as réplicas da otimização do
// DVD). O lote é validado antes de gravar: alvo fora do grupo é recusado.
func (a *App) ImportImageGroup(kind, id, ddsPath string, targets []string, version common.GameVersion) (dto.ImageImportResult, error) {
	if a.MetadataService == nil {
		return dto.ImageImportResult{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ImportImageGroup(kind, id, ddsPath, targets, version)
}

// RefreshImageDuplicates descarta o índice de duplicatas em cache (rebuild
// na próxima consulta) — para edição externa em hex editor.
func (a *App) RefreshImageDuplicates(kind string, version common.GameVersion) error {
	if a.MetadataService == nil {
		return fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.RefreshImageDuplicates(kind, version)
}

// ImageDuplicates devolve só as cópias da textura (sem a imagem em base64):
// é o que o menu de contexto e os diálogos pedem para montar os alvos de
// extrair / replicar / deletar sobre o NÓ CLICADO, que pode não ser a entry
// selecionada.
func (a *App) ImageDuplicates(kind, id string, version common.GameVersion) (dto.ImageDuplicates, error) {
	if a.MetadataService == nil {
		return dto.ImageDuplicates{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ImageDuplicates(kind, id, version)
}

// ImageExists responde se a textura ainda existe (data/ OU mods/). O frontend
// usa na revalidação da seleção após reload (ex.: delete apagou o id — não
// re-carregar).
func (a *App) ImageExists(kind, id string, version common.GameVersion) (bool, error) {
	if a.MetadataService == nil {
		return false, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ImageExists(kind, id, version)
}

// ExtractImageGroup extrai .dds + .png de uma textura e das cópias
// escolhidas (o "só esta ou todas as cópias" do diálogo de extração).
func (a *App) ExtractImageGroup(kind, id string, targets []string, version common.GameVersion) (dto.BatchResult, error) {
	if a.MetadataService == nil {
		return dto.BatchResult{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ExtractImageGroup(kind, id, targets, version)
}

// ExtractImageSelection extrai somente as imagens explicitamente selecionadas.
func (a *App) ExtractImageSelection(kind string, ids []string, version common.GameVersion) (dto.BatchResult, error) {
	if a.MetadataService == nil {
		return dto.BatchResult{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ExtractImageSelection(kind, ids, version)
}

// ReplicateImage reempacota a imagem ABERTA em mods/ das cópias escolhidas
// — o dupe sem diálogo de arquivo: fonte é o próprio conteúdo do painel.
func (a *App) ReplicateImage(kind, id string, targets []string, version common.GameVersion) (dto.BatchResult, error) {
	if a.MetadataService == nil {
		return dto.BatchResult{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ReplicateImage(kind, id, targets, version)
}

// DeleteImages apaga a textura e das cópias confirmadas no escopo pedido
// ("data" | "mods" | "both"), sempre com os artefatos derivados. O escopo e
// as cópias vêm EXPLICITAMENTE do diálogo — nunca é cascata implícita.
func (a *App) DeleteImages(kind, id string, targets []string, scope string, version common.GameVersion) (dto.BatchResult, error) {
	if a.MetadataService == nil {
		return dto.BatchResult{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.DeleteImages(kind, id, targets, scope, version)
}

// DeleteImageSelection deleta as imagens selecionadas no mesmo escopo. A
// opção withCopies inclui, uma vez cada, as cópias de payload de cada seleção.
func (a *App) DeleteImageSelection(kind string, ids []string, withCopies bool, scope string, version common.GameVersion) (dto.BatchResult, error) {
	if a.MetadataService == nil {
		return dto.BatchResult{}, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.DeleteImageSelection(kind, ids, withCopies, scope, version)
}

// ImageSelectionCopies retorna, sem decodificar imagens, a união única das
// cópias adicionais dos ids selecionados.
func (a *App) ImageSelectionCopies(kind string, ids []string, version common.GameVersion) ([]string, error) {
	if a.MetadataService == nil {
		return nil, fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.ImageSelectionCopies(kind, ids, version)
}

// RevealEntryFile abre o explorador com o arquivo da entrada selecionado
// ("Abrir até o arquivo"), resolvido mods-first como o app lê.
func (a *App) RevealEntryFile(kind, id string, version common.GameVersion) error {
	if a.MetadataService == nil {
		return fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.RevealEntryFile(kind, id, version)
}

// SaveImage grava .dds ou .png no caminho escolhido pelo usuário (o
// "salvar em disco" do painel de imagem).
func (a *App) SaveImage(kind, id, format, destPath string, version common.GameVersion) error {
	if a.MetadataService == nil {
		return fmt.Errorf("metadata service not initialized")
	}
	return a.MetadataService.SaveImage(kind, id, format, destPath, version)
}

// SelectImageFile abre o seletor nativo para escolher um .dds a importar,
// começando no diretório de trabalho da textura (mods/images/<raiz>/<dir do
// id>) — o lugar onde o .dds extraído dela já mora. Só .dds: é o único
// formato aceito no importe (o repack usa os bytes crus — um PNG/PDV teria
// de ser re-encodado para DXT, com perda e sem mips).
// Devolve "" quando o usuário cancela.
func (a *App) SelectImageFile(id string, version common.GameVersion) string {
	selection, err := runtime.OpenFileDialog(interactions.NewInteractionService().Ctx, runtime.OpenDialogOptions{
		Title:            "Selecionar textura para importar",
		DefaultDirectory: services.ImageWorkDir(version, id),
		Filters: []runtime.FileFilter{
			{DisplayName: "DDS (*.dds)", Pattern: "*.dds"},
		},
	})
	if err != nil {
		return ""
	}
	return selection
}

// SelectImageSavePath abre o diálogo "Salvar como" para .dds/.png já no
// diretório de trabalho da textura, para o arquivo nunca sair solto em
// mods/images. Devolve "" quando o usuário cancela.
func (a *App) SelectImageSavePath(format, suggestedName, id string, version common.GameVersion) string {
	pattern := "*.dds"
	display := "DDS (*.dds)"
	if format == "png" {
		pattern, display = "*.png", "PNG (*.png)"
	}
	selection, err := runtime.SaveFileDialog(interactions.NewInteractionService().Ctx, runtime.SaveDialogOptions{
		Title:            "Salvar imagem",
		DefaultDirectory: services.ImageWorkDir(version, id),
		DefaultFilename:  suggestedName,
		Filters: []runtime.FileFilter{
			{DisplayName: display, Pattern: pattern},
			{DisplayName: "Todos os arquivos (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return ""
	}
	return selection
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
