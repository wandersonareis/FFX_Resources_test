package interactions

import "context"

type Interaction struct {
	Ctx               context.Context
	GameLocation      *GameLocation
	GamePart          *FfxGamePart
	GamePartOptions   *GamePartOptions
	ExtractLocation   *ExtractLocation
	TranslateLocation *TranslateLocation
	ImportLocation    *ImportLocation
}

var interactionInstance *Interaction

func NewInteraction() *Interaction {
	if interactionInstance == nil {
		gamePart := NewFfxGamePart()
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
