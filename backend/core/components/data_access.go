package components

type DataAccess struct{}

var (
	//MOVES               = make([]*model.CommandDataObject, 0x10000)
	EVENTS = make(map[string]*EventFile)
	//ENCOUNTERS          = make(map[string]*atel.EncounterFile)
	//MONSTERS            = make([]*atel.MonsterFile, 0x1000)
	//MENUMAIN            *atel.AtelScriptObject
	MACRODICTFILE = make(map[string][][]*MacroString)
	//ENCOUNTER_TABLES    []*model.FieldEncounterTableDataObject
	//PLAYER_CHAR_STATS   []*model.PlayerCharStatDataObject
	//PLAYER_ROM          []*model.PlayerRomDataObject
	//WEAPON_NAMES        []*model.WeaponNameDataObject
	//AUTO_ABILITIES      []*model.AutoAbilityDataObject
	KEY_ITEMS []*KeyItemDataObject
	//TREASURES           []*model.TreasureDataObject
	//MIX_COMBINATIONS    []*model.MixCombinationDataObject
	//CTB_BASE            []*model.CtbBaseDataObject
	//WEAPON_PICKUPS      []*model.GearDataObject
	//BUYABLE_GEAR        []*model.GearDataObject
	//GEAR_SHOPS          []*model.GearShopDataObject
	//ITEM_SHOPS          []*model.ItemShopDataObject
	//SG_NODE_TYPES       []*spheregrid.SphereGridNodeTypeDataObject
	//SG_SPHERE_TYPES     []*spheregrid.SphereGridSphereTypeDataObject
	//OSG_LAYOUT          *spheregrid.SphereGridLayoutDataObject
	//SSG_LAYOUT          *spheregrid.SphereGridLayoutDataObject
	//ESG_LAYOUT          *spheregrid.SphereGridLayoutDataObject
	//GEAR_CUSTOMIZATIONS []*model.CustomizationDataObject
	//AEON_CUSTOMIZATIONS []*model.CustomizationDataObject

	ARMS_TEXT   []*NameDescriptionTextObject
	BTLEND_TEXT []*NameDescriptionTextObject
	BUILD_TEXT  []*NameDescriptionTextObject
	CONFIG_TEXT []*NameDescriptionTextObject
	ITEM_TEXT   []*NameDescriptionTextObject
	MENU_TEXT   []*NameDescriptionTextObject
	MMAIN_TEXT  []*NameDescriptionTextObject
	NAME_TEXT   []*NameDescriptionTextObject
	SAVE_TEXT   []*NameDescriptionTextObject
	STATS_TEXT  []*NameDescriptionTextObject
	SUMMON_TEXT []*NameDescriptionTextObject
)

var dummyObject = NameableFunc(func(string) string {
	return "null"
})

type Nameable interface {
	GetName(string) string
}

type NameableFunc func(string) string

func (f NameableFunc) GetName(locale string) string {
	return f(locale)
}

func GetNameableObject(typ string, idx int) Nameable {
	switch typ {
	/* case "command":
	if cmd := GetCommand(idx); cmd != nil {
		return cmd
	} */
	/* case "monster":
	if m := GetMonster(idx); m != nil {
		return m
	} */
	case "keyItem":
		if k := GetKeyItem(idx); k != nil {
			return k
		}
		/* case "treasure":
		if t := GetTreasure(idx); t != nil {
			return t
		} */
		/* case "sgNodeType":
		if s := GetSgNodeType(idx); s != nil {
			return s
		}*/
	}
	return nil
}

func GetEvent(id string) *EventFile {
	return EVENTS[id]
}

/* func GetEncounter(id string) *atel.EncounterFile {
	return ENCOUNTERS[id]
} */

/* func GetCommand(idx int) *model.CommandDataObject {
	if MOVES == nil || idx >= len(MOVES) {
		return nil
	}
	return MOVES[idx]
} */

/* func GetAutoAbility(idx int) *model.AutoAbilityDataObject {
	if idx == 0x00FF || AUTO_ABILITIES == nil {
		return nil
	}
	actual := idx - 0x8000
	if actual >= 0 && actual < len(AUTO_ABILITIES) {
		return AUTO_ABILITIES[actual]
	}
	return nil
} */

/* func GetSgNodeType(idx int) *spheregrid.SphereGridNodeTypeDataObject {
	if SG_NODE_TYPES == nil || idx < 0 || idx >= len(SG_NODE_TYPES) {
		return nil
	}
	return SG_NODE_TYPES[idx]
} */

func GetKeyItem(idx int) *KeyItemDataObject {
	if KEY_ITEMS == nil {
		return nil
	}
	actual := idx - 0xA000
	if actual >= 0 && actual < len(KEY_ITEMS) {
		return KEY_ITEMS[actual]
	}
	return nil
}

/* func GetTreasure(idx int) *model.TreasureDataObject {
	if TREASURES == nil || idx >= len(TREASURES) {
		return nil
	}
	return TREASURES[idx]
} */

/* func AddMonsterLocalizations(localizations []*model.MonsterStatDataObject) {
	if localizations == nil {
		return
	}
	for i := 0; i < len(localizations) && i < len(MONSTERS); i++ {
		if MONSTERS[i] != nil {
			MONSTERS[i].MonsterStatData.SetLocalizations(localizations[i])
		}
	}
}
*/
/* func GetMonster(idx int) *atel.MonsterFile {
	if MONSTERS == nil {
		return nil
	}
	actual := idx - 0x1000
	if actual >= 0 && actual < len(MONSTERS) {
		return MONSTERS[actual]
	}
	return nil
} */
