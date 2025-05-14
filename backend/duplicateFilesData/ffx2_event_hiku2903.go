package duplicateFilesData

import "sync"

var (
	ffx2EventHiku2903Map  DuplicateMap
	ffx2EventHiku2903Once sync.Once
)

func (d *Ffx2DuplicateFileMap) GetFfx2EventHiku2903() DuplicateMap {
	ffx2EventHiku2903Once.Do(func() {
		ffx2EventHiku2903Map = DuplicateMap{
			"hiku2903": []string{
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\hi\\hiku3400\\hiku3400.bin",
			},
		}
	})
	return ffx2EventHiku2903Map
}
