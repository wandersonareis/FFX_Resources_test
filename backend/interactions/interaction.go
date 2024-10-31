package interactions

import "context"

type Interaction struct {
	Ctx               context.Context
	GameLocation      *GameLocation
	GamePart          *FfxGamePart
	ExtractLocation   *ExtractLocation
	TranslateLocation *TranslateLocation
	ImportLocation    *ImportLocation
}

var interactionInstance *Interaction

func NewInteraction() *Interaction {
	if interactionInstance == nil {
		interactionInstance = &Interaction{
			Ctx:               context.Background(),
			GameLocation:      NewGameLocation(),
			GamePart:          NewFfxGamePart(),
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
