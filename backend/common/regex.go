package common

import (
	"regexp"
)

var (
	//RegexFullSpira = regexp.MustCompile(`^.*\\new_uspc\\(?:event\\obj_ps3(?!.*crcr0000|.*crjiten|.*monlist)|lastmiss\\kernel\\|battle\\kernel\\|cloudsave|menu\\tutorial*|menu\\macrodic*|battle\\btl\\(?:bika0([789]|1[34])(?=_2[345])|bsil0[57]|crsm0[09][09][02]|genk(0[0]|1[56])|kami(0[03])|kino04(?=_2[345]|(?=_1))|klyt00|mcfr(0[012])|nagi00|stbv03|system_01|tuto0000)).+\.(bin|msb|dcp)$.*`)

	IsValidPath = regexp.MustCompile(`^.*\\ffx_ps2\\ffx2\\master\\new_[^\\]*?pc.+\.(bin|msb|dcp)$.*`)
)
