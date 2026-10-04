package ffx

// *_txt.bin (tabelas de texto de batalha/menu) — src/core/ffx/text.cs.
// Chaves C#: o nome do campo no struct (ver "Convenções" do pacote).

// HelpText (8B) — btl_txt.bin (individual = 8).
// O struct C# tem um único campo `help` (par standard/simplified).
type HelpText struct {
	Help TextPair
}

// NameHelpText (16B) — arms/config/item/mmain/menu/summon/status/btlend/
// build/name/save_txt.bin (individual = 16).
// Os dois campos do struct C# são exportados: `command` @0x00..0x08 e
// `help` @0x08..0x10.
type NameHelpText struct {
	Command TextPair
	Help    TextPair
}
