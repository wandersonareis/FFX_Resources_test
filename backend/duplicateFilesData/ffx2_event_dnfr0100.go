package duplicateFilesData

import "sync"

var (
	ffx2EventDnfr0100Map  DuplicateMap
	ffx2EventDnfr0100Once sync.Once
)

func (d *Ffx2DuplicateFileMap) GetFfx2EventDnfr0100() DuplicateMap {
	ffx2EventDnfr0100Once.Do(func() {
		ffx2EventDnfr0100Map = DuplicateMap{
			"dnfr0100": []string{
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr0300\\dnfr0300.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr0500\\dnfr0500.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr0700\\dnfr0700.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr0900\\dnfr0900.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr1100\\dnfr1100.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr1300\\dnfr1300.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr1500\\dnfr1500.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr1700\\dnfr1700.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr1800\\dnfr1800.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr2000\\dnfr2000.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr2100\\dnfr2100.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr2300\\dnfr2300.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr2500\\dnfr2500.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr2700\\dnfr2700.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr2900\\dnfr2900.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr3100\\dnfr3100.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr3300\\dnfr3300.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr3500\\dnfr3500.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr3700\\dnfr3700.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr4000\\dnfr4000.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr4100\\dnfr4100.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr4300\\dnfr4300.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr4500\\dnfr4500.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr4700\\dnfr4700.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr4900\\dnfr4900.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr5100\\dnfr5100.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr5300\\dnfr5300.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr5500\\dnfr5500.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr5700\\dnfr5700.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr5800\\dnfr5800.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr6000\\dnfr6000.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr6100\\dnfr6100.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr6300\\dnfr6300.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr6500\\dnfr6500.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr6700\\dnfr6700.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr6800\\dnfr6800.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr6900\\dnfr6900.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr7000\\dnfr7000.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr7100\\dnfr7100.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr7200\\dnfr7200.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr7300\\dnfr7300.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr7500\\dnfr7500.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr7600\\dnfr7600.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr7700\\dnfr7700.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr7800\\dnfr7800.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\dn\\dnfr7900\\dnfr7900.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\lm\\lmys0000\\lmys0000.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\sp\\spdn0100\\spdn0100.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\sp\\spdn0400\\spdn0400.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\sp\\spdn0401\\spdn0401.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\sp\\spdn0402\\spdn0402.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\sp\\spdn0403\\spdn0403.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\sp\\spdn0500\\spdn0500.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\sp\\spdn0501\\spdn0501.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\sp\\spdn0502\\spdn0502.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\sp\\spdn0503\\spdn0503.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\sp\\spdn0600\\spdn0600.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\sp\\spdn0601\\spdn0601.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\sp\\spdn0602\\spdn0602.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\sp\\spdn0603\\spdn0603.bin",
				"ffx_ps2\\ffx2\\master\\new_uspc\\event\\obj_ps3\\sp\\spdn0700\\spdn0700.bin",
			},
		}
	})
	return ffx2EventDnfr0100Map
}
