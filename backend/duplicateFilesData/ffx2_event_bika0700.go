package duplicateFilesData

import "sync"

var (
  ffx2EventBika0700Map  DuplicateMap
  ffx2EventBika0700Once sync.Once
)

func (d *Ffx2DuplicateFileMap) GetFfx2EventBika0700() DuplicateMap {
  ffx2EventBika0700Once.Do(func() {
    ffx2EventBika0700Map = DuplicateMap{
      "bika0700": []string{
        "ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\bi\\bika1100\\bika1100.bin",
      },
    }
  })
  return ffx2EventBika0700Map
}
