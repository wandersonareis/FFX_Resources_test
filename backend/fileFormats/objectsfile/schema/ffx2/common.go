// Package ffx2 mapeia os binários de battle/kernel (FFX-2) e lastmiss/kernel
// (Last Mission) em structs Go, no estilo das structs [StructLayout] do
//
// Origem dos campos sem padding implícito —
// encoding/binary escreve apenas os campos exportados, contíguos):
//
// TextRef identifica explicitamente uma referência textual. Os nomes Go dos
// campos são os rótulos exibidos; índices de slot, não strings, identificam
// os textos durante a importação. Reserve*/Dummy* preservam bytes não textuais.
package ffx2

import "ffxresources/backend/models"

// TextRef: referência textual (offset na string table + u16 key).
type TextRef = models.TextRef

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
	Damage9999     uint8
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
type AutoAbilityEffectsMap struct {
	Words [3]uint16
}

// StatChanges (10 bytes): deltas de stat.
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

// UnlockableAbility (4 bytes): habilidade desbloqueável
// (requirement/ability são contextuais: gate de Garment Grid, AP, nível...).
type UnlockableAbility struct {
	Requirement uint16
	Ability     uint16
}
