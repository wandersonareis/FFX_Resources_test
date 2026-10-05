package chunkmap

import "testing"

// TestSnake fixa a regra de chave: snake_case do identificador Go, com
// tratamento de corrida de maiúsculas (inicialismos Go) e de dígitos.
// É o que faz a chave bater com o nome do campo do struct C#.
func TestSnake(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"Name", "name"},
		{"Help", "help"},
		{"SensorText", "sensor_text"},
		{"EffectDescription", "effect_description"},
		{"CreatureData", "creature_data"},
		{"Messages", "messages"},
		// corrida de maiúsculas
		{"HPMax", "hp_max"},
		{"MPMax", "mp_max"},
		{"CostATB", "cost_atb"},
		{"CostMP", "cost_mp"},
		{"AP", "ap"},
		{"GetAP", "get_ap"},
		{"BTLSequence", "btl_sequence"},
		{"BtlSequence", "btl_sequence"},
		{"IsSOS", "is_sos"},
		// dígitos
		{"SubMenuCat2", "sub_menu_cat2"},
		{"OrderingIdx1", "ordering_idx1"},
		{"OrderingIdx2", "ordering_idx2"},
		{"StatusInflict1", "status_inflict1"},
		{"Reserve1", "reserve1"},
		{"Dummy10", "dummy10"},
		{"Anim1", "anim1"},
		{"ProhibitConsump2MP", "prohibit_consump_2mp"},
		{"Foo12Bar", "foo_12bar"},
		{"Dummy17", "dummy17"},
		// comuns
		{"DmgFormula", "dmg_formula"},
		{"MagicDefense", "magic_defense"},
		{"GilToSteal", "gil_to_steal"},
		{"AlwaysZero", "always_zero"},
		{"NameYN", "name_yn"},
		{"StairchkYN", "stairchk_yn"},
		{"", ""},
	} {
		if got := snake(c.in); got != c.want {
			t.Errorf("snake(%q) = %q, esperado %q", c.in, got, c.want)
		}
	}
}

// TestSnakeNaoQuebraPalavraCotidiana garante que a regra não insere `_` a
// mais em nomes simples do projeto.
func TestSnakeNaoQuebraPalavraCotidiana(t *testing.T) {
	for _, in := range []string{
		"Death", "Poison", "Reflect", "CommandList", "ForcedMove",
		"OverkillThreshold", "SpecialResistances", "CommandPdata",
	} {
		got := snake(in)
		if got == "" || got[0] == '_' || got[len(got)-1] == '_' {
			t.Errorf("snake(%q) = %q: underscore sobrando", in, got)
		}
		for i := 1; i < len(got); i++ {
			if got[i] == '_' && got[i-1] == '_' {
				t.Errorf("snake(%q) = %q: underscore duplo", in, got)
				break
			}
		}
	}
}
