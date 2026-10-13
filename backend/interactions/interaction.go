package interactions

import (
	"ffxresources/backend/common"
	"path/filepath"
	"sync"
)

// InteractionService é o ponto de acesso único às localizações do jogo
// (gamefiles, translated) e ao config persistido. O singleton é criado sob
// demanda e protegido por mu para acesso concorrente seguro (goroutines
// de preload, workers de import).
type InteractionService struct {
	mu                sync.RWMutex
	ffxAppConfig      IAppConfig
	GameLocation      IGameLocation
	TranslateLocation ITranslateLocation
}

var (
	interactionInstance *InteractionService
	interactionOnce     sync.Once
)

func resolveLocationPath(config IAppConfig, configKey, defaultDirName string) string {
	if config != nil {
		if path := config.GetLocation(configKey); path != "" {
			return path
		}
	}

	fa, err := common.NewFileAccessor(defaultDirName)
	if err != nil {
		return ""
	}
	return fa.ResolvedPath
}

func resolveTranslatePath(config IAppConfig, gameDir string) string {
	if config != nil {
		if path := config.GetLocation("TranslateLocation"); path != "" {
			return path
		}
	}

	if gameDir != "" {
		return common.DefaultTranslatedDir(gameDir)
	}

	fa, err := common.NewFileAccessor(common.DirTranslated)
	if err != nil {
		return ""
	}
	return fa.ResolvedPath
}

func defaultAppConfig() *AppConfig {
	return &AppConfig{
		filePath:    filepath.Join(common.GetExecDir(), "config", "config.json"),
		locations:   defaultLocations(),
		gameVersion: common.GameVersionFFX,
	}
}

// applyConfig injeta o config no singleton, resolve as localizações e
// sincroniza os globals (GameFilesRoot, CurrentGameVersion). Usado pelo
// construtor normal e pelo construtor com config customizado (testes/scripts).
func (s *InteractionService) applyConfig(config IAppConfig) {
	gameDir := resolveLocationPath(config, "GameFilesLocation", common.DirData)
	translateDir := resolveTranslatePath(config, gameDir)

	s.mu.Lock()
	s.ffxAppConfig = config
	s.GameLocation = newGameLocation(gameDir, config)
	s.TranslateLocation = newTranslateLocation(translateDir, config)
	s.mu.Unlock()

	if gameDir != "" {
		common.SetGameFilesRoot(gameDir)
	}
	common.SetCurrentGameVersion(config.GetGameVersion())
}

// NewInteractionService devolve o singleton, criando na primeira chamada
// a partir do config.json padrão.
func NewInteractionService() *InteractionService {
	interactionOnce.Do(func() {
		config := NewAppConfig()
		if config == nil {
			config = defaultAppConfig()
		}
		interactionInstance = &InteractionService{}
		interactionInstance.applyConfig(config)
	})
	return interactionInstance
}

// NewInteractionServiceWithConfig re-injeta um config customizado no
// singleton. Existe para testes e scripts CLI que precisam apontar para
// diretórios de teste em vez dos arquivos reais do jogo.
func NewInteractionServiceWithConfig(config *AppConfig) *InteractionService {
	s := NewInteractionService()
	s.applyConfig(config)
	return s
}

// FFXAppConfig devolve o config atual (leitura pontual; valores são
// repassados aos chamadores, sem acoplamento ao estado interno).
func (i *InteractionService) FFXAppConfig() IAppConfig {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.ffxAppConfig
}
