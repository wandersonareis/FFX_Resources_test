package main

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/loggingService"
	"ffxresources/backend/services"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ConfigService reúne a configuração do app: localizações do jogo, preferência
// mods-first e escolha do executável — tudo o que o diálogo de configuração
// toca. O equivalente vivia no App (o "god object" do v2).
type ConfigService struct {
	notifier services.INotificationService
}

func (c *ConfigService) SelectDirectory(title string) string {
	return pickDirectory(title, common.GetExecDir())
}

// GetGameFilesLocation devolve o diretório dos arquivos do jogo
// (config ou default <exec>/data). Consulta sob demanda para o
// diálogo de config não depender do evento de startup.
func (c *ConfigService) GetGameFilesLocation() string {
	return interactions.NewInteractionService().GameLocation.GetTargetDirectory()
}

// GetTranslateLocation devolve o diretório de tradução usado no
// reimport (<game>/mods/translated por padrão). Consulta sob demanda.
func (c *ConfigService) GetTranslateLocation() string {
	return interactions.NewInteractionService().TranslateLocation.GetTargetDirectory()
}

// GetEnableMods devolve a preferência mods-first da carga de binários
// (config EnableMods). Consulta sob demanda para o checkbox de config.
func (c *ConfigService) GetEnableMods() bool {
	return interactions.NewInteractionService().FFXAppConfig().GetEnableMods()
}

// SetEnableMods persiste e aplica o toggle mods-first imediatamente.
func (c *ConfigService) SetEnableMods(enabled bool) error {
	interactions.NewInteractionService().FFXAppConfig().SetEnableMods(enabled)
	return nil
}

// GetGameExeLocation devolve o executável do jogo configurado — é dele que
// sai a pasta onde os .vbf do container moram (FFX_Data.vbf, FFX2_Data.vbf).
// "" = o usuário ainda não escolheu (sem raízes VBF na sidebar).
func (c *ConfigService) GetGameExeLocation() string {
	return interactions.NewInteractionService().FFXAppConfig().GetGameExeLocation()
}

// SetGameExeLocation persiste o caminho do executável do jogo. Os containers
// abertos são fechados: a pasta de busca mudou e os handles antigos não são
// mais alcançáveis (a próxima ListVbfRoots reabre os da nova pasta).
func (c *ConfigService) SetGameExeLocation(path string) {
	interactions.NewInteractionService().FFXAppConfig().SetGameExeLocation(path)
	services.CloseVbfArchives()
}

// SetGameVersion persiste a versão do jogo.
func (c *ConfigService) SetGameVersion(version string) error {
	v, err := common.ParseGameVersionStrict(version)
	if err != nil {
		return err
	}
	interactions.NewInteractionService().FFXAppConfig().SetGameVersion(v)
	return nil
}

// SetGameLocation persiste o diretório dos arquivos do jogo, invalida caches
// e acompanha o translated se ele era derivado.
func (c *ConfigService) SetGameLocation(path string) error {
	service := interactions.NewInteractionService()
	oldGameDir := service.GameLocation.GetTargetDirectory()
	oldDefaultTranslate := common.DefaultTranslatedDir(oldGameDir)

	service.GameLocation.SetTargetDirectory(path)
	services.InvalidateViewCaches()

	// Se o translated ainda apontava para o default derivado do gamefiles anterior,
	// acompanha o novo gamefiles.
	if service.TranslateLocation.GetTargetDirectory() == oldDefaultTranslate {
		newDefaultTranslate := common.DefaultTranslatedDir(path)
		_ = service.TranslateLocation.SetTargetDirectory(newDefaultTranslate)
	}
	return nil
}

// SetTranslateLocation persiste o diretório de tradução.
func (c *ConfigService) SetTranslateLocation(path string) error {
	interactions.NewInteractionService().TranslateLocation.SetTargetDirectory(path)
	return nil
}

// SelectGameExeFile abre o seletor nativo para escolher o executável do
// jogo. O filtro é LITERAL — só FFX.exe e FFX-2.exe, sem curinga *.exe —
// e a escolha é conferida de novo aqui: o diálogo do sistema aceita digitar
// qualquer nome no campo de arquivo. Devolve "" no cancelamento e numa
// escolha inválida.
func (c *ConfigService) SelectGameExeFile() string {
	selection := pickFile(
		"Selecionar o executável do jogo (FFX.exe ou FFX-2.exe)", "",
		application.FileFilter{
			DisplayName: "Executável do jogo (FFX.exe; FFX-2.exe)",
			Pattern:     "FFX.exe;FFX-2.exe",
		},
	)
	if selection == "" {
		return ""
	}
	normalized, ok := interactions.NormalizeGameExe(selection)
	if !ok {
		// Só o caminho vai para o diagnóstico — nunca conteúdo de arquivo.
		loggingService.DiagWarn("vbf",
			"executável fora de FFX.exe/FFX-2.exe recusado",
			map[string]any{"path": selection})
		if c.notifier != nil {
			c.notifier.NotifyWarn(
				"Escolha um dos executáveis do jogo: FFX.exe ou FFX-2.exe.")
		}
		return ""
	}
	return normalized
}
