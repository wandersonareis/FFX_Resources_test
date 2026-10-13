package chunkmap

import (
	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/objectsfile/schema/ffx2"
)

// registerFFX2Schemas registra todos os schemas chunkmap de FFX-2. A chave do
// registro é "ffx2/<patternPath>" e o patternPath aponta para battle/kernel.
func registerFFX2Schemas() {
	registerSchema("ffx2/battle/kernel/command.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.PCommand](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.PCommand](v, "battle/kernel/command.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.PCommand](v, "battle/kernel/command.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.PCommand](v, "battle/kernel/command.bin", rows)
		})

	registerSchema("ffx2/battle/kernel/item.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.Item](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.Item](v, "battle/kernel/item.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.Item](v, "battle/kernel/item.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.Item](v, "battle/kernel/item.bin", rows)
		})

	registerSchema("ffx2/battle/kernel/important.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.KeyItem](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.KeyItem](v, "battle/kernel/important.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.KeyItem](v, "battle/kernel/important.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.KeyItem](v, "battle/kernel/important.bin", rows)
		})

	registerSchema("ffx2/battle/kernel/accessory.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.Accessory](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.Accessory](v, "battle/kernel/accessory.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.Accessory](v, "battle/kernel/accessory.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.Accessory](v, "battle/kernel/accessory.bin", rows)
		})

	registerSchema("ffx2/battle/kernel/job.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.Job](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.Job](v, "battle/kernel/job.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.Job](v, "battle/kernel/job.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.Job](v, "battle/kernel/job.bin", rows)
		})

	registerSchema("ffx2/battle/kernel/plate.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.Plate](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.Plate](v, "battle/kernel/plate.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.Plate](v, "battle/kernel/plate.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.Plate](v, "battle/kernel/plate.bin", rows)
		})

	registerSchema("ffx2/battle/kernel/a_ability.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx2.AutoAbility](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.AutoAbility](v, "battle/kernel/a_ability.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.AutoAbility](v, "battle/kernel/a_ability.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.AutoAbility](v, "battle/kernel/a_ability.bin", rows)
		})

	registerSchema("ffx2/battle/kernel/menu_txt.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.MenuTxt](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.MenuTxt](v, "battle/kernel/menu_txt.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.MenuTxt](v, "battle/kernel/menu_txt.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.MenuTxt](v, "battle/kernel/menu_txt.bin", rows)
		})

	registerSchema("ffx2/battle/kernel/btlend_txt.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.BtlEndTxt](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.BtlEndTxt](v, "battle/kernel/btlend_txt.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.BtlEndTxt](v, "battle/kernel/btlend_txt.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.BtlEndTxt](v, "battle/kernel/btlend_txt.bin", rows)
		})

	registerSchema("ffx2/battle/kernel/monster.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.Monster](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.Monster](v, "battle/kernel/monster.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.Monster](v, "battle/kernel/monster.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.Monster](v, "battle/kernel/monster.bin", rows)
		})

	registerSchema("ffx2/battle/kernel/monster2.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.Monster2](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.Monster2](v, "battle/kernel/monster2.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.Monster2](v, "battle/kernel/monster2.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.Monster2](v, "battle/kernel/monster2.bin", rows)
		})

	registerSchema("ffx2/battle/kernel/oversoul.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.Oversoul](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.Oversoul](v, "battle/kernel/oversoul.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.Oversoul](v, "battle/kernel/oversoul.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.Oversoul](v, "battle/kernel/oversoul.bin", rows)
		})

	registerSchema("ffx2/battle/kernel/monmagic.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.MCommand](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.MCommand](v, "battle/kernel/monmagic.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.MCommand](v, "battle/kernel/monmagic.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.MCommand](v, "battle/kernel/monmagic.bin", rows)
		})

	registerSchema("ffx2/battle/kernel/btl_txt.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx2.BtlTxt](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx2.BtlTxt](v, "battle/kernel/btl_txt.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx2.BtlTxt](v, "battle/kernel/btl_txt.bin", e)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx2.BtlTxt](v, "battle/kernel/btl_txt.bin", rows)
		})
}
