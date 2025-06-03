package common

const FFX_DIR_MARKER = "ffx_ps2"
const FFX_TEXT_FORMAT_SEPARATOR = "--------------------------------------------------End--"
const (
	MonsterMaxIndex            = 360
	PathFfxRoot                = "ffx_ps2/ffx/master/"
	OriginalsFolder            = "jppc/"
	PathOriginalsRoot          = PathFfxRoot + OriginalsFolder
	PathOriginalsKernel        = PathOriginalsRoot + "battle/kernel/"
	PathMonsterFolder          = PathOriginalsRoot + "battle/mon/"
	PathOriginalsEncounter     = PathOriginalsRoot + "battle/btl/"
	PathInternationalEncounter = PathFfxRoot + "inpc/battle/btl/"
	PathOriginalsEvent         = PathOriginalsRoot + "event/obj/"
	PathAbmap                  = PathOriginalsRoot + "menu/abmap/"
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
