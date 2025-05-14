package duplicateFilesData

import "sync"

var (
	ffx2EventKlyt0500Map  DuplicateMap
	ffx2EventKlyt0500Once sync.Once
)

func (d *Ffx2DuplicateFileMap) GetFfx2EventKlyt0500() DuplicateMap {
	ffx2EventKlyt0500Once.Do(func() {
		ffx2EventKlyt0500Map = DuplicateMap{
			"klyt0500": []string{
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\kl\\klyt1600\\klyt1600.bin",
			},
		}
	})
	return ffx2EventKlyt0500Map
}
