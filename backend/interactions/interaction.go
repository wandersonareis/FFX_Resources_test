package interactions

import (
	"context"
	"ffxresources/backend/common"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"path/filepath"
	"sync"
)

type InteractionService struct {
	Ctx               context.Context
	activeCtx         context.Context
	cancel            context.CancelFunc
	mu                sync.Mutex
	ffxAppConfig      IFFXAppConfig
	ffxGameVersion    models.IGameVersionProvider
	ffxTextFormat     interfaces.ITextFormatter
	GameLocation      IGameLocation
	ExtractLocation   IExtractLocation
	TranslateLocation ITranslateLocation
	ImportLocation    IImportLocation
}

var (
	interactionInstance *InteractionService
	once                sync.Once
)

func NewInteractionService() *InteractionService {
	once.Do(func() {
		filePath := filepath.Join(common.GetExecDir(), "config.json")
		ffxAppConfig := NewAppConfig(filePath)
		if err := ffxAppConfig.FromJson(); err != nil {
			panic(err)
		}

		gameVersion := models.NewFFXGameVersion(ffxAppConfig.FFXGameVersion)

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
	s := NewInteractionService()

	gameVersion := models.NewFFXGameVersion(config.FFXGameVersion)

	s.mu.Lock()
	s.ffxAppConfig = config
	s.ffxGameVersion = gameVersion
	s.mu.Unlock()

	return s
}

func NewInteractionWithCtx(ctx context.Context) *InteractionService {
	// Não travamos o mutex global novamente, confiamos em NewInteractionService para inicializar.
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

func (i *InteractionService) FFXAppConfig() IFFXAppConfig {
	return i.ffxAppConfig
}

func (i *InteractionService) FFXGameVersion() models.IGameVersionProvider {
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
