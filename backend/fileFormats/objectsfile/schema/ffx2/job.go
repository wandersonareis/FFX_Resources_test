package ffx2

// job.bin — ffx_ps2/ffx2/master/jppc/battle/kernel/job.h
// (C#: src/core/ffx2/job.cs). individual = 0xE4 (228) — igual ao struct C#.
//
// Chaves: name@0x00, help@0x04, creature_data_help@0xAC
// (o JobLayout legado lê name, description@0x04 e effect@0xAC).

// StatGrowthHp (3 bytes) — hp = base + linear_mult*level - level^2/(quadratic_div/10).
type StatGrowthHp struct {
	LinearMult   uint8
	QuadraticDiv uint8
	BaseAmount   uint8
}

// StatGrowthMp (3 bytes) — mp = base + (linear_mult/10)*level - level^2/quadratic_div.
type StatGrowthMp struct {
	LinearMult   uint8
	QuadraticDiv uint8
	BaseAmount   uint8
}

// StatGrowthGeneric (5 bytes) — stats físicos/mágicos.
type StatGrowthGeneric struct {
	LinearMult    uint8
	LinearDiv     uint8
	BaseAmount    uint8
	QuadraticDivA uint8
	QuadraticDivB uint8
}

// JobWeaponData (4 bytes) — InlineArray(4) dentro de JobWeapons.
type JobWeaponData struct {
	WeaponModel    uint16
	WeaponPosition uint16
}

// JobWeapons (16 bytes) — InlineArray(4) de JobWeaponData.
type JobWeapons [4]JobWeaponData

// JobCreatureData (0x38) — StructLayout Explicit, Size=0x38 no C#: os gaps
// 0x0C..0x1C e 0x26..0x28 não têm campo no C#; aqui viram reserves nomeados.
type JobCreatureData struct {
	Help        TextRef // C# `creature_data.help`
	Abilities   [2]UnlockableAbility
	Reserve1    [0x10]uint8 // gap 0x0C..0x1C não nomeado no C#
	StatChanges StatChanges
	Reserve2    [2]uint8 // gap 0x26..0x28 não nomeado no C#
	LevelGrowth FeedStatChanges
}

// Job (228 bytes).
type Job struct {
	Name TextRef
	Help TextRef

	User          uint8
	Data          uint8
	OrderingIdx   uint8
	Icon          uint8
	BerserkAction T_X2CommandId

	GrowthHP StatGrowthHp
	GrowthMP StatGrowthMp

	GrowthStrength     StatGrowthGeneric
	GrowthDefense      StatGrowthGeneric
	GrowthMagic        StatGrowthGeneric
	GrowthMagicDefense StatGrowthGeneric
	GrowthAgility      StatGrowthGeneric
	GrowthEvasion      StatGrowthGeneric
	GrowthAccuracy     StatGrowthGeneric
	GrowthLuck         StatGrowthGeneric

	Abilities  [16]UnlockableAbility
	WeaponData [3]JobWeapons

	CreatureData JobCreatureData
}
