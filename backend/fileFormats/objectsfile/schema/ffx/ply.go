package ffx

// ply_rom.bin / ply_save.bin — src/core/ffx/{plyrom,plysave}.cs.
// Tamanhos confirmados por test/core/excel.cs:
//
//	SizeOf<PlyRom>()  == 0x2C (44)  -> ply_rom.bin
//	SizeOf<PlySave>() == 0x94 (148) -> ply_save.bin
//
// ply_rom.bin tem 4 segmentos (chaves C# switch_text e scan_text, cada um
// standard+simplified); ply_save.bin tem 1 (name).

// PlyRom (44 bytes) — PlySave não existe aqui: é o "Player Read-Only Memory".
type PlyRom struct {
	SwitchText TextPair
	ScanText   TextPair

	Gender PlyGender // enum : byte — man/woman/aeon

	SlvReqMultA uint8 // cubica (ax^3/100 + bx^2/10 + c(x+1))
	SlvReqMultB uint8 // quadratica
	SlvReqMultC uint8 // linear
	SlvReqMax   int32 // AP maximo ate o proximo nivel (>= 101o sphere level)

	AeonStatScaling AeonStatBoostsScaling
	DoomDuration    uint8
	Reserve1        uint8 // C# `__0x2B` (chr.ram.__0x199)
}

// PlySave (148 bytes) — StructLayout Sequential, Pack=2: em Go não há
// padding implícito e todos os campos já caem nos mesmos offsets que no C#.
type PlySave struct {
	Name TextRef

	BaseHP           uint32
	BaseMP           uint32
	BaseStrength     uint8
	BaseDefense      uint8
	BaseMagic        uint8
	BaseMagicDefense uint8
	BaseAgility      uint8
	BaseLuck         uint8
	BaseEvasion      uint8
	BaseAccuracy     uint8

	TotalAP uint32
	AP      uint32

	HP    uint32
	MP    uint32
	MaxHP uint32
	MaxMP uint32

	PlyFlags  uint8 // bit0 = join, bit4 = joined
	WpnInvIdx uint8
	ArmInvIdx uint8

	Strength     uint8
	Defense      uint8
	Magic        uint8
	MagicDefense uint8
	Agility      uint8
	Luck         uint8
	Evasion      uint8
	Accuracy     uint8

	PoisonDmg uint8

	LimitModeIndex uint8
	LimitCharge    uint8
	LimitChargeMax uint8

	SlvAvailable         uint8
	SlvSpent             uint8
	BattlesUntilRecovery uint8

	ABIMap             AbilityMap
	AutoAbilityEffects AutoAbilityEffectsMap

	BattleCount     uint32
	EnemiesDefeated uint32
	Deaths          uint32
	LimitsCharged   uint32

	LimitModeCounters  [20]uint16
	ObtainedLimitModes OverdriveModeFlags

	Reserve1 uint32 // C# `__0x8C`
	Reserve2 uint32 // C# `__0x90`
}

// OverdriveModeFlags enum : uint (src/core/ffx/plysave.cs).
type OverdriveModeFlags uint32
