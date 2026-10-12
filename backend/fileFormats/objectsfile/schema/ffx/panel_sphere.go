package ffx

// panel.bin (nós da Sphere Grid) — individual = 24B.
type SphereGridNodeType struct {
	Name         TextPair
	Help         TextPair
	SphereEffect SphereTargets
	AbilityID    T_XCommandId
	Amount       uint16
	IconID       uint16
}

// sphere.bin (esferas) — individual = 16B.
type Sphere struct {
	Help        TextPair
	Type        SphereBehavior
	Activates   SphereTargets
	Range       SphereRange
	SpecialRole uint8
	Reserve     uint16
}
