package chunkmap

import (
	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/objectsfile/schema/ffx2"
)

// registerLastMissSchemas registra todos os schemas chunkmap de LastMission.
// A chave do registro é "lastmiss/<patternPath>" e o patternPath já inclui o
// diretório lastmiss/kernel (os binários ficam sob a raiz de FFX-2).
func registerLastMissSchemas() {
	registerSchema("lastmiss/kernel/lm_accesary.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx2.LmAccesary](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.LmAccesary](v, "lastmiss/kernel/lm_accesary.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.LmAccesary](v, "lastmiss/kernel/lm_accesary.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.LmAccesary](v, "lastmiss/kernel/lm_accesary.bin", rows)
		})

	registerSchema("lastmiss/kernel/lm_command.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.LmCommand](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.LmCommand](v, "lastmiss/kernel/lm_command.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.LmCommand](v, "lastmiss/kernel/lm_command.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.LmCommand](v, "lastmiss/kernel/lm_command.bin", rows)
		})

	registerSchema("lastmiss/kernel/lm_monmagic.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx2.LmMonMagic](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.LmMonMagic](v, "lastmiss/kernel/lm_monmagic.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.LmMonMagic](v, "lastmiss/kernel/lm_monmagic.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.LmMonMagic](v, "lastmiss/kernel/lm_monmagic.bin", rows)
		})

	registerSchema("lastmiss/kernel/lm_monster.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.LmMonster](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.LmMonster](v, "lastmiss/kernel/lm_monster.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.LmMonster](v, "lastmiss/kernel/lm_monster.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.LmMonster](v, "lastmiss/kernel/lm_monster.bin", rows)
		})

	registerSchema("lastmiss/kernel/lm_item.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.LmItem](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.LmItem](v, "lastmiss/kernel/lm_item.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.LmItem](v, "lastmiss/kernel/lm_item.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.LmItem](v, "lastmiss/kernel/lm_item.bin", rows)
		})

	registerSchema("lastmiss/kernel/lm_dress.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.LmDress](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.LmDress](v, "lastmiss/kernel/lm_dress.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.LmDress](v, "lastmiss/kernel/lm_dress.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.LmDress](v, "lastmiss/kernel/lm_dress.bin", rows)
		})

	registerSchema("lastmiss/kernel/lm_trap.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.LmTrap](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.LmTrap](v, "lastmiss/kernel/lm_trap.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.LmTrap](v, "lastmiss/kernel/lm_trap.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.LmTrap](v, "lastmiss/kernel/lm_trap.bin", rows)
		})

	registerSchema("lastmiss/kernel/lm_mes.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.LmMes](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.LmMes](v, "lastmiss/kernel/lm_mes.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.LmMes](v, "lastmiss/kernel/lm_mes.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.LmMes](v, "lastmiss/kernel/lm_mes.bin", rows)
		})

	registerSchema("lastmiss/kernel/lm_player.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.LmPlayer](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.LmPlayer](v, "lastmiss/kernel/lm_player.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.LmPlayer](v, "lastmiss/kernel/lm_player.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.LmPlayer](v, "lastmiss/kernel/lm_player.bin", rows)
		})

	registerSchema("lastmiss/kernel/lm_warehouse.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx2.LmWarehouse](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.LmWarehouse](v, "lastmiss/kernel/lm_warehouse.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.LmWarehouse](v, "lastmiss/kernel/lm_warehouse.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.LmWarehouse](v, "lastmiss/kernel/lm_warehouse.bin", rows)
		})

	registerSchema("lastmiss/kernel/lm_floorname.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx2.LmFloorName](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.LmFloorName](v, "lastmiss/kernel/lm_floorname.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.LmFloorName](v, "lastmiss/kernel/lm_floorname.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.LmFloorName](v, "lastmiss/kernel/lm_floorname.bin", rows)
		})
}
