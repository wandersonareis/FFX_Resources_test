// Package ffx2 mapeia os binários de battle/kernel (FFX-2) e lastmiss/kernel
// (Last Mission) em structs Go, no estilo das structs [StructLayout] do
// Fahrenheit (C#): https://github.com/fahrenheit-crew/fahrenheit
//
// Origem dos campos (adaptados de C# para Go, sem padding implícito —
// encoding/binary escreve apenas os campos exportados, contíguos):
//
//	src/core/ffx2/{plate,accessory,aability,LmCommand,LmMonMagic,LmMonster}.cs
//	src/core/ffx2/{status,element,job,typedefs}.cs, src/core/excel.cs
//
// Convenções (idênticas ao pacote ffx): TextRef/TextPair (models.Segment)
// para referência de texto, "Reserve*" para bytes privados/gaps, "Dummy*"
// para dummies já presentes no C#, bool C# vira uint8. Não há tag: o tipo
// do campo diz se é texto ou dado.
//
// Chaves (chaves JSON/DTO) — snake_case do nome do campo Go, que é a forma
// Go idiomática do nome do campo no struct C# (PascalCase, inicialismos em
// maiúsculas):
//   - um TextRef gera só `{campo}`;
//   - um TextPair gera DUAS chaves: `{campo}` (standard) e
//     `{campo}_simplified` (simplified), na ordem em que os pares aparecem no
//     struct (src/core/excel.cs);
//   - arrays geram `{campo}_{indice}` (índice 0-based);
//   - campos aninhados são prefixados pelo campo pai, com `_`:
//     `creature_data.help` -> `creature_data_help`.
//
// Divergências de nome — onde snake(Go) não reproduz a grafia do C#, o campo
// carrega um comentário `// C# ...` com a origem (audit: 3 em ffx, 3 em
// ffx2; o resto é Reserve*/Dummy*/Extra* do próprio C#):
//   - Command.Anim1/Anim2 <- anim_1/anim_2 (a chave fica anim1/anim2);
//   - StatusMap.Damage9999 <- damage_9999 (a chave fica damage9999);
//   - AutoAbilityEffectsMap.Words <- `_u` (palavra de bits; no C# viram as
//     props has_*).
//
// Os typos do C# são preservados de propósito (as chaves têm de casar com o
// struct de origem): gill/steal_gill/def_zantetu, category_abilty,
// prohibit_posion, prohibit_movestop, prohibit_timestop, cursol_dist_yn ...
//
// Para os arquivos sem struct C# equivalente a chave deriva do layout legado
// em snake_case, com `description` -> `help` (é o nome usado pelo C# em todos
// os structs que têm os dois campos).
package ffx2

import "ffxresources/backend/models"

// TextRef: referência a texto (offset na string table + u16 "key"
// desconhecido no C# — models.Segment cobre os mesmos 4 bytes).
type TextRef = models.Segment

// T_X2CommandId é o alias C# global using T_X2CommandId = System.UInt16.
type T_X2CommandId = uint16

// ElementFlags enum : byte (fire/ice/thunder/water/gravity/holy).
type ElementFlags uint8

// AAbilityFlags enum : uint (SOS/STAT/TURBO).
type AAbilityFlags uint32

// StatusFlags enum : uint — auto-statuses do grupo 1 (death..auto_life).
type StatusFlags uint32

// StatusFlags2 enum : uint — auto-statuses do grupo 2 (shell..invincible).
// No a_ability do PC (new_uspc) apenas StatusFlags cabe; ver AutoAbility.
type StatusFlags2 uint32

// StatusMap (24 bytes): chance/valor por status do grupo 1.
// src/core/ffx2/status.cs
type StatusMap struct {
	Death          uint8
	Petrification  uint8
	Sleep          uint8
	Silence        uint8
	Darkness       uint8
	Poison         uint8
	Confusion      uint8
	Berserk        uint8
	Curse          uint8
	Sentinel       uint8
	Eject          uint8
	DoubleHP       uint8
	DoubleMP       uint8
	Spellspring    uint8
	Damage9999     uint8 // C# damage_9999 (snake_case geraria damage9999)
	AlwaysCritical uint8
	Pointless      uint8
	Itchy          uint8
	AutoLife       uint8
	Unused1        uint8
	Unused2        uint8
	Unused3        uint8
	Unused4        uint8
	Unused5        uint8
}

// StatusMap2 (24 bytes): grupo 2 (shell..invincible).
type StatusMap2 struct {
	Shell                  uint8
	Protect                uint8
	Reflect                uint8
	Regen                  uint8
	Haste                  uint8
	Slow                   uint8
	Stop                   uint8
	StrengthBonus          uint8
	MagicBonus             uint8
	DefenseBonus           uint8
	MagicDefenseBonus      uint8
	AccuracyBonus          uint8
	EvasionBonus           uint8
	LuckBonus              uint8
	DoomCounter            uint8
	ImmunityPhysicalDamage uint8
	ImmunityMagicalDamage  uint8
	Invincible             uint8
	Unused1                uint8
	Unused2                uint8
	Unused3                uint8
	Unused4                uint8
	Unused5                uint8
	Unused6                uint8
}

// StatusDurationMap2 (24 bytes): duração (turnos) por status do grupo 2.
type StatusDurationMap2 struct {
	Shell                  int8
	Protect                int8
	Reflect                int8
	Regen                  int8
	Haste                  int8
	Slow                   int8
	Stop                   int8
	StrengthBonus          int8
	MagicBonus             int8
	DefenseBonus           int8
	MagicDefenseBonus      int8
	AccuracyBonus          int8
	EvasionBonus           int8
	LuckBonus              int8
	DoomCounter            int8
	ImmunityPhysicalDamage int8
	ImmunityMagicalDamage  int8
	Invincible             int8
	Unused1                int8
	Unused2                int8
	Unused3                int8
	Unused4                int8
	Unused5                int8
	Unused6                int8
}

// AutoAbilityEffectsMap (6 bytes): palavras de bits das auto-abilities
// (InlineArray(3) de ushort no C#).
type AutoAbilityEffectsMap struct {
	Words [3]uint16 // C# _u (palavra de bits; props has_*)
}

// StatChanges (10 bytes, src/core/ffx2/job.cs): deltas de stat.
type StatChanges struct {
	HP           int8
	MP           int8
	Strength     int8
	Defense      int8
	Magic        int8
	MagicDefense int8
	Agility      int8
	Accuracy     int8
	Evasion      int8
	Luck         int8
}

// UnlockableAbility (4 bytes, src/core/ffx2/job.cs): habilidade desbloqueável
// (requirement/ability são contextuais: gate de Garment Grid, AP, nível...).
type UnlockableAbility struct {
	Requirement uint16
	Ability     uint16
}
