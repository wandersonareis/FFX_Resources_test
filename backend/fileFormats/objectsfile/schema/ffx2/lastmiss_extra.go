package ffx2

// Last Mission — tabelas SEM struct C# correspondente
// (fahrenheit/src/core/ffx2 só tem LmCommand, LmMonMagic e LmMonster).
//
// Os offsets dos campos de texto vêm dos LayoutSet legados registrados em
// objectsfile.FileLayouts e foram confirmados byte a byte no binário
// new_uspc (Segment = par offset+chave, o segundo u16 nunca é zero; os
// valores restantes são codificados como reserves nomeados porque não há
// fonte C# para batizá-los).
//
// Chaves derivadas das chaves legadas com description->help:
//   - lm_accesary (52): name@0, help@4, effect@8, effect_description@12
//   - lm_dress    (40): name@0, help@4, effect@8
//   - lm_trap     (36): name@0, help@4, effect@8
//   - lm_item     (60): name@0, help@8, effect@16 (buracos 4..8 e 12..16)
//   - lm_mes       (8): name@0, help@4  (sem tail — 2 Segment = 8 bytes)
//   - lm_player   (60): name@0, help@4
//   - lm_warehouse(60): name@0, help@4
//   - lm_floorname (4): name@0          (sem tail — 1 Segment = 4 bytes)

// LmAccesary (52 bytes) — acessórios de Last Mission.
type LmAccesary struct {
	Name              TextRef
	Help              TextRef
	Effect            TextRef
	EffectDescription TextRef
	Reserve           [36]uint8 // 0x10..0x34 sem fonte C#
}

// LmDress (40 bytes) — vestes de Last Mission.
type LmDress struct {
	Name    TextRef
	Help    TextRef
	Effect  TextRef
	Reserve [28]uint8 // 0x0C..0x28 sem fonte C#
}

// LmTrap (36 bytes) — armadilhas de Last Mission.
type LmTrap struct {
	Name    TextRef
	Help    TextRef
	Effect  TextRef
	Reserve [24]uint8 // 0x0C..0x24 sem fonte C#
}

// LmItem (60 bytes) — itens de Last Mission. Os textos são espaçados de 8:
// existem quatro bytes não-texto entre name e help e entre help e effect.
type LmItem struct {
	Name     TextRef
	Reserve1 [4]uint8 // 0x04..0x08 não é par offset+chave
	Help     TextRef
	Reserve2 [4]uint8 // 0x0C..0x10 não é par offset+chave
	Effect   TextRef
	Reserve3 [40]uint8 // 0x14..0x3C sem fonte C#
}

// LmMes (8 bytes) — mensagens: ocupa exatamente os dois Segment, sem tail.
type LmMes struct {
	Name TextRef
	Help TextRef
}

// LmPlayer (60 bytes) — personagens jogáveis de Last Mission.
type LmPlayer struct {
	Name    TextRef
	Help    TextRef
	Reserve [52]uint8 // 0x08..0x3C sem fonte C#
}

// LmWarehouse (60 bytes) — depósito de Last Mission.
type LmWarehouse struct {
	Name    TextRef
	Help    TextRef
	Reserve [52]uint8 // 0x08..0x3C sem fonte C#
}

// LmFloorName (4 bytes) — nomes de andar: ocupa exatamente um Segment.
type LmFloorName struct {
	Name TextRef
}
