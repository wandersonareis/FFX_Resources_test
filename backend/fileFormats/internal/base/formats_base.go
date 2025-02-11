package base

import (
	"context"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
)

type (
	IFormatsBase interface {
		GetCtx() context.Context
		GetSource() interfaces.ISource
		GetDestination() locations.IDestination
	}

	FormatsBase struct {
		Ctx         context.Context
		Source      interfaces.ISource
		Destination locations.IDestination
	}
)

func NewFormatsBase(source interfaces.ISource, destination locations.IDestination) *FormatsBase {
	return &FormatsBase{
		Ctx:         interactions.NewInteractionService().Ctx,
		Source:      source,
		Destination: destination,
	}
}

func (f *FormatsBase) GetCtx() context.Context {
	return f.Ctx
}

func (f *FormatsBase) GetSource() interfaces.ISource {
	return f.Source
}

func (f *FormatsBase) GetDestination() locations.IDestination {
	return f.Destination
}
