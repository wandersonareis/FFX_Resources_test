package ffx2

// a_ability.bin
// PC new_uspc: individual = 0xB0 (176);
//
// Chaves legadas (FileLayouts "ffx2/battle/kernel/a_ability.bin" =
// CommandV2Layout): name@0x00, description@0x04.
type AutoAbility struct {
	Name TextRef
	Help TextRef

	Reserve1 [4]int16

	CommandMenu       T_X2CommandId
	CastTimeReduction int8
	Reserve2          uint8

	Flags       AAbilityFlags
	ElemStrike  ElementFlags
	ElemAbsorb  ElementFlags
	ElemIgnore  ElementFlags
	ElemResist  ElementFlags
	ElemWeak    ElementFlags
	StatChanges StatChanges

	StatusInflict1 StatusMap
	StatusInflict2 StatusMap2
	StatusResist1  StatusMap
	StatusResist2  StatusMap2

	ReservePad uint8 // alinhamento do uint seguinte (offset 0x87) no layout PC

	StatusAuto1 StatusFlags

	StatusTime StatusDurationMap2
	Effects    AutoAbilityEffectsMap
	Icon       uint8
	Reserve3   uint8
	Reserve4   uint16
	AP         uint16
}
