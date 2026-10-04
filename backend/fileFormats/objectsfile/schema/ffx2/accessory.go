package ffx2

// accessory.bin — ffx2/master/jppc/battle/kernel/accessory.h
// (C#: src/core/ffx2/accessory.cs) individual = 0x54 (84) — igual ao C#.
//
// Chaves (nomes dos campos do struct C#, ver "Convenções" do
// pacote): name@0x00, help@0x04, creature_data_help@0x24.

// FeedStatChanges (16 bytes) — hp/mp int + 8 stats byte no C#.
type FeedStatChanges struct {
	HP           int32
	MP           int32
	Strength     uint8
	Defense      uint8
	Magic        uint8
	MagicDefense uint8
	Agility      uint8
	Luck         uint8
	Evasion      uint8
	Accuracy     uint8
}

// AccessoryCreatureData (0x30) — Explicit, Size=0x30 no C#: o gap
// 0x0C..0x1D não é nomeado no C#; aqui vira Reserve1.
type AccessoryCreatureData struct {
	Help           TextRef // C# `creature_data.help`
	Abilities      [2]UnlockableAbility
	Reserve1       [0x11]uint8 // gap 0x0C..0x1D não nomeado no C#
	FeedAmount     uint8
	AbilityToLearn uint16
	FeedStats      FeedStatChanges
}

// Accessory (84 bytes) — Sequential no C#.
type Accessory struct {
	Name         TextRef
	Help         TextRef
	ExtData      uint8
	Equip        uint8
	User         uint8
	Icon         uint8
	OrderingIdx  uint8
	Reserve      uint8 // C# `reserve` (privado)
	StatChanges  StatChanges
	Abilities    [2]UnlockableAbility
	Price        uint32
	CreatureData AccessoryCreatureData
}
