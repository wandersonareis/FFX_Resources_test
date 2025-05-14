package duplicateFilesData

import "sync"

var (
	ffx2EventHiku3000Map  DuplicateMap
	ffx2EventHiku3000Once sync.Once
)

func (d *Ffx2DuplicateFileMap) GetFfx2EventHiku3000() DuplicateMap {
	ffx2EventHiku3000Once.Do(func() {
		ffx2EventHiku3000Map = DuplicateMap{
			"hiku3000": []string{
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\hi\\hiku3500\\hiku3500.bin",
			},
		}
	})
	return ffx2EventHiku3000Map
}
