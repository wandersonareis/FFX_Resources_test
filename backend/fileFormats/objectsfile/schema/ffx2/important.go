package ffx2

// important.bin (itens-chave) — ffx_ps2/ffx2/master/jppc/battle/kernel/
// KeyItem (12 bytes).
type KeyItem struct {
	Name      TextRef
	Help      TextRef
	ItemType  uint8
	ItemValue uint8
	Icon      uint8
	Number    uint8
}
