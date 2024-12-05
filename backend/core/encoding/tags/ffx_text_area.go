package tags

import "fmt"

type FFXTextTagArea struct {
	ffxTagsBase
}

func NewTextTagArea() *FFXTextTagArea {
	return &FFXTextTagArea{
		ffxTagsBase: ffxTagsBase{},
	}
}

func (a *FFXTextTagArea) FFXTextAreaCodePage() []string {
	return a.ffxTagsBase.processCodePage(&areasPage{areaByte: 0x20})
}

type areasPage struct {
	areaByte byte
}

func (a *areasPage) getMap() map[byte]string {
	areasMap := map[byte]string{
		0x30: "noname2", // unknown
		0x31: "Zanarkand",
		0x32: "Zanarkand Ruins",
		0x33: "Spira",
		0x34: "Baaj",
		0x35: "Baaj Island",
		0x36: "Temple of Yevon-Baaj",
		0x37: "Baaj Temple",
		0x38: "Besaid",
		0x39: "Besaid Island",
		0x3A: "Besaid Village",
		0x3B: "Temple of Yevon-Besaid",
		0x3C: "Besaid Temple",
		0x3D: "Chamber of the Fayth",
		0x3E: "S.S. Liki",
		0x3F: "Kilika",
		0x40: "Kilika Island",
		0x41: "Kilika Port",
		0x42: "Kilika Woods",
		0x43: "Temple of Yevon-Kilika",
		0x44: "Kilika Temple",
		0x45: "S.S. Winno",
		0x46: "Djose",
		0x47: "Djose Continent",
		0x48: "Luca",
		0x49: "Luca Seaport",
		0x4A: "Luca Stadium",
		0x4B: "Blitzball Stadium",
		0x4C: "Mi'ihen Highroad",
		0x4D: "Mi'ihen Newroad",
		0x4E: "Mi'ihen Oldroad",
		0x4F: "Mi'ihen Highroad's End",
		0x50: "Rin Travel Agency",
		0x51: "Mushroom Rock Road",
		0x52: "Temple of Yevon-Djose",
		0x53: "Djose Temple",
		0x54: "Moonflow",
		0x55: "Guadosalam",
		0x56: "Farplane",
		0x57: "Bilghen Plains",
		0x58: "Thunder Plains",
		0x59: "Macalania",
		0x5A: "Macalania Woods",
		0x5B: "Lake Macalania",
		0x5C: "Temple of Yevon-Macalania",
		0x5D: "Macalania Temple",
		0x5E: "Bikanel",
		0x5F: "Bikanel Island",
		0x60: "Sanubia Desert",
		0x61: "Home2", // Al Bhed Home
		0x62: "Home3", // Al Bhed Home
		0x63: "Bevelle",
		0x64: "St. Bevelle",
		0x65: "Temple of Yevon-Bevelle",
		0x66: "Bevelle Temple",
		0x67: "Highbridge",
		0x68: "Frost Wilds",
		0x69: "Calm Lands",
		0x6A: "Gagazet",
		0x6B: "Mt. Gagazet",
		0x6C: "Dome",
	}

	return areasMap
}

func (a *areasPage) generateCode(key byte, value string) string {
	return fmt.Sprintf("\\x%02X\\x%02X={%s}", a.areaByte, key, value)
}
