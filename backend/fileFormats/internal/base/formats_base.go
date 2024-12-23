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

func NewFormatsBase(source interfaces.ISource, destination locations.IDestination) *FormatsBase {
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
