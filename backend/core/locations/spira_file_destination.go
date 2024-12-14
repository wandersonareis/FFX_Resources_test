package locations

import (
	"ffxresources/backend/interfaces"
	"os"
	"strings"
)

/* type ITargetExtractLocation interface {
	GetExtractLocation() IExtractLocation
}

type ITargetTranslateLocation interface {
	GetTranslateLocation() ITranslateLocation
}

type ITargetImportLocation interface {
	GetImportLocation() IImportLocation
} */

/* type IDestination interface {
	ITargetExtractLocation
	ITargetTranslateLocation
	ITargetImportLocation

	InitializeLocations(formatter formatters.ITextFormatterDev)
	CreateRelativePath(gameLocationPath string)
} */

type IDestination interface {
	//ITargetExtractLocation
	//ITargetTranslateLocation
	//ITargetImportLocation
	//IExtractLocationInfo

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
	return &Destination{
		ExtractLocation:   NewExtractLocationInfo(),
		TranslateLocation: NewTranslateLocationInfo(),
		ImportLocation:    NewImportLocationInfo(),
	}
}

func (g *Destination) InitializeLocations(source interfaces.ISource, formatter interfaces.ITextFormatterDev) {
	g.ExtractLocation.Get().BuildTargetOutput(source, formatter)
	g.TranslateLocation.Get().BuildTargetOutput(source, formatter)
	g.ImportLocation.Get().BuildTargetOutput(source, formatter)
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
