package ffx2

// important.bin (itens-chave) — ffx_ps2/ffx2/master/jppc/battle/kernel/
// important.h (C#: src/core/ffx2/important.cs). individual = 12B — igual ao
// struct C# (CommandV2Layout legado: name@0x00, description@0x04).

// KeyItem (12 bytes).
type KeyItem struct {
	Name      TextRef
	Help      TextRef
	ItemType  uint8
	ItemValue uint8
	Icon      uint8
	Number    uint8
}
