package ffx2

// command.bin / item.bin / monmagic.bin — ffx_ps2/ffx2/master/jppc/
//
// Tamanhos do binário PC (new_uspc):
//
//	command.bin  = 140 = Command (134) + PCommandData (6)
//	item.bin     = 160 = Command (134) + ItemData (26)
//	monmagic.bin = 136 = Command (134) + MCommandData (2)
//
// Command sem padding final ocupa 0x86 (134) bytes: o structé
// Sequential e termina em get_ap@0x85, mas o PC não grava os 2 bytes de
// alinhamento — por isso os três tails começam logo em 0x86 e as
// contas fecham exatas.
//
// Chaves: name@0x00, help@0x04 (CommandV2Layout legado).
type (
	MenuFlags         uint8
	TargetFlags       uint32
	MiscFlags         uint32
	DamageFlags       uint32
	PartyPreviewFlags uint16
	DamageClass       uint8  // enum : byte
	FiendSpecies      uint16 // enum : ushort — espécies com dano dobrado
	SpecialImmunities uint16 // enum : ushort
)

type T_X2JobId = uint16

// Command (134 bytes).
type Command struct {
	Name TextRef
	Help TextRef

	Anim1        uint16
	Anim2        uint16
	CasterAnim   uint8
	FlagsMenu    MenuFlags
	SubMenuCat2  uint8
	SubMenuCat   uint8
	FlagsTarget  TargetFlags
	FlagsMisc    MiscFlags
	Reserve1     uint32
	FlagsDamage  DamageFlags
	PartyPreview PartyPreviewFlags
	CostATB      uint16
	CostCast     uint16
	CostMP       uint8

	DamageClass   DamageClass
	DmgFormula    uint8
	CritBonus     uint8
	Accuracy      uint8
	Power         uint8
	HitCount      uint8
	ShatterChance uint8
	Element       ElementFlags

	StatusInflict1 StatusMap
	StatusInflict2 StatusMap2
	StatusTime     StatusDurationMap2

	Icon                 uint8
	SpeciesEffectiveness FiendSpecies // espécie alvo do dano dobrado
	MagicCancel          uint8
	OrderingIdx1         uint8
	BlueBullet           T_X2CommandId
	OrderingIdx2         uint16

	Reserve2    uint32 // "cast animation"?
	BTLSequence uint8
	GetAP       uint8
}

// PCommandData (6 bytes) — tail de command.bin
type PCommandData struct {
	AP      uint16
	JobUse  T_X2JobId
	Reserve uint16
}

// PCommand (140 bytes) — command.bin.
type PCommand struct {
	Command
	CommandPdata PCommandData
}

// MCommandData (2 bytes) — tail de monmagic.bin.
type MCommandData struct {
	AP uint16
}

// MCommand (136 bytes) — monmagic.bin.
type MCommand struct {
	Command
	CommandMdata MCommandData
}

// ItemData (26 bytes) — tail de item.bin. Sem padding em Go os offsets
// relativos são exatamente os do arquivo (o campo `price` cai em 0x02 porque
// a base 0x86 ≡ 2 mod 4).
type ItemData struct {
	Element        uint8
	Level          uint8
	Price          uint32
	Reserve        uint8
	FeedAmount     uint8
	AbilityToLearn uint16
	FeedStats      FeedStatChanges
}

// Item (160 bytes) — item.bin.
type Item struct {
	Command
	ItemData ItemData
}
