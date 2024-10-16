package lib

import "context"

type Interaction struct {
	Ctx               context.Context
	GameLocation      *GameLocation
	ExtractLocation   *ExtractLocation
	TranslateLocation *TranslateLocation
	ImportLocation    *ImportLocation
	TempProvider      *TempProvider
	WorkingLocation   *WorkDirectory
}

const ffx2_marker = "ffx_ps2"

var rootDirectory = GetExecDir()
var rootDirectoryName = ""
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
			WorkingLocation:   NewWorkDirectory(),
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

func GetPathMarker() string {
	return ffx2_marker
}

func GetWorkdirectory() *WorkDirectory {
	return NewWorkDirectory()
}
