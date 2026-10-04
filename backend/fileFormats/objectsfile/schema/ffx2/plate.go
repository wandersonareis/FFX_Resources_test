package ffx2

// plate.bin — ffx_ps2/ffx2/master/.../battle/kernel/plate.h (C#: src/core/ffx2/plate.cs)
// individual = 0x80 (128) — igual ao struct C# (Pack=4/Sequential).
//
// Chaves (nomes dos campos do struct C#, ver "Convenções" do
// pacote): name@0x00, help@0x04, messages_{0..3}@0x08..0x18,
// creature_data_help@0x48.

// PlateMessages (16 bytes) — InlineArray(4) de TextRef no C#
// (chaves messages_0..messages_3).
type PlateMessages [4]TextRef

// PlateCreatureData (0x38) — StructLayout Explicit, Size=0x38 no C#: o gap
// 0x0C..0x2C e o padding final 0x36..0x38 não têm campo no C#; aqui viram
// reserves nomeados para round-trip fiel.
type PlateCreatureData struct {
	Help        TextRef // C# `creature_data.help`
	Abilities   [2]UnlockableAbility
	Reserve1    [0x20]uint8 // gap 0x0C..0x2C não nomeado no C#
	StatChanges StatChanges
	Reserve2    [2]uint8 // padding final do Size=0x38 (0x36..0x38)
}

// Plate (128 bytes) — Sequential no C#.
type Plate struct {
	Name         TextRef
	Help         TextRef
	Messages     PlateMessages
	Bonus        uint16
	Icon         uint8
	StatChanges  StatChanges
	Reserve1     [3]uint8 // C# reserve1..reserve3 (privados)
	Skill        [8]UnlockableAbility
	CreatureData PlateCreatureData
}
