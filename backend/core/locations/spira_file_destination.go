package locations

import (
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
)

type IDestination interface {
	Extract() IExtractLocation
	Translate() ITranslateLocation
	Import() IImportLocation
	InitializeLocations(source interfaces.ISource, formatter interfaces.ITextFormatter) error
}

type Destination struct {
	ExtractLocation   IExtractLocation   `json:"extract_location"`
	TranslateLocation ITranslateLocation `json:"translate_location"`
	ImportLocation    IImportLocation    `json:"import_location"`
}

func NewDestination() IDestination {
	interactionService := interactions.NewInteractionService()

	gameVersionDir := interactionService.FFXGameVersion().GetGameVersion().String()

	extractPath := interactionService.ExtractLocation.GetTargetDirectory()
	translatePath := interactionService.TranslateLocation.GetTargetDirectory()
	importPath := interactionService.ImportLocation.GetTargetDirectory()

	destination := &Destination{
		ExtractLocation:   NewExtractLocation("extracted", extractPath, gameVersionDir),
		TranslateLocation: NewTranslateLocation("translated", translatePath, gameVersionDir),
		ImportLocation:    NewImportLocation("reimported", importPath, gameVersionDir),
	}
	return destination
}

func (g *Destination) InitializeLocations(source interfaces.ISource, formatter interfaces.ITextFormatter) error {
	if err := g.ExtractLocation.BuildExtractOutput(source, formatter); err != nil {
		return err
	}

	if err := g.TranslateLocation.BuildExtractOutput(source, formatter); err != nil {
		return err
	}

	if err := g.ImportLocation.BuildImportOutput(source, formatter); err != nil {
		return err
	}
	return nil
}

func (g *Destination) Extract() IExtractLocation {
	return g.ExtractLocation
}

func (g *Destination) Translate() ITranslateLocation {
	return g.TranslateLocation
}

func (g *Destination) Import() IImportLocation {
	return g.ImportLocation
}
