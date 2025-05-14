package duplicateFilesData

import "sync"

var (
  ffx2EventBsil0300Map  DuplicateMap
  ffx2EventBsil0300Once sync.Once
)

func (d *Ffx2DuplicateFileMap) GetFfx2EventBsil0300() DuplicateMap {
  ffx2EventBsil0300Once.Do(func() {
    ffx2EventBsil0300Map = DuplicateMap{
      "bsil0300": []string{
        "ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\bs\\bsil0500\\bsil0500.bin",
      },
    }
  })
  return ffx2EventBsil0300Map
}
