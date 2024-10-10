package lib

type Interaction struct {
	ExtractLocation *ExtractLocation
	GameLocation    *GameLocation
	WorkingLocation *WorkDirectory
}

var ffx2_marker = "ffx_ps2"
var interaction *Interaction

func NewInteraction() *Interaction {
	if interaction == nil {
		interaction = &Interaction{
			ExtractLocation: NewExtractLocation(),
			GameLocation:    NewGameLocation(),
			WorkingLocation: NewWorkDirectory(),
		}
	}
	return interaction
}

func GetPathMarker() string {
	return ffx2_marker
}

func GetWorkdirectory() *WorkDirectory {
	return NewWorkDirectory()
}
