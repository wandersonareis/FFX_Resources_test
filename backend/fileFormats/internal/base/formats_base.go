package base

import (
	"context"
	"ffxresources/backend/core"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"

	"github.com/rs/zerolog"
)

type FormatsBase struct {
	Ctx      context.Context
	Log      zerolog.Logger
	dataInfo interactions.IGameDataInfo
}

func NewFormatsBase(dataInfo interactions.IGameDataInfo) *FormatsBase {
	return &FormatsBase{
		Ctx:      interactions.NewInteraction().Ctx,
		Log:      logger.Get(),
		dataInfo: dataInfo,
	}
}

func (f *FormatsBase) GetFileInfo() interactions.IGameDataInfo {
	return f.dataInfo
}

func (f *FormatsBase) SetFileInfo(dataInfo interactions.IGameDataInfo) {
	f.dataInfo = dataInfo
}

func (f *FormatsBase) GetGameData() *core.GameFiles {
	return f.dataInfo.GetGameData()
}

func (f *FormatsBase) GetExtractLocation() *interactions.ExtractLocation {
	return f.dataInfo.GetExtractLocation()
}

func (f *FormatsBase) GetTranslateLocation() *interactions.TranslateLocation {
	return f.dataInfo.GetTranslateLocation()
}

func (f *FormatsBase) GetImportLocation() interactions.ImportLocation {
	return *f.dataInfo.GetImportLocation()
}
