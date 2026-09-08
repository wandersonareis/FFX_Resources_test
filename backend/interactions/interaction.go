package interactions

import (
	"context"
	"ffxresources/backend/common"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"sync"
)

type InteractionService struct {
	Ctx               context.Context
	activeCtx         context.Context
	cancel            context.CancelFunc
	mu                sync.Mutex
	ffxAppConfig      IAppConfig
	ffxTextFormat     interfaces.ITextFormatter
	GameLocation      IGameLocation
	ExtractLocation   IExtractLocation
	TranslateLocation ITranslateLocation
	ImportLocation    IImportLocation
}

var interactionInstance *InteractionService

func resolveLocationPath(config IAppConfig, configKey, defaultDirName string) string {
	path := config.GetLocation(configKey)
	if path != "" {
		return path
	}

	fa, err := common.NewFileAccessor(defaultDirName)
	if err != nil {
		return ""
	}
	return fa.ResolvedPath
}

func NewInteractionService() *InteractionService {
	if interactionInstance == nil {
		config := NewAppConfig()

		gameDir := resolveLocationPath(config, "GameFilesLocation", common.DirData)
		extractDir := resolveLocationPath(config, "ExtractLocation", common.DirExtracted)
		translateDir := resolveLocationPath(config, "TranslateLocation", common.DirTranslated)
		importDir := resolveLocationPath(config, "ImportLocation", common.DirReimported)

		interactionInstance = &InteractionService{
			Ctx:               context.Background(),
			ffxAppConfig:      config,
			GameLocation:      newGameLocation(gameDir, config),
			ExtractLocation:   newExtractLocation(extractDir, config),
			TranslateLocation: newTranslateLocation(translateDir, config),
			ImportLocation:    newImportLocation(importDir, config),
		}

		if gameDir != "" {
			common.SetGameFilesRoot(gameDir)
		}
	}
	return interactionInstance
}

func NewInteractionServiceWithConfig(config *AppConfig) *InteractionService {
	s := NewInteractionService()

	gameDir := resolveLocationPath(config, "GameFilesLocation", common.DirData)
	extractDir := resolveLocationPath(config, "ExtractLocation", common.DirExtracted)
	translateDir := resolveLocationPath(config, "TranslateLocation", common.DirTranslated)
	importDir := resolveLocationPath(config, "ImportLocation", common.DirReimported)

	s.mu.Lock()
	s.ffxAppConfig = config
	s.GameLocation = newGameLocation(gameDir, config)
	s.ExtractLocation = newExtractLocation(extractDir, config)
	s.TranslateLocation = newTranslateLocation(translateDir, config)
	s.ImportLocation = newImportLocation(importDir, config)
	s.mu.Unlock()

	if gameDir != "" {
		common.SetGameFilesRoot(gameDir)
	}

	return s
}

func NewInteractionWithCtx(ctx context.Context) *InteractionService {
	s := NewInteractionService()

	s.mu.Lock()
	s.Ctx = ctx
	activeCtx, cancel := context.WithCancel(ctx)
	s.activeCtx = activeCtx
	s.cancel = cancel
	s.mu.Unlock()

	return s
}

func NewInteractionWithTextFormatter(formatter interfaces.ITextFormatter) *InteractionService {
	s := NewInteractionService()

	s.mu.Lock()
	s.ffxTextFormat = formatter
	s.mu.Unlock()
	return s
}

func (i *InteractionService) FFXAppConfig() IAppConfig {
	return i.ffxAppConfig
}

func (i *InteractionService) FFXGameVersion() models.IGameVersionProvider {
	return models.NewFFXGameVersion(i.ffxAppConfig.GetGameVersion())
}

func (i *InteractionService) TextFormatter() interfaces.ITextFormatter {
	return i.ffxTextFormat
}

func (i *InteractionService) Start() context.Context {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.activeCtx, i.cancel = context.WithCancel(i.Ctx)

	return i.activeCtx
}

func (i *InteractionService) Stop() {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.cancel != nil {
		i.cancel()
		i.cancel = nil
	}
}

func (i *InteractionService) GetActiveCtx() context.Context {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.activeCtx
}
