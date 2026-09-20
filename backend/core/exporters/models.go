package exporters

type (
	EventFileData struct {
		ID      string            `json:"id"`
		Strings []EventStringData `json:"strings"`
	}
	EventStringData struct {
		Index int               `json:"index"`
		Text  map[string]string `json:"text"`
	}
)
