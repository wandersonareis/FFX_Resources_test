package locations

import "ffxresources/backend/bases"

type ITranslateLocationInfo interface {
	Get() *TranslateLocation
	Set(translateLocation TranslateLocation)
}

type TranslateLocationInfo struct {
	TranslateLocation TranslateLocation `json:"translate_location"`
}

func NewTranslateLocationInfo(opts ...bases.LocationBaseOption) TranslateLocationInfo {
	options := bases.ProcessOpts(opts)
	return TranslateLocationInfo{
		TranslateLocation: *NewTranslateLocation(options),
	}
}

func (t *TranslateLocationInfo) Get() *TranslateLocation {
	return &t.TranslateLocation
}

func (t *TranslateLocationInfo) Set(translateLocation TranslateLocation) {
	t.TranslateLocation = translateLocation
}
