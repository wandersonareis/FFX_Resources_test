package tags

import "fmt"

type FFXTextTagLocation struct{}

func NewTextTagLocation() *FFXTextTagLocation {
	return &FFXTextTagLocation{}
}

func (l *FFXTextTagLocation) FFXTextLocationPage() []string {
	locationLetters := l.getLocationsLetters()
	locations := make([]string, 0, len(*locationLetters))

	l.generateLocationsCodePage(&locations)

	return locations
}

func (l *FFXTextTagLocation) generateLocationsCodePage(locations *[]string) {
	locationLetters := l.getLocationsLetters()

	generateLocationCode := func(value string) string {
		return fmt.Sprintf("%s=%s", value, value)
	}

	for _, value := range *locationLetters {
		*locations = append(*locations, generateLocationCode(value))
	}
}

func (l *FFXTextTagLocation) getLocationsLetters() *[]string {
	return &[]string{
		"À",
		"Á",
		"Â",
		"Ä",
		"Ç",
		"È",
		"É",
		"Ê",
		"Ë",
		"Ì",
		"Í",
		"Î",
		"Ï",
		"Ñ",
		"Ò",
		"à",
		"á",
		"â",
		"ä",
		"ç",
		"è",
		"é",
		"ê",
		"ë",
		"ì",
		"í",
		"î",
		"ï",
		"ñ",
		"ò",
	}
}
