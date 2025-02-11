package locations

import "ffxresources/backend/core/locations/locationsBase"

type ITranslateLocationInfo interface {
	Get() *TranslateLocation
	Set(translateLocation TranslateLocation)
}

type TranslateLocationInfo struct {
	TranslateLocation TranslateLocation `json:"translate_location"`
}

func NewTranslateLocationInfo(opts ...locationsBase.LocationBaseOption) TranslateLocationInfo {
	options := locationsBase.ProcessOpts(opts)
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
