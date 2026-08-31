package sharedutils

var (
	playerCharMap = map[byte]string{
		0x00: "TIDUS",
		0x01: "YUNA",
		0x02: "AURON",
		0x03: "KIMAHRI",
		0x04: "WAKKA",
		0x05: "LULU",
		0x06: "RIKKU",
		0x07: "SEYMOUR",
		0x08: "VALEFOR",
		0x09: "IFRIT",
		0x0A: "IXION",
		0x0B: "SHIVA",
		0x0C: "BAHAMUT",
		0x0D: "ANIMA",
		0x0E: "YOJIMBO",
		0x0F: "CINDY",
		0x10: "SANDY",
		0x11: "MINDY",
		0x12: "DUMMY",
		0x13: "DUMMY2",
	}
	iconMap = map[byte]string{
		0x20: "?L1 (SWITCH)",
		0x2D: "Dummy",
		0x2E: "Dummy2",
		0x30: "TRIANGLE",
		0x31: "X",
		0x32: "CIRCLE",
		0x33: "SQUARE",
		0x34: "L1",
		0x35: "R1",
		0x36: "L2",
		0x37: "R2",
		0x38: "START",
		0x39: "SELECT",
		0x40: "Direcional",
		0x41: "Direcional UP",
		0x42: "Direcional RIGHT",
		0x43: "Direcional Up+Right",
		0x44: "Direcional DOWN",
		0x45: "Direcional Up+Down",
		0x46: "Direcional Down+Right",
		0x47: "Direcional Up+Right+Down",
		0x48: "Direcional LEFT",
		0x49: "Direcional Up+Left",
		0x4A: "Direcional Left+Right",
		0x4B: "Direcional Up+Left+Right",
		0x4C: "Direcional Left+Down",
		0x4D: "Direcional Up+Left+Down",
		0x4E: "Direcional Left+Down+Right",
		0x4F: "Direcional All",
		0x80: "Red Gate",
		0x81: "Green Gate",
		0x82: "Yellow Gate",
		0x83: "Blue Gate",
	}
)

func GetPlayerChar(pc byte) string {
	if name, ok := playerCharMap[pc]; ok {
		return name
	}
	return "?"
}

func GetIconName(iconIdx byte) string {
	if name, ok := iconMap[iconIdx]; ok {
		return name
	}
	return "?"
}
