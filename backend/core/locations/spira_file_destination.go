package locations

import (
	"ffxresources/backend/core/locations/base"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"os"
	"strings"
)

type IDestination interface {
	Extract() IExtractLocationInfo
	Translate() ITranslateLocationInfo
	Import() IImportLocationInfo
	InitializeLocations(source interfaces.ISource, formatter interfaces.ITextFormatterDev)
	CreateRelativePath(source interfaces.ISource, gameLocationPath string)
}

type Destination struct {
	ExtractLocation   ExtractLocationInfo   `json:"extract_location"`
	TranslateLocation TranslateLocationInfo `json:"translate_location"`
	ImportLocation    ImportLocationInfo    `json:"import_location"`
}

func NewDestination() IDestination {
	_interactions := interactions.NewInteractionService()

	extractPath := _interactions.ExtractLocation.GetTargetDirectory()
	translatePath := _interactions.TranslateLocation.GetTargetDirectory()
	importPath := _interactions.ImportLocation.GetTargetDirectory()

	destination := &Destination{
		ExtractLocation:   NewExtractLocationInfo(internal.WithDirectoryName("extracted"), internal.WithTargetDirectory(extractPath)),
		TranslateLocation: NewTranslateLocationInfo(internal.WithDirectoryName("translated"), internal.WithTargetDirectory(translatePath)),
		ImportLocation:    NewImportLocationInfo(internal.WithDirectoryName("reimported"), internal.WithTargetDirectory(importPath)),
	}
	return destination
}

func (g *Destination) InitializeLocations(source interfaces.ISource, formatter interfaces.ITextFormatterDev) {
	g.ExtractLocation.Get().BuildTargetReadOutput(source, formatter)
	g.TranslateLocation.Get().BuildTargetReadOutput(source, formatter)
	g.ImportLocation.Get().BuildTargetWriteOutput(source, formatter)
}

func (g *Destination) Extract() IExtractLocationInfo {
	return &g.ExtractLocation
}

func (g *Destination) Translate() ITranslateLocationInfo {
	return &g.TranslateLocation
}

func (g *Destination) Import() IImportLocationInfo {
	return &g.ImportLocation
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
