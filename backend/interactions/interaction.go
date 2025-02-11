package interactions

import (
	"context"
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/interfaces"
	"path/filepath"
	"sync"
)

type InteractionService struct {
	Ctx               context.Context
	activeCtx         context.Context
	cancel            context.CancelFunc
	mu                sync.Mutex
	ffxAppConfig      IFFXAppConfig
	ffxGameVersion    core.IFfxGameVersion
	ffxTextFormat     interfaces.ITextFormatter
	GameLocation      IGameLocation
	ExtractLocation   IExtractLocation
	TranslateLocation ITranslateLocation
	ImportLocation    IImportLocation
}

var interactionInstance *InteractionService

var initOnce sync.Once

func NewInteractionService() *InteractionService {
	initOnce.Do(func() {
		filePath := filepath.Join(common.GetExecDir(), "config.json")

		ffxAppConfig := newAppConfig(filePath)
		if err := ffxAppConfig.FromJson(); err != nil {
			panic(err)
		}

		gameVersion := core.NewFFXGameVersion()
		gameVersion.SetGameVersionNumber(ffxAppConfig.FFXGameVersion)

		ffxAppConfig.FFXGameVersion = gameVersion.GetGameVersionNumber()

		interactionInstance = &InteractionService{
			Ctx:               context.Background(),
			ffxAppConfig:      ffxAppConfig,
			ffxGameVersion:    gameVersion,
			GameLocation:      newGameLocation(),
			ExtractLocation:   newExtractLocation(),
			TranslateLocation: newTranslateLocation(),
			ImportLocation:    newImportLocation(),
		}
	})

	return interactionInstance
}

func NewInteractionServiceWithConfig(config *FFXAppConfig) *InteractionService {
	initOnce.Do(func() {
		gameVersion := core.NewFFXGameVersion()
		gameVersion.SetGameVersionNumber(config.FFXGameVersion)

		interactionInstance = &InteractionService{
			Ctx:               context.Background(),
			ffxAppConfig:      config,
			ffxGameVersion:    gameVersion,
			GameLocation:      newGameLocation(),
			ExtractLocation:   newExtractLocation(),
			TranslateLocation: newTranslateLocation(),
			ImportLocation:    newImportLocation(),
		}
	})

	return interactionInstance
}

func NewInteractionWithCtx(ctx context.Context) *InteractionService {
	if interactionInstance == nil {
		interactionInstance = NewInteractionService()
	}

	interactionInstance.Ctx = ctx

	activeCtx, cancel := context.WithCancel(ctx)

	interactionInstance.activeCtx = activeCtx
	interactionInstance.cancel = cancel

	return interactionInstance
}

func NewInteractionWithTextFormatter(formatter interfaces.ITextFormatter) *InteractionService {
	if interactionInstance == nil {
		interactionInstance = NewInteractionService()
	}

	interactionInstance.ffxTextFormat = formatter

	return interactionInstance
}

func (i *InteractionService) FFXAppConfig() IFFXAppConfig {
	return i.ffxAppConfig
}

func (i *InteractionService) FFXGameVersion() core.IFfxGameVersion {
	return i.ffxGameVersion
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
