package duplicateFilesData

import "sync"

var (
  ffx2EventHiku2800Map  DuplicateMap
  ffx2EventHiku2800Once sync.Once
)

func (d *Ffx2DuplicateFileMap) GetFfx2EventHiku2800() DuplicateMap {
  ffx2EventHiku2800Once.Do(func() {
    ffx2EventHiku2800Map = DuplicateMap{
      "hiku2800": []string{
        "ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\hi\\hiku3200\\hiku3200.bin",
      },
    }
  })
  return ffx2EventHiku2800Map
}
