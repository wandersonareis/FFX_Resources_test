package ffx2

// a_ability.bin — ffx_ps2/ffx2/master/.../battle/kernel/a_ability.h
// (C#: src/core/ffx2/aability.cs). PC new_uspc: individual = 0xB0 (176);
// o struct do Switch (C#) tem 180 — o PC omite status_auto2 (4 bytes).
//
// Chaves legadas (FileLayouts "ffx2/battle/kernel/a_ability.bin" =
// CommandV2Layout): name@0x00, description@0x04.
type AutoAbility struct {
	Name TextRef
	Help TextRef

	Reserve1 [4]int16 // C# InlineArray4<short> (privado)

	CommandMenu       T_X2CommandId
	CastTimeReduction int8 // C# sbyte (percentual)
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

	// PC guarda só 4 bytes: o C# Switch tem ainda status_auto2 (StatusFlags2).
	StatusAuto1 StatusFlags

	StatusTime StatusDurationMap2
	Effects    AutoAbilityEffectsMap
	Icon       uint8
	Reserve3   uint8
	Reserve4   uint16
	AP         uint16
}
