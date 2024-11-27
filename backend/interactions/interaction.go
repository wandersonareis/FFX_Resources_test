package interactions

import (
	"context"
	"ffxresources/backend/core"
	"sync"
)

type Interaction struct {
	Ctx               context.Context
	activeCtx         context.Context
	cancel            context.CancelFunc
	mu                sync.Mutex
	GameLocation      IGameLocation
	GamePart          core.IFfxGamePart
	GamePartOptions   IGamePartOptions
	ExtractLocation   *ExtractLocation
	TranslateLocation ITranslateLocation
	ImportLocation    IImportLocation
}

var interactionInstance *Interaction

func NewInteraction() *Interaction {
	if interactionInstance == nil {
		gamePart := core.NewFfxGamePart()
		interactionInstance = &Interaction{
			Ctx:               context.Background(),
			GameLocation:      NewGameLocation(),
			GamePart:          gamePart,
			GamePartOptions:   NewGamePartOptions(gamePart),
			ExtractLocation:   NewExtractLocation(),
			TranslateLocation: NewTranslateLocation(),
			ImportLocation:    NewImportLocation(),
		}
	}
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

func Get() *Interaction {
	return interactionInstance
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
