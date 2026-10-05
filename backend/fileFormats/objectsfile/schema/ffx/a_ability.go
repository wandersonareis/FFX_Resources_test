package ffx

// a_ability.bin (auto-abilities FFX) — src/core/ffx/aability.cs.
// individual = 108B no PC new_uspc — layout exato da struct C# (Switch HD):
// name(8) + desc(8) + is_sos(1) + 5×ElementFlags(5) + StatusMap(25) +
// StatusDurationMap(13) + StatusMap(25) + amount(1) + StatIncreaseFlags(2) +
// 5×flags u16 (10) + effects(6) + icon/group_idx/group_level/intl (4) = 108.
//
// As chaves espelham os campos do struct C#: `name` e `desc`
// (ver "Convenções" no doc do pacote).
type AutoAbility struct {
	Name TextPair
	Desc TextPair // C# `desc`

	IsSOS uint8 // C# bool — 7% dos chunks = 1

	ElemStrike ElementFlags
	ElemAbsorb ElementFlags
	ElemIgnore ElementFlags
	ElemResist ElementFlags
	ElemWeak   ElementFlags

	StatusInflict  StatusMap         // 25B @0x16
	StatusDuration StatusDurationMap // 13B @0x2F
	StatusResist   StatusMap         // 25B @0x3C

	StatIncAmount uint8             // @0x55 (18% dos chunks, <=30)
	StatIncFlags  StatIncreaseFlags // u16 @0x56 (byte alto = bits HP/MP/bônus)

	StatusAutoPermanent StatusPermanentFlags // u16 @0x58
	StatusAutoTemporal  StatusTemporalFlags  // u16 @0x5A
	StatusAutoExtra     StatusExtraFlags     // u16 @0x5C
	StatusInflictExtra  StatusExtraFlags     // u16 @0x5E
	StatusResistExtra   StatusExtraFlags     // u16 @0x60

	AutoAbilityEffects AutoAbilityEffectsMap // 6B @0x62..0x68

	Icon                  uint8 // @0x68 (= 0x14 no chunk0: índice no menu)
	GroupIdx              uint8 // @0x69 (= 1)
	GroupLevel            uint8 // @0x6A (<= 4)
	InternationalBonusIdx uint8 // @0x6B (0 em todos os chunks do PC)
}
