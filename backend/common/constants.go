package common

const FFX_DIR_MARKER = "ffx_ps2"
const FFX_TEXT_FORMAT_SEPARATOR = "--------------------------------------------------End--"
const (
	MonsterMaxIndex            = 360
	OriginalsFolder            = "jppc/"
	PathTextOutputRoot         = "target/text/"
	DefaultLocalization        = "us"
	SkipBlitzballEvents        = true
)

var Localizations = map[string]string{
	"ch": "Chinese",
	"de": "German",
	"fr": "French",
	"it": "Italian",
	"jp": "Japanese",
	"kr": "Korean",
	"sp": "Spanish",
	"us": "English",
}

var Charsets = []string{"ch", "cn", "jp", "kr", "us"}
