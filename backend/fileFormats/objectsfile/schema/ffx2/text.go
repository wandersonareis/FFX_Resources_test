package ffx2

// *_txt.bin e oversoul.bin (tabelas de texto e tabela de Oversoul)
// Oversoul (8B) — oversoul.bin (individual = 8).
type Oversoul struct {
	Name  TextRef
	Count [2]int16
}

// BtlTxt (4B) — btl_txt.bin (individual = 4).
type BtlTxt struct {
	Help TextRef
}

// BtlEndTxt (8B) — btlend_txt.bin (individual = 8).
type BtlEndTxt struct {
	Command TextRef
	Help    TextRef
}

// MenuTxt (8B) — menu_txt.bin (individual = 8).
type MenuTxt struct {
	Command TextRef
	Help    TextRef
}
