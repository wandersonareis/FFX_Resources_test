package ffxencoding

var localizationMap = map[string]string{
	"ch": "ch",
	"kr": "kr",
	"jp": "jp",
}

func GetCharsetForLanguage(languageCode string) string {
	if charset, ok := localizationMap[languageCode]; ok {
		return charset
	}
	return "us"
}
