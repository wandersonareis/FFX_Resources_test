package ffx

// command.bin / item.bin / monmagic{1|2}.bin — src/core/ffx/command.cs.
//
// O PC new_uspc tem três tamanhos de chunk distintos (test/core/excel.cs):
//
//	SizeOf<PCommandData>() == 0x04
//	SizeOf<Command>()      == 0x5C (92)  → monmagic{1|2}.bin
//	SizeOf<PCommand>()     == 0x60 (96)  → {item|command}.bin
//
// PCommand = Command + PCommandData (4 bytes no fim do chunk).
//
// As chaves são os nomes dos campos do struct C#, em snake_case
// (ver "Convenções" no doc do pacote): name/name_simplified,
// desc/desc_simplified, help/help_simplified.
//
// Alinhamento validado por stats (320 chunks de command.bin): steals_gil 0/1
// @0x21 (2 comandos), mp_cost <=99 @0x25, JINX bit5 @0x56, overdrive_200 bit6
// @0x5A, 0x5D..0x5F sempre zero (PCommandData._0x02 + sphere_grid_role).

// CommandBody (76 bytes @0x10..0x5C) — trecho comum a Command e PCommand:
// animações, flags de menu/dano, custos, mapa de status e buffs.
type CommandBody struct {
	Anim1      uint16 // C# anim_1 (snake_case geraria anim1)
	Anim2      uint16 // C# anim_2 (snake_case geraria anim2)
	Icon       uint8
	CasterAnim uint8

	FlagsMenu    uint8
	SubMenuCat2  uint8
	SubMenuCat   uint8
	UserID       uint8 // 0xFF = nenhum usuário
	FlagsTarget  uint8
	FlagsUsage   uint8
	FlagsMisc    uint32
	FlagsDamage  uint8
	StealsGil    uint8 // C# bool — só Steal/Mug (2 comandos)
	PartyPreview uint8

	FlagsDamageClass uint8
	CtbRank          uint8
	MpCost           uint8
	LimitCost        uint8

	CritBonus     uint8
	DmgFormula    uint8
	Accuracy      uint8 // sempre 0 no PC
	Power         uint8
	HitCount      uint8
	ShatterChance uint8

	Element ElementFlags

	StatusMap         StatusMap         // 25 bytes @0x2E
	StatusDurationMap StatusDurationMap // 13 bytes @0x47
	FlagsStatusExtra  StatusExtraFlags  // u16 @0x54 (bit14 = DOOM)

	FlagsBuffsStat    uint16 // u16 @0x56 (bit5 = JINX)
	OverdriveCategory uint8  // @0x58
	BuffAmount        uint8  // @0x59
	FlagsBuffsMix     uint16 // u16 @0x5A (bit6 = overdrive_200)
}

// MonMagic (92 bytes) = C# FFX.Command — monmagic1/2.bin.
type MonMagic struct {
	Name TextPair
	Desc TextPair
	CommandBody
}

// Command (96 bytes) = C# FFX.PCommand — command.bin/item.bin.
type Command struct {
	Name TextPair
	Desc TextPair // C# `desc`
	CommandBody

	// PCommandData @0x5C..0x60 (src/core/ffx/command.cs).
	OrderingIdx    uint8  // índice de ordenação no menu (1..23)
	SphereGridRole uint8  // sempre 0 no PC
	ReserveData    uint16 // C# _0x02 (reserva do layout)
}

// important.bin (itens-chave) — src/core/ffx/important.cs. individual = 20B.
// Mesmo formato de command.bin (4 chaves @0x00..0x10); sem PCommandData.
type KeyItem struct {
	Name      TextPair
	Help      TextPair // C# `help`
	ItemType  uint8
	ItemValue uint8
	Icon      uint8
	Number    uint8
}
