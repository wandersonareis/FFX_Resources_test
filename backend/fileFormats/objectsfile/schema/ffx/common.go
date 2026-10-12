// Package ffx mapeia os binários de battle/kernel (FFX) em structs Go, no
// estilo das structs [StructLayout] do Fahrenheit (C#):
// https://github.com/fahrenheit-crew/fahrenheit
//
// Origem dos campos (adaptados de C# para Go, sem padding implícito):
//
//	src/core/ffx/{command,important,aability,panel,sphere,text,plyrom,plysave}.cs
//	src/core/ffx/battle/mon_stats.cs
//	src/core/ffx/{status,element,typedefs,summon,abimap}.cs, src/core/excel.cs
//
// TextRef/TextPair identificam explicitamente os campos textuais. O chunkmap
// usa os nomes dos campos Go como rótulos e preserva os outros campos como
// bytes de dados; não há tags nem nomes de campo externos usados para lookup.
//
// Divergências PC (testData/FFX .../new_uspc) vs struct C# (Switch HD),
// resolvidas empiricamente:
//   - AutoAbility: 108B = layout exato do C# (validado byte-a-byte).
//   - Command: 96B = PCommand, i.e. Command (0x5C) + PCommandData (4B) no
//     fim do chunk (SizeOf<PCommand> == 0x60 em test/core/excel.cs).
//   - MonMagic: 92B = Command puro (SizeOf<Command> == 0x5C) — monmagic{1|2}.bin.
package ffx

import "ffxresources/backend/models"

// TextRef: referência textual de 4 bytes (offset + key).
type TextRef = models.TextRef

// TextPair: dois segmentos físicos; a camada textual só exporta Simplified
// quando seu conteúdo difere de Standard.
type TextPair = models.TextPair

// T_XCommandId é o alias System.UInt16.
type T_XCommandId = uint16

// ElementFlags enum : byte.
type ElementFlags uint8

// Flags de status.
type (
	StatusPermanentFlags uint16
	StatusTemporalFlags  uint16
	StatusExtraFlags     uint16
	ChrResistFlags       uint16
	StatIncreaseFlags    uint16
)

// Sphere enums (src/core/ffx/sphere.cs).
type (
	SphereBehavior uint16 // enum
	SphereTargets  uint16 // [Flags] enum
	SphereRange    uint8  // [Flags] enum
)

// AutoAbilityEffectsMap (6 bytes)
// (src/core/ffx/aabimap.cs: auto-abilities, overdrives, stroll/capture).
type AutoAbilityEffectsMap struct {
	Words [3]uint16
}

// StatusMap (25 bytes, src/core/ffx/status.cs): valores por status.
type StatusMap struct {
	Death         uint8
	Zombie        uint8
	Petrification uint8
	Poison        uint8
	PowerBreak    uint8
	MagicBreak    uint8
	ArmorBreak    uint8
	MentalBreak   uint8
	Confusion     uint8
	Berserk       uint8
	Provoke       uint8
	Threaten      uint8
	Sleep         uint8
	Silence       uint8
	Darkness      uint8
	Shell         uint8
	Protect       uint8
	Reflect       uint8
	NulTide       uint8
	NulBlaze      uint8
	NulShock      uint8
	NulFrost      uint8
	Regen         uint8
	Haste         uint8
	Slow          uint8
}

// PlyGender enum : byte (src/core/ffx/plyrom.cs).
type PlyGender uint8

// AbilityMap (12 bytes)
// (src/core/ffx/abimap.cs: overdrives, white magic, black magic).
type AbilityMap struct {
	Words [3]uint32
}

// AeonStatBoostsScaling (18 bytes, src/core/ffx/summon.cs): quanto de cada
// stat dos Aeons vem do total do grupo vs. do stat específico (Luck não tem
// scaling).
type AeonStatBoostsScaling struct {
	HPFromTotalStats             uint8
	HPFromSpecificStat           uint8
	MPFromTotalStats             uint8
	MPFromSpecificStat           uint8
	StrengthFromTotalStats       uint8
	StrengthFromSpecificStat     uint8
	DefenseFromTotalStats        uint8
	DefenseFromSpecificStat      uint8
	MagicFromTotalStats          uint8
	MagicFromSpecificStat        uint8
	MagicDefenseFromTotalStats   uint8
	MagicDefenseFromSpecificStat uint8
	AgilityFromTotalStats        uint8
	AgilityFromSpecificStat      uint8
	EvasionFromTotalStats        uint8
	EvasionFromSpecificStat      uint8
	AccuracyFromTotalStats       uint8
	AccuracyFromSpecificStat     uint8
}

// StatusDurationMap (13 bytes): duração (turnos) por status temporário.
type StatusDurationMap struct {
	Sleep    uint8
	Silence  uint8
	Darkness uint8
	Shell    uint8
	Protect  uint8
	Reflect  uint8
	NulTide  uint8
	NulBlaze uint8
	NulShock uint8
	NulFrost uint8
	Regen    uint8
	Haste    uint8
	Slow     uint8
}
