package locations

import (
	"ffxresources/backend/core/locations/locationsBase"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"os"
	"strings"
)

type IDestination interface {
	Extract() IExtractLocation
	Translate() ITranslateLocationInfo
	Import() IImportLocation
	InitializeLocations(source interfaces.ISource, formatter interfaces.ITextFormatter)
	CreateRelativePath(source interfaces.ISource, gameLocationPath string)
}

type Destination struct {
	ExtractLocation   IExtractLocation   `json:"extract_location"`
	TranslateLocation TranslateLocationInfo `json:"translate_location"`
	ImportLocation    IImportLocation    `json:"import_location"`
}

func NewDestination() IDestination {
	interactionService := interactions.NewInteractionService()

	gameVersionDir := interactionService.FFXGameVersion().GetGameVersion().String()

	extractPath := interactionService.ExtractLocation.GetTargetDirectory()
	translatePath := interactionService.TranslateLocation.GetTargetDirectory()
	importPath := interactionService.ImportLocation.GetTargetDirectory()

	destination := &Destination{
		ExtractLocation: NewExtractLocation("extracted", extractPath, gameVersionDir),
		TranslateLocation: NewTranslateLocationInfo(locationsBase.WithDirectoryName("translated"), locationsBase.WithTargetDirectory(translatePath), locationsBase.WithGameVersionDir(gameVersionDir)),
		ImportLocation: NewImportLocation("reimported", importPath, gameVersionDir),
	}
	return destination
}

func (g *Destination) InitializeLocations(source interfaces.ISource, formatter interfaces.ITextFormatter) {
	g.ExtractLocation.BuildTargetReadOutput(source, formatter)
	g.TranslateLocation.Get().BuildTargetReadOutput(source, formatter)
	g.ImportLocation.BuildTargetWriteOutput(source, formatter)
}

func (g *Destination) Extract() IExtractLocation {
	return g.ExtractLocation
}

func (g *Destination) Translate() ITranslateLocationInfo {
	return &g.TranslateLocation
}

func (g *Destination) Import() IImportLocation {
	return g.ImportLocation
}

// CreateRelativePath sets the RelativeGameDataPath of the source to a path relative to the given gameLocationPath.
// If the FullFilePath of the source starts with the gameLocationPath, the gameLocationPath is trimmed from the FullFilePath
// and the result is assigned to RelativeGameDataPath.
//
// Parameters:
//   - gameLocationPath: The path for game original files to which the FullFilePath should be made relative.
func (g *Destination) CreateRelativePath(source interfaces.ISource, gameLocationPath string) {
	if strings.HasPrefix(source.Get().Path, gameLocationPath) {
		source.Get().RelativePath = strings.TrimPrefix(source.Get().Path, gameLocationPath+string(os.PathSeparator))
	}
}
