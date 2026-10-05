package ffx2

// monster.bin / monster2.bin — ffx_ps2/ffx2/master/jppc/battle/kernel/
// {monster,monster2}.h (C#: src/core/ffx2/{monster,Monster2}.cs).
// individual = 0xC8 (200) nos dois, igual aos structs C#.

// ChrItemLoot (24 bytes) — 6 pares (item/amount) comuns+raros de
// drop/steal/bribe (src/core/ffx2/battle/btldrop.cs).
type ChrItemLoot struct {
	ItemCommon        uint16
	ItemCommonAmount  uint16
	ItemRare          uint16
	ItemRareAmount    uint16
	StealCommon       uint16
	StealCommonAmount uint16
	StealRare         uint16
	StealRareAmount   uint16
	BribeCommon       uint16
	BribeCommonAmount uint16
	BribeRare         uint16
	BribeRareAmount   uint16
}

// ChrLoot (40 bytes) — recompensas do monstro.
type ChrLoot struct {
	Exp         int32
	Gil         int32
	GilToSteal  int32
	AP          uint16
	DropChance  uint8
	StealChance uint8
	Loot        ChrItemLoot
}

// Monster (200 bytes) — src/core/ffx2/monster.cs.
// Chaves: name@0x00, help@0x04 (NameOnlyV2Layout legado lê só
// name@0x00).
type Monster struct {
	Name TextRef
	Help TextRef

	MaxHP uint32
	MaxMP uint32

	Level        uint8
	Strength     uint8
	Defense      uint8
	Magic        uint8
	MagicDefense uint8
	Agility      uint8
	Accuracy     uint8
	Evasion      uint8
	Luck         uint8
	ThinkingTime uint8

	SpecialImmunities SpecialImmunities

	ElemAbsorb ElementFlags
	ElemIgnore ElementFlags
	ElemResist ElementFlags
	ElemWeak   ElementFlags

	StatusResist1 StatusMap
	StatusResist2 StatusMap2
	StatusAuto1   StatusFlags
	StatusAuto2   StatusFlags2
	StatusTime    StatusDurationMap2

	Abilities     [16]T_X2CommandId
	BerserkAction T_X2CommandId

	Model  uint16
	Motion uint16
	Sound  uint16

	Oversoul    uint16
	MonsterType FiendSpecies // espécie alvo do dano dobrado
	Loot        ChrLoot

	ZantetsuDefense uint8
	Reserve1        uint8
	Reserve2        uint16
}

// Monster2 (200 bytes) — src/core/ffx2/Monster2.cs (struct `partial`,
// descompilado: os buffers são InlineArray(24/24/24/16/4/4/4)).
// `name`/`help` são `uint` no C#, mas são pares offset+chave (Segment).
type Monster2 struct {
	Name TextRef
	Help TextRef

	HPMax uint32
	MPMax uint32

	Level        uint8
	Str          uint8
	Vit          uint8
	Mag          uint8
	Spirit       uint8
	Dex          uint8
	Hit          uint8
	Avoid        uint8
	Luck         uint8
	ThinkingTime uint8

	Special uint16

	AbsElement  uint8
	InvElement  uint8
	HalfElement uint8
	WeakElement uint8

	DefStatus  [24]uint8
	DefStatus2 [24]uint8

	AutoStatus  int32
	AutoStatus2 int32

	StatusTime [24]int8

	Waza [16]uint16

	Basaku       uint16
	MonModel     uint16
	MonMotion    uint16
	MonsterSound uint16
	Oversoul     uint16
	MonsterType  uint16

	Exp         int32
	Gill        int32
	StealGill   int32
	GetAP       uint16
	Drop        uint8
	Steal       uint8
	DropItem    [4]uint16
	StealItem   [4]uint16
	BriberyItem [4]uint16

	DefZantetu uint8
	Reserve1   uint8
	Reserve2   uint16
}
