package ffx2

// accessory.bin
// FeedStatChanges (16 bytes) — hp/mp int + 8 stats byte.
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

// AccessoryCreatureData (0x30) — Explicit, Size=0x30 : o gap
// 0x0C..0x1D não é nomeado; aqui vira Reserve1.
type AccessoryCreatureData struct {
	Help           TextRef
	Abilities      [2]UnlockableAbility
	Reserve1       [0x11]uint8
	FeedAmount     uint8
	AbilityToLearn uint16
	FeedStats      FeedStatChanges
}

// Accessory (84 bytes)
type Accessory struct {
	Name         TextRef
	Help         TextRef
	ExtData      uint8
	Equip        uint8
	User         uint8
	Icon         uint8
	OrderingIdx  uint8
	Reserve      uint8
	StatChanges  StatChanges
	Abilities    [2]UnlockableAbility
	Price        uint32
	CreatureData AccessoryCreatureData
}
