package base

import (
	"context"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"

	"github.com/rs/zerolog"
)

type FormatsBase struct {
	Ctx         context.Context
	Log         zerolog.Logger
	source      interfaces.ISource
	destination locations.IDestination
}

func NewFormatsBaseDev(source interfaces.ISource, destination locations.IDestination) *FormatsBase {
	return &FormatsBase{
		Ctx:         interactions.NewInteraction().Ctx,
		Log:         logger.Get(),
		source:      source,
		destination: destination,
	}
}

func (f *FormatsBase) GetFileInfo() interfaces.ISource {
	return f.source
}

func (f *FormatsBase) SetFileInfo(dataInfo interfaces.ISource) {
	f.source = dataInfo
}

func (f *FormatsBase) GetGameData() interfaces.ISource {
	return f.source
}

func (f *FormatsBase) Source() interfaces.ISource {
	return f.source
}

func (f *FormatsBase) Destination() locations.IDestination {
	return f.destination
}

/* func (f *FormatsBase) GetExtractLocation() *interactions.ExtractLocation {
	return f.destination.GetExtractLocation()
}

func (f *FormatsBase) GetTranslateLocation() *interactions.TranslateLocation {
	return f.dataInfo.GetTranslateLocation()
}

func (f *FormatsBase) GetImportLocation() *interactions.ImportLocation {
	return f.dataInfo.GetImportLocation()
} */
