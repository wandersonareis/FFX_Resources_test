package lib

import "context"

type Interaction struct {
	Ctx               context.Context
	GameLocation      *GameLocation
	ExtractLocation   *ExtractLocation
	TranslateLocation *TranslateLocation
	ImportLocation    *ImportLocation
	TempProvider      *TempProvider
}

var interactionInstance *Interaction

func NewInteraction() *Interaction {
	if interactionInstance == nil {
		interactionInstance = &Interaction{
			Ctx:               context.Background(),
			GameLocation:      NewGameLocation(),
			ExtractLocation:   NewExtractLocation(),
			TranslateLocation: NewTranslateLocation(),
			ImportLocation:    NewImportLocation(),
			TempProvider:      NewTempProvider(),
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
