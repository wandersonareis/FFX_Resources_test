package duplicateFilesData

import "sync"

var (
	ffx2EventStbv1200Map  DuplicateMap
	ffx2EventStbv1200Once sync.Once
)

func (d *Ffx2DuplicateFileMap) GetFfx2EventStbv1200() DuplicateMap {
	ffx2EventStbv1200Once.Do(func() {
		ffx2EventStbv1200Map = DuplicateMap{
			"stbv1200": []string{
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\st\\stbv1300\\stbv1300.bin",
			},
		}
	})
	return ffx2EventStbv1200Map
}
