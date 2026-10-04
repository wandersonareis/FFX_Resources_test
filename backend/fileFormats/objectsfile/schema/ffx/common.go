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
// Convenções (idênticas ao pacote ffx2): TextRef/TextPair (models.Segment)
// para referência de texto, "Reserve*" para bytes privados/gaps, bool C#
// vira uint8. Não há tag: o tipo do campo diz se é texto ou dado.
//
// Chaves (chaves JSON/DTO) — snake_case do nome do campo Go, que é a forma
// Go idiomática do nome do campo no struct C# (PascalCase, inicialismos em
// maiúsculas):
//   - um TextPair gera DUAS chaves: `{campo}` (standard) e
//     `{campo}_simplified` (simplified), na ordem em que os pares aparecem no
//     struct (src/core/excel.cs);
//   - um TextRef gera só `{campo}`;
//   - arrays geram `{campo}_{indice}` (índice 0-based);
//   - campos aninhados são prefixados pelo campo pai, com `_`:
//     `creature_data.help` -> `creature_data_help`.
//
// Divergências de nome — onde snake(Go) não reproduz a grafia do C#, o campo
// carrega um comentário `// C# ...` com a origem (audit: 3 em ffx, 3 em
// ffx2; o resto é Reserve*/Dummy*/Extra* do próprio C#):
//   - CommandBody.Anim1/Anim2 <- anim_1/anim_2 (a chave fica anim1/anim2);
//   - AutoAbilityEffectsMap.Words <- `_u` (palavra de bits; no C# viram as
//     props has_*), idem AbilityMap.Words <- `_u`.
//
// Para os arquivos sem struct C# equivalente a chave deriva do layout legado
// em snake_case, com `description` -> `help` (é o nome usado pelo C# em todos
// os structs que têm os dois campos).
//
// Divergências PC (testData/FFX .../new_uspc) vs struct C# (Switch HD),
// resolvidas empiricamente:
//   - AutoAbility: 108B = layout exato do C# (validado byte-a-byte).
//   - Command: 96B = PCommand, i.e. Command (0x5C) + PCommandData (4B) no
//     fim do chunk (SizeOf<PCommand> == 0x60 em test/core/excel.cs).
//   - MonMagic: 92B = Command puro (SizeOf<Command> == 0x5C) — monmagic{1|2}.bin.
package ffx

import "ffxresources/backend/models"

// TextRef: referência a texto (offset + u16 "key" desconhecido no
// C# — models.Segment cobre os mesmos 4 bytes).
type TextRef = models.Segment

// TextPair: par (standard, simplified) de referências de
// texto (src/core/excel.cs) — 8 bytes / 2 segmentos.
type TextPair struct {
	Standard   TextRef
	Simplified TextRef
}

// T_XCommandId é o alias C# global using T_XCommandId = System.UInt16.
type T_XCommandId = uint16

// ElementFlags enum : byte.
type ElementFlags uint8

// Flags de status (C# : ushort).
type (
	StatusPermanentFlags uint16
	StatusTemporalFlags  uint16
	StatusExtraFlags     uint16
	ChrResistFlags       uint16
	StatIncreaseFlags    uint16
)

// Sphere enums (src/core/ffx/sphere.cs).
type (
	SphereBehavior uint16 // enum : ushort
	SphereTargets  uint16 // [Flags] enum : ushort
	SphereRange    uint8  // [Flags] enum : byte
)

// AutoAbilityEffectsMap (6 bytes) — InlineArray(3) de ushort no C#
// (src/core/ffx/aabimap.cs: auto-abilities, overdrives, stroll/capture).
type AutoAbilityEffectsMap struct {
	Words [3]uint16 // C# _u (palavra de bits; props has_*)
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

// AbilityMap (12 bytes) — InlineArray(3) de uint no C#
// (src/core/ffx/abimap.cs: overdrives, white magic, black magic).
type AbilityMap struct {
	Words [3]uint32 // C# _u (palavra de bits; props has_*)
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
