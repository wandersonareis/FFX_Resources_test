package duplicateFilesData

import "sync"

var (
	ffx2EventKlyt0900Map  DuplicateMap
	ffx2EventKlyt0900Once sync.Once
)

func (d *Ffx2DuplicateFileMap) GetFfx2EventKlyt0900() DuplicateMap {
	ffx2EventKlyt0900Once.Do(func() {
		ffx2EventKlyt0900Map = DuplicateMap{
			"klyt0900": []string{
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\kl\\klyt1700\\klyt1700.bin",
			},
		}
	})
	return ffx2EventKlyt0900Map
}
