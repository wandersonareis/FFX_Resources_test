package locations

type ITranslateLocationInfo interface {
	Get() *TranslateLocation
	Set(translateLocation TranslateLocation)
}

type TranslateLocationInfo struct {
	TranslateLocation TranslateLocation `json:"translate_location"`
}

func NewTranslateLocationInfo() TranslateLocationInfo {
	return TranslateLocationInfo{
		TranslateLocation: *NewTranslateLocation(),
	}
}

func (t *TranslateLocationInfo) Get() *TranslateLocation {
	return &t.TranslateLocation
}

func (t *TranslateLocationInfo) Set(translateLocation TranslateLocation) {
	t.TranslateLocation = translateLocation
}
