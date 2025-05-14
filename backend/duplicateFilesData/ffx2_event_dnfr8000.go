package duplicateFilesData

import "sync"

var (
	ffx2EventDnfr8000Map  DuplicateMap
	ffx2EventDnfr8000Once sync.Once
)

func (d *Ffx2DuplicateFileMap) GetFfx2EventDnfr8000() DuplicateMap {
	ffx2EventDnfr8000Once.Do(func() {
		ffx2EventDnfr8000Map = DuplicateMap{
			"dnfr8000": []string{
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr8001\\dnfr8001.bin",
			},
		}
	})
	return ffx2EventDnfr8000Map
}
