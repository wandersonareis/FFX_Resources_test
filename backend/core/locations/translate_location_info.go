package locations

import internal "ffxresources/backend/core/locations/base"

type ITranslateLocationInfo interface {
	Get() *TranslateLocation
	Set(translateLocation TranslateLocation)
}

type TranslateLocationInfo struct {
	TranslateLocation TranslateLocation `json:"translate_location"`
}

func NewTranslateLocationInfo(opts ...internal.LocationBaseOption) TranslateLocationInfo {
	options := internal.ProcessOpts(opts)
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
