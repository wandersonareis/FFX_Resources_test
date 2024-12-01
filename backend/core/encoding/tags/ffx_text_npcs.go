package tags

import "fmt"

type FFXTextTagNPC struct {
	npcByte byte
}

func NewTextTagNPC() *FFXTextTagNPC {
	npcByte := byte(0x19)

	return &FFXTextTagNPC{
		npcByte: npcByte,
	}
}

func (n *FFXTextTagNPC) FFXTextNPCCodePage() []string {
	return n.generateNPCs()
}

func (n *FFXTextTagNPC) generateNPCs() []string {
	npcsMap := n.getNPCsMap()
	codePage := make([]string, 0, len(npcsMap))

	for key, value := range npcsMap {
		codePage = append(codePage, n.generateNPCCode(key, value))
	}

	return codePage
}

func (n *FFXTextTagNPC) generateNPCCode(key byte, value string) string {
	return fmt.Sprintf("\\x%02X\\x%02X={%s}", n.npcByte, key, value)
}

func (n *FFXTextTagNPC) getNPCsMap() map[byte]string {
	npcsMap := map[byte]string{
		0x30: "noname", // unknown
		0x31: "Gatta",
		0x32: "Luzzu",
		0x33: "O'aka",
		0x34: "O'aka XXIII",
		0x35: "Dona",
		0x36: "Barthello",
		0x37: "Wantz",
		0x38: "Isaaru",
		0x39: "Maroda",
		0x3A: "Pacce",
		0x3B: "Shelinda",
		0x3C: "Maechen",
		0x3D: "Lucil",
		0x3E: "Elma",
		0x3F: "Clasko",
		0x40: "Tromell",
		0x41: "Tromell Guado",
		0x42: "Biran",
		0x43: "Biran Ronso",
		0x44: "Yenke",
		0x45: "Yenke Ronso",
		0x46: "Rin",
		0x47: "Tidus's Mother",
		0x48: "Chappu",
		0x49: "Mika",
		0x4A: "Maester Mika",
		0x4B: "Maester Mika2",
		0x4C: "Seymour's Mother",
		0x4D: "Mira",
		0x4E: "Summoner Mira",
		0x4F: "Zuke",
		0x50: "Ex-Summoner Zuke",
		0x51: "Thorton",
		0x52: "Summoner Thorton",
		0x53: "Yunalesca",
		0x54: "Zaon",
		0x55: "Braska",
		0x56: "High Summoner Braska",
		0x57: "Jyscal",
		0x58: "Jyscal Guado",
		0x59: "Cid",
		0x5A: "Yuna's Mother",
		0x5B: "Brother",
		0x5C: "Kinoc",
		0x5D: "Maester Kinoc",
		0x5E: "Wen Kinoc",
		0x5F: "Kelk",
		0x60: "Master Kelk Ronso",
		0x61: "Belgemine",
		0x62: "Summoner Belgemine",
		0x63: "Seymor",
		0x64: "Master Seymor",
		0x65: "Yocun",
		0x66: "High Summoner Yocun",
		0x67: "Ohalland",
		0x68: "High Summoner Ohalland",
		0x69: "Gandof",
		0x6A: "High Summoner Gandof",
		0x6B: "Mi'ihen",
		0x6C: "Operation Mi'ihen",
		0x6D: "fayth",
		0x6E: "Yevon",
		0x6F: "Yevon2", // Yevon
		0x70: "Sin",
		0x71: "blitzball",
		0x72: "Besaid Aurochs",    // blitzball team
		0x73: "Aurochs",           // blitzball team
		0x74: "Kilika Beasts",     // blitzball team
		0x75: "Luca Goers",        // blitzball team
		0x76: "Ronso Fangs",       // blitzball team
		0x77: "Guado Glories",     // blitzball team
		0x78: "Al Bhed Psyches",   // blitzball team
		0x79: "Bevelle Bells",     // blitzball team
		0x7A: "Yocun Nomads",      // blitzball team
		0x7B: "Zanarkand Abes",    // blitzball team
		0x7C: "Zanarkand Duggles", // blitzball team
		0x7D: "Al Bhed",
		0x7E: "Al Bhed2", // Al Bhed
		0x7F: "Guado",
		0x80: "Guado2", // Guado
		0x81: "Ronso",
		0x82: "Ronso2", // Ronso
		0x83: "Crusader",
		0x84: "Jecht",
		0x85: "sinspawn",
		0x86: "fiend",
		0x87: "summoner",
		0x88: "airship",
		0x89: "Kimahri Ronso",
		0x8A: "Kelk Ronso",
		0x8B: "password", //unknown, but using character byte 0x13
		0x8C: "Al Bhed3", // Al Bhed
		0x8D: "Home",     // Al Bhed Home
		0x8E: "guardian",
		0x8F: "one minute",   // 1 minute, but using character byte 0x13
		0x90: "Three minute", // 3 minutes, but using character byte 0x13
		0x91: "1000 years",
		0x92: "Gah-hah-hah-hah",
		0x93: "Calm",
		0x94: "magic",
		0x95: "aeons",
		0x96: "Yunie",
		0x97: "blitz",
		0x98: "Rikku2", // Rikku
		0x99: "FINAL FANTASY X",
		0x9A: "machina",
		0x9B: "summoner2",
		0x9C: "guradians",
	}

	return npcsMap
}
