package ffx2

// Last Mission (ffx_ps2/ffx2/master/jppc/lastmiss/kernel/*.h)
// (C#: src/core/ffx2/{LmCommand,LmMonMagic,LmMonster}.cs)
//
// Divergências PC vs struct C# (resolvidas empiricamente; preservadas como
// reserves nomeadas):
//   - LmCommand: C# termina em 0x43 (67B), binário PC = 68B (byte final = 0).
//   - LmMonMagic: C# termina em 0x45 (69B), binário PC = 72B (3 bytes finais).
//   - LmMonster: C# termina em 0x7B (123B), binário PC = 128B (flag@0x7B +
//     bitfield u32@0x7C sempre não-zero <=127) — struct C# é `partial`.
//
// Chaves = nomes dos campos do struct C# (ver "Convenções" do
// pacote): name@0x00, help@0x08, information@0x10 em LmCommand/LmMonMagic;
// name@0x00, help@0x04 em LmMonster.

// LmCommand (68 bytes). `name`/`help`/`information` são `uint` no C# — pares
// offset+chave de 4 bytes, i.e. TextRef.
type LmCommand struct {
	Name               TextRef
	NameYN             uint8
	Dummy1             uint8
	Dummy2             uint8
	Dummy3             uint8
	Help               TextRef
	HelpYN             uint8
	Dummy4             uint8
	Dummy5             uint8
	Dummy6             uint8
	Information        TextRef
	PersonalJob        uint16
	CharJob            uint8
	ShortDist          uint8
	ShotRange          uint8
	LongDist           uint8
	LongRange          uint8
	StairchkYN         uint8
	CursolDistYN       uint8
	CursolCat          uint8
	TargetPos          uint8
	Cmdend             uint8
	CmdendMenu         uint8
	Dummy7             uint8
	EffectNo           int16
	ReadMotion         uint8
	ReadEffect         uint8
	TrunChange         uint8
	Motion             uint8
	StealSt            uint8
	StealHit           uint8
	UseMP              uint8
	UseHP              uint8
	TargetParam        uint8
	RetdmgMotion       uint8
	Critical           uint8
	CalcID             uint8
	CalcNo             uint16
	AtkCnt             uint16
	TargetItemCategory uint16
	CategoryAbilty     uint8
	CategoryDmgRet     uint8
	Hit                uint8
	DarkHit            uint8
	HitCalcID          uint8
	BlueMagic          uint8
	AttributeAtk       uint8
	ConfUse            uint8
	Jibaku             uint8
	StatusChgTarget    uint8
	StatusChg          uint8
	StatusOnoff        uint8
	StatusHit          uint8
	Reserve1           uint8 // byte final do PC (0x43; sempre 0 no binário real)
}

// LmMonMagic (72 bytes) — mesma base de LmCommand com outra ordem interna
// (motion antes de trun_change, atk_no no lugar de atk_cnt, dummy8 +
// blue_magic u16 extras).
type LmMonMagic struct {
	Name               TextRef
	NameYN             uint8
	Dummy1             uint8
	Dummy2             uint8
	Dummy3             uint8
	Help               TextRef
	HelpYN             uint8
	Dummy4             uint8
	Dummy5             uint8
	Dummy6             uint8
	Information        TextRef
	PersonalJob        uint16
	CharJob            uint8
	ShortDist          uint8
	ShotRange          uint8
	LongDist           uint8
	LongRange          uint8
	StairchkYN         uint8
	CursolDistYN       uint8
	CursolCat          uint8
	TargetPos          uint8
	Cmdend             uint8
	CmdendMenu         uint8
	Dummy7             uint8
	EffectNo           int16
	ReadMotion         uint8
	ReadEffect         uint8
	Motion             uint8
	TrunChange         uint8
	StealSt            uint8
	StealHit           uint8
	UseMP              uint8
	UseHP              uint8
	TargetParam        uint8
	RetdmgMotion       uint8
	Critical           uint8
	CalcID             uint8
	CalcNo             uint16
	AtkNo              uint16
	TargetItemCategory uint16
	CategoryAbilty     uint8
	CategoryDmgRet     uint8
	Hit                uint8
	DarkHit            uint8
	HitCalcID          uint8
	Dummy8             uint8
	BlueMagic          uint16
	AttributeAtk       uint8
	ConfUse            uint8
	Jibaku             uint8
	StatusChgTarget    uint8
	StatusChg          uint8
	StatusOnoff        uint8
	StatusHit          uint8
	Reserve1           [3]uint8 // PC: C# termina em 0x45; binário = 0x48
}

// LmMonster (128 bytes). Os campos finais (ExtraFlag/ExtraFlags) não existem
// no struct parcial do C#; são lidos do binário PC (0x7B..0x80) para round-trip.
type LmMonster struct {
	Name                TextRef
	Help                TextRef
	Lv                  uint8
	Dummy1              uint8
	Dummy2              uint8
	Dummy3              uint8
	HP                  uint32
	MP                  uint32
	Str                 uint8
	Mag                 uint8
	Vit                 uint8
	Spirit              uint8
	Hit                 uint8
	Avoid               uint8
	Dummy4              uint8
	Dummy5              uint8
	OsHP                uint32
	OsMP                uint32
	OsStr               uint8
	OsMag               uint8
	OsVit               uint8
	OsSpirit            uint8
	OsHit               uint8
	OsAvoid             uint8
	Move                uint8
	FixDmg              uint8
	StairMove           uint8
	SizeSquare          uint8
	Dummy6              uint8
	Dummy7              uint8
	SizeReal            uint32
	ThinkMovepat        uint8
	Dummy8              uint8
	Dummy9              uint8
	Dummy10             uint8
	ViewDistNormal      uint32
	ViewRangeNormal     uint8
	Dummy11             uint8
	Dummy12             uint8
	Dummy13             uint8
	ViewDistBattle      uint32
	ViewRangeBattle     uint8
	ViewObstacle        uint8
	EffeZantetsu        uint8
	HitZantetsu         uint8
	HitCarryMon         uint8
	EleHoly             uint8
	EleGravit           uint8
	EleFire             uint8
	EleThunder          uint8
	EleIce              uint8
	EleWater            uint8
	Dummy14             uint8
	Item                uint16
	ItemData            uint8
	StealItemHit        uint8
	OsItem              uint16
	OsItemData          uint8
	OsStealItemHit      uint8
	StealMonSkill       int32 // alinhado em 0x58 (sem pad: 0x57 já termina par)
	StealMonSkillHit    uint8 // taxas em % (<=100)
	Dummy15             uint8
	Dummy16             uint8
	Dummy17             uint8
	Exp                 int32
	Ap                  uint8
	StealExpHit         uint8
	StealExpRate        uint8
	MgunAtk             uint8
	TypeSky             uint8
	ProhibitTimestop    uint8
	ProhibitSmellPlayer uint8
	ProhibitRatedmg     uint8
	ProhibitPosion      uint8
	ProhibitBlindness   uint8
	ProhibitSleep       uint8
	ProhibitConfusion   uint8
	ProhibitStop        uint8
	ProhibitDeadCount   uint8
	ProhibitSilence     uint8
	ProhibitBerserk     uint8
	ProhibitSlow        uint8
	ProhibitMovestop    uint8
	ProhibitConsump2MP  uint8
	ProhibitDropMoney   uint8
	ProhibitOversoul    uint8
	ProhibitChangeMoney uint8
	ProhibitChangeDress uint8
	ExtraFlag           uint8  // 0x7B: flag ausente no C# (98% dos monstros = 1)
	ExtraFlags          uint32 // 0x7C..0x80: bitfield ausente no C# (sempre 1..127)
}
