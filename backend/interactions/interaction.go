package interactions

import (
	"context"
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"path/filepath"
	"sync"
)

type Interaction struct {
	Ctx                 context.Context
	activeCtx           context.Context
	cancel              context.CancelFunc
	mu                  sync.Mutex
	ffxAppConfig        IFFXAppConfig
	ffxGameVersion      core.IFfxGameVersion
	GameLocation        IGameLocation
	DcpAndLockitOptions IDcpAndLockitOptions
	ExtractLocation     IExtractLocation
	TranslateLocation   ITranslateLocation
	ImportLocation      IImportLocation
}

var interactionInstance *Interaction

var initOnce sync.Once

func NewInteraction() *Interaction {
	initOnce.Do(func() {
		filePath := filepath.Join(common.GetExecDir(), "config.json")

		ffxAppConfig := newAppConfig(filePath)
		if err := ffxAppConfig.FromJson(); err != nil {
			panic(err)
		}

		gameVersion := core.NewFFXGameVersion()
		gameVersion.SetGameVersionNumber(ffxAppConfig.FFXGameVersion)

		ffxAppConfig.FFXGameVersion = gameVersion.GetGameVersionNumber()

		interactionInstance = &Interaction{
			Ctx:                 context.Background(),
			ffxAppConfig:        ffxAppConfig,
			ffxGameVersion:      gameVersion,
			GameLocation:        newGameLocation(),
			ExtractLocation:     newExtractLocation(),
			TranslateLocation:   newTranslateLocation(),
			ImportLocation:      newImportLocation(),
			DcpAndLockitOptions: newDcpAndLockitOptions(gameVersion),
		}
	})

	return interactionInstance
}

func NewInteractionWithCtx(ctx context.Context) *Interaction {
	if interactionInstance == nil {
		interactionInstance = NewInteraction()
	}

	interactionInstance.Ctx = ctx

	activeCtx, cancel := context.WithCancel(ctx)

	interactionInstance.activeCtx = activeCtx
	interactionInstance.cancel = cancel

	return interactionInstance
}

/* func Get() *Interaction {
	return interactionInstance
} */

func (i *Interaction) FFXAppConfig() IFFXAppConfig {
	return i.ffxAppConfig
}

func (i *Interaction) FFXGameVersion() core.IFfxGameVersion {
	return i.ffxGameVersion
}

func (i *Interaction) Start() context.Context {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.activeCtx, i.cancel = context.WithCancel(i.Ctx)

	return i.activeCtx
}

func (i *Interaction) Stop() {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.cancel != nil {
		i.cancel()
		i.cancel = nil
	}
}

func (i *Interaction) GetActiveCtx() context.Context {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.activeCtx
}
