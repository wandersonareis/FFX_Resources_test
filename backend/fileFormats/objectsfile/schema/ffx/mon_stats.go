package ffx

// monster1/2/3.bin (stats de monstros) — src/core/ffx/battle/mon_stats.cs.
// individual = 128B — igual ao struct C# (validado: sem divergência PC).
//
// Chaves C# (src/core/ffx/battle/mon_stats.cs): `name` é TextRef
// (1 chave @0x00); sensor_text e scan_text são simplificáveis (2 chaves
// cada, @0x04..0x14).
type MonStats struct {
	Name       TextRef
	SensorText TextPair
	ScanText   TextPair

	Hp                uint32
	Mp                uint32
	OverkillThreshold uint32

	Strength     uint8
	Defense      uint8
	Magic        uint8
	MagicDefense uint8
	Agility      uint8
	Luck         uint8
	Evasion      uint8
	Accuracy     uint8

	SpecialResistances ChrResistFlags
	PoisonDmg          uint8
	ElemAbsorb         ElementFlags
	ElemIgnore         ElementFlags
	ElemResist         ElementFlags
	ElemWeak           ElementFlags

	StatusResist        StatusMap
	StatusAutoPermanent StatusPermanentFlags
	StatusAutoTemporal  StatusTemporalFlags
	StatusAutoExtra     StatusExtraFlags
	StatusResistExtra   StatusExtraFlags

	CommandList     [16]uint16
	ForcedMove      uint16
	MonsterIdx      uint16
	ModelIdx        uint16
	CtbIconType     uint8
	DoomCounter     uint8
	MonsterArenaIdx uint16
	SoundBankRef    uint16
	AlwaysZero      uint32 // C# `always_zero` (privado)
}
