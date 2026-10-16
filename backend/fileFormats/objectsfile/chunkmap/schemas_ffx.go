package chunkmap

import (
	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/objectsfile/schema/ffx"
)

// registerFFXSchemas registra todos os schemas chunkmap de FFX. A chave do
// registro é "ffx/<patternPath>" e o patternPath aponta para battle/kernel.
func registerFFXSchemas() {
	registerSchema("ffx/battle/kernel/command.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx.Command](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.Command](v, "battle/kernel/command.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.Command](v, "battle/kernel/command.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.Command](v, "battle/kernel/command.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.Command](v, "battle/kernel/command.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.Command](v, "battle/kernel/command.bin", rows)
		})

	registerSchema("ffx/battle/kernel/item.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx.Command](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.Command](v, "battle/kernel/item.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.Command](v, "battle/kernel/item.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.Command](v, "battle/kernel/item.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.Command](v, "battle/kernel/item.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.Command](v, "battle/kernel/item.bin", rows)
		})

	registerSchema("ffx/battle/kernel/important.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx.KeyItem](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.KeyItem](v, "battle/kernel/important.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.KeyItem](v, "battle/kernel/important.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.KeyItem](v, "battle/kernel/important.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.KeyItem](v, "battle/kernel/important.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.KeyItem](v, "battle/kernel/important.bin", rows)
		})

	registerSchema("ffx/battle/kernel/a_ability.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx.AutoAbility](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.AutoAbility](v, "battle/kernel/a_ability.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.AutoAbility](v, "battle/kernel/a_ability.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.AutoAbility](v, "battle/kernel/a_ability.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.AutoAbility](v, "battle/kernel/a_ability.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.AutoAbility](v, "battle/kernel/a_ability.bin", rows)
		})

	registerSchema("ffx/battle/kernel/panel.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx.SphereGridNodeType](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.SphereGridNodeType](v, "battle/kernel/panel.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.SphereGridNodeType](v, "battle/kernel/panel.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.SphereGridNodeType](v, "battle/kernel/panel.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.SphereGridNodeType](v, "battle/kernel/panel.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.SphereGridNodeType](v, "battle/kernel/panel.bin", rows)
		})

	registerSchema("ffx/battle/kernel/sphere.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx.Sphere](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.Sphere](v, "battle/kernel/sphere.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.Sphere](v, "battle/kernel/sphere.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.Sphere](v, "battle/kernel/sphere.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.Sphere](v, "battle/kernel/sphere.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.Sphere](v, "battle/kernel/sphere.bin", rows)
		})

	registerSchema("ffx/battle/kernel/monmagic1.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx.MonMagic](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.MonMagic](v, "battle/kernel/monmagic1.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.MonMagic](v, "battle/kernel/monmagic1.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.MonMagic](v, "battle/kernel/monmagic1.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.MonMagic](v, "battle/kernel/monmagic1.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.MonMagic](v, "battle/kernel/monmagic1.bin", rows)
		})

	registerSchema("ffx/battle/kernel/monmagic2.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx.MonMagic](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.MonMagic](v, "battle/kernel/monmagic2.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.MonMagic](v, "battle/kernel/monmagic2.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.MonMagic](v, "battle/kernel/monmagic2.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.MonMagic](v, "battle/kernel/monmagic2.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.MonMagic](v, "battle/kernel/monmagic2.bin", rows)
		})

	registerSchema("ffx/battle/kernel/monster1.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx.MonStats](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.MonStats](v, "battle/kernel/monster1.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.MonStats](v, "battle/kernel/monster1.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.MonStats](v, "battle/kernel/monster1.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.MonStats](v, "battle/kernel/monster1.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.MonStats](v, "battle/kernel/monster1.bin", rows)
		})

	registerSchema("ffx/battle/kernel/monster2.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx.MonStats](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.MonStats](v, "battle/kernel/monster2.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.MonStats](v, "battle/kernel/monster2.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.MonStats](v, "battle/kernel/monster2.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.MonStats](v, "battle/kernel/monster2.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.MonStats](v, "battle/kernel/monster2.bin", rows)
		})

	registerSchema("ffx/battle/kernel/monster3.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx.MonStats](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.MonStats](v, "battle/kernel/monster3.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.MonStats](v, "battle/kernel/monster3.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.MonStats](v, "battle/kernel/monster3.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.MonStats](v, "battle/kernel/monster3.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.MonStats](v, "battle/kernel/monster3.bin", rows)
		})

	registerSchema("ffx/battle/kernel/btl_txt.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx.HelpText](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.HelpText](v, "battle/kernel/btl_txt.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.HelpText](v, "battle/kernel/btl_txt.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.HelpText](v, "battle/kernel/btl_txt.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.HelpText](v, "battle/kernel/btl_txt.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.HelpText](v, "battle/kernel/btl_txt.bin", rows)
		})

	registerSchema("ffx/battle/kernel/arms_txt.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx.NameHelpText](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.NameHelpText](v, "battle/kernel/arms_txt.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.NameHelpText](v, "battle/kernel/arms_txt.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.NameHelpText](v, "battle/kernel/arms_txt.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.NameHelpText](v, "battle/kernel/arms_txt.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.NameHelpText](v, "battle/kernel/arms_txt.bin", rows)
		})

	registerSchema("ffx/battle/kernel/btlend_txt.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx.NameHelpText](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.NameHelpText](v, "battle/kernel/btlend_txt.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.NameHelpText](v, "battle/kernel/btlend_txt.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.NameHelpText](v, "battle/kernel/btlend_txt.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.NameHelpText](v, "battle/kernel/btlend_txt.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.NameHelpText](v, "battle/kernel/btlend_txt.bin", rows)
		})

	registerSchema("ffx/battle/kernel/build_txt.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx.NameHelpText](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.NameHelpText](v, "battle/kernel/build_txt.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.NameHelpText](v, "battle/kernel/build_txt.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.NameHelpText](v, "battle/kernel/build_txt.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.NameHelpText](v, "battle/kernel/build_txt.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.NameHelpText](v, "battle/kernel/build_txt.bin", rows)
		})

	registerSchema("ffx/battle/kernel/config_txt.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx.NameHelpText](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.NameHelpText](v, "battle/kernel/config_txt.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.NameHelpText](v, "battle/kernel/config_txt.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.NameHelpText](v, "battle/kernel/config_txt.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.NameHelpText](v, "battle/kernel/config_txt.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.NameHelpText](v, "battle/kernel/config_txt.bin", rows)
		})

	registerSchema("ffx/battle/kernel/item_txt.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx.NameHelpText](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.NameHelpText](v, "battle/kernel/item_txt.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.NameHelpText](v, "battle/kernel/item_txt.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.NameHelpText](v, "battle/kernel/item_txt.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.NameHelpText](v, "battle/kernel/item_txt.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.NameHelpText](v, "battle/kernel/item_txt.bin", rows)
		})

	registerSchema("ffx/battle/kernel/menu_txt.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx.NameHelpText](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.NameHelpText](v, "battle/kernel/menu_txt.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.NameHelpText](v, "battle/kernel/menu_txt.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.NameHelpText](v, "battle/kernel/menu_txt.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.NameHelpText](v, "battle/kernel/menu_txt.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.NameHelpText](v, "battle/kernel/menu_txt.bin", rows)
		})

	registerSchema("ffx/battle/kernel/mmain_txt.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx.NameHelpText](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.NameHelpText](v, "battle/kernel/mmain_txt.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.NameHelpText](v, "battle/kernel/mmain_txt.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.NameHelpText](v, "battle/kernel/mmain_txt.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.NameHelpText](v, "battle/kernel/mmain_txt.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.NameHelpText](v, "battle/kernel/mmain_txt.bin", rows)
		})

	registerSchema("ffx/battle/kernel/name_txt.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx.NameHelpText](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.NameHelpText](v, "battle/kernel/name_txt.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.NameHelpText](v, "battle/kernel/name_txt.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.NameHelpText](v, "battle/kernel/name_txt.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.NameHelpText](v, "battle/kernel/name_txt.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.NameHelpText](v, "battle/kernel/name_txt.bin", rows)
		})

	registerSchema("ffx/battle/kernel/save_txt.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx.NameHelpText](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.NameHelpText](v, "battle/kernel/save_txt.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.NameHelpText](v, "battle/kernel/save_txt.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.NameHelpText](v, "battle/kernel/save_txt.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.NameHelpText](v, "battle/kernel/save_txt.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.NameHelpText](v, "battle/kernel/save_txt.bin", rows)
		})

	registerSchema("ffx/battle/kernel/status_txt.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx.NameHelpText](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.NameHelpText](v, "battle/kernel/status_txt.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.NameHelpText](v, "battle/kernel/status_txt.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.NameHelpText](v, "battle/kernel/status_txt.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.NameHelpText](v, "battle/kernel/status_txt.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.NameHelpText](v, "battle/kernel/status_txt.bin", rows)
		})

	registerSchema("ffx/battle/kernel/summon_txt.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) {
			return loadGeneric[ffx.NameHelpText](v, p)
		},
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.NameHelpText](v, "battle/kernel/summon_txt.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.NameHelpText](v, "battle/kernel/summon_txt.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.NameHelpText](v, "battle/kernel/summon_txt.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.NameHelpText](v, "battle/kernel/summon_txt.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.NameHelpText](v, "battle/kernel/summon_txt.bin", rows)
		})

	registerSchema("ffx/battle/kernel/ply_rom.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx.PlyRom](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.PlyRom](v, "battle/kernel/ply_rom.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.PlyRom](v, "battle/kernel/ply_rom.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.PlyRom](v, "battle/kernel/ply_rom.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.PlyRom](v, "battle/kernel/ply_rom.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.PlyRom](v, "battle/kernel/ply_rom.bin", rows)
		})

	registerSchema("ffx/battle/kernel/ply_save.bin",
		func(v common.GameVersion, p string) (dto.Collection, error) { return loadGeneric[ffx.PlySave](v, p) },
		func(v common.GameVersion, p string, src common.FileSource) (dto.Collection, error) {
			return loadGenericFrom[ffx.PlySave](v, "battle/kernel/ply_save.bin", src)
		},
		func(v common.GameVersion, e dto.FileEntry) error {
			return applyGeneric[ffx.PlySave](v, "battle/kernel/ply_save.bin", e)
		},
		func(v common.GameVersion, data []byte) (dto.Collection, error) {
			return loadGenericFromBytes[ffx.PlySave](v, "battle/kernel/ply_save.bin", data)
		},
		func(v common.GameVersion, data []byte, rows []dto.TextRow) ([]byte, error) {
			return rebuildGeneric[ffx.PlySave](v, "battle/kernel/ply_save.bin", data, rows)
		},
		func(v common.GameVersion, p string, rows []dto.TextRow) (int, error) {
			return measureGeneric[ffx.PlySave](v, "battle/kernel/ply_save.bin", rows)
		})
}
