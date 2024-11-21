package interactions

import (
	"context"
	"ffxresources/backend/core"
)

type Interaction struct {
	Ctx               context.Context
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
	return interactionInstance
}
