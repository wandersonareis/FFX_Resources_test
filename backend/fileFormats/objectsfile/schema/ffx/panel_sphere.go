package ffx

// panel.bin (nós da Sphere Grid) — src/core/ffx/panel.cs. individual = 24B.
// Chaves C#: name/name_simplified + help/help_simplified @0x00..0x10.
type SphereGridNodeType struct {
	Name         TextPair
	Help         TextPair
	SphereEffect SphereTargets
	AbilityID    T_XCommandId
	Amount       uint16
	IconID       uint16
}

// sphere.bin (esferas) — src/core/ffx/sphere.cs. individual = 16B.
// O único campo de texto do struct C# é `help` (2 chaves @0x00..0x08).
type Sphere struct {
	Help        TextPair
	Type        SphereBehavior
	Activates   SphereTargets
	Range       SphereRange
	SpecialRole uint8
	Reserve     uint16 // C# `_0x0E` (privado)
}
