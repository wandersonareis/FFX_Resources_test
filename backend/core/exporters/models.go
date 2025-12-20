package exporters

type (
	NameOnlyData struct {
		ID   int               `json:"id"`
		Name map[string]string `json:"name"`
	}

	NameDescriptionData struct {
		NameOnlyData
		Description map[string]string `json:"description"`
	}

	EventFileData struct {
		ID      string            `json:"id"`
		Strings []EventStringData `json:"strings"`
	}

	EventStringData struct {
		Index int               `json:"index"`
		Text  map[string]string `json:"text"`
	}
)
