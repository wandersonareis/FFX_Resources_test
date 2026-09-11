package main

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/core/writer"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
	"fmt"
)

/*
MENU PRINCIPAL DE EXEMPLOS
=========================

Este arquivo contém o main() principal que organiza e executa todas as funções de exemplo
dos editores de eventos JSON.
*/
func main() {
	fmt.Println("=== EXEMPLOS DO SISTEMA DE EVENTOS FFX ===")
	fmt.Println()

	// ===== FFX (v1) =====
	interactions.NewInteractionService().FFXAppConfig().SetGameVersion(common.GameVersionFFX)
	common.SetGameFilesRoot("/home/mestre/FFX_Resources/build/bin/data/") // Defina o caminho correto para os arquivos do jogo
	common.SetVerboseMode(false)                                          // Ativa o modo verboso para depuração

	// Inicialização obrigatória
	fmt.Println("=== FFX (v1) ===")
	fmt.Println("Inicializando sistema (FFX v1)...")
	if err := reader.InitializeInternals(); err != nil {
		fmt.Printf("Erro ao inicializar sistema: %v\n", err)
		return
	}
	// Executa os exemplos para FFX v1
	//runFFXExamples()

	// ===== FFX-2 (v2) =====
	interactions.NewInteractionService().FFXAppConfig().SetGameVersion(common.GameVersionFFX2)
	common.SetGameFilesRoot("/home/mestre/FFX_Resources/build/bin/data/") // Defina o caminho correto para os arquivos do jogo

	fmt.Println("\n=== FFX-2 (v2) ===")
	fmt.Println("Reinicializando dicionários para FFX-2...")
	if err := reader.InitializeInternals(); err != nil {
		fmt.Printf("Erro ao inicializar sistema: %v\n", err)
		return
	}
	runFFX2Examples()
}

func runFFXExamples() {

	// important.bin
	keyItemsBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/important.bin")
	common.LogInfo("[FFX] ✓ Itens-chave carregados: %d\n", keyItemsBinaryFile.GetObjects().Len())
	if err := keyItemsBinaryFile.ExportToJson("key_items_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting key items to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Itens-chave exportados para JSON")
	if err := keyItemsBinaryFile.ImportFromJson("key_items_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing key items from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Itens-chave editados e salvos com sucesso")
	if err := keyItemsBinaryFile.SaveToBinary("battle/kernel/important.bin"); err != nil {
		common.LogError("[FFX] Error saving key items to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Itens-chave salvos com sucesso")

	// command.bin
	commandsBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/command.bin")
	common.LogInfo("[FFX] commands carregados: %d\n", commandsBinaryFile.GetObjects().Len())
	if err := commandsBinaryFile.ExportToJson("commands_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting commands to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Comandos exportados para JSON")
	if err := commandsBinaryFile.ImportFromJson("commands_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing commands from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Comandos editados e salvos com sucesso")
	if err := commandsBinaryFile.SaveToBinary("battle/kernel/command.bin"); err != nil {
		common.LogError("[FFX] Error saving commands to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Comandos salvos com sucesso")

	// a_ability.bin
	aAbilityBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/a_ability.bin")
	common.LogInfo("[FFX] a_ability carregados: %d\n", aAbilityBinaryFile.GetObjects().Len())
	if err := aAbilityBinaryFile.ExportToJson("a_ability_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting a_ability text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ a_ability text exportados para JSON")
	if err := aAbilityBinaryFile.ImportFromJson("a_ability_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing a_ability text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ a_ability text editados e salvos com sucesso")
	if err := aAbilityBinaryFile.SaveToBinary("battle/kernel/a_ability.bin"); err != nil {
		common.LogError("[FFX] Error saving a_ability text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ a_ability text salvos com sucesso")

	// item.bin
	itemsBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/item.bin")
	common.LogInfo("[FFX] items carregados: %d\n", itemsBinaryFile.GetObjects().Len())
	if err := itemsBinaryFile.ExportToJson("items_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting items to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Itens exportados para JSON")
	if err := itemsBinaryFile.ImportFromJson("items_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing items from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Itens editados e salvos com sucesso")
	if err := itemsBinaryFile.SaveToBinary("battle/kernel/item.bin"); err != nil {
		common.LogError("[FFX] Error saving items to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Itens salvos com sucesso")

	// arms_txt.bin
	armsTextBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/arms_txt.bin")
	common.LogInfo("[FFX] arms text carregados: %d\n", armsTextBinaryFile.GetObjects().Len())
	if err := armsTextBinaryFile.ExportToJson("arms_text_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting arms text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Arms text exportados para JSON")
	if err := armsTextBinaryFile.ImportFromJson("arms_text_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing arms text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Arms text editados e salvos com sucesso")
	if err := armsTextBinaryFile.SaveToBinary("battle/kernel/arms_txt.bin"); err != nil {
		common.LogError("[FFX] Error saving arms text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Arms text salvos com sucesso")

	// btl_txt.bin
	battleTextBinaryFile := objectsfile.ReadNameOnlyLocalizations("battle/kernel/btl_txt.bin")
	common.LogInfo("[FFX] ✓ Battle text carregados: %d\n", battleTextBinaryFile.GetObjects().Len())
	if err := battleTextBinaryFile.ExportToJson("battle_text_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting battle text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Battle text exportados para JSON")
	if err := battleTextBinaryFile.ImportFromJson("battle_text_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing battle text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Battle text editados e salvos com sucesso")
	if err := battleTextBinaryFile.SaveToBinary("battle/kernel/btl_txt.bin"); err != nil {
		common.LogError("[FFX] Error saving battle text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Battle text salvos com sucesso")

	// btlend_txt.bin
	battleEndTextBinaryFile := objectsfile.ReadNameOnlyLocalizations("battle/kernel/btlend_txt.bin")
	common.LogInfo("[FFX] ✓ Battle end text carregados: %d\n", battleEndTextBinaryFile.GetObjects().Len())
	if err := battleEndTextBinaryFile.ExportToJson("battle_end_text_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting battle end text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Battle end text exportados para JSON")
	if err := battleEndTextBinaryFile.ImportFromJson("battle_end_text_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing battle end text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Battle end text editados e salvos com sucesso")
	if err := battleEndTextBinaryFile.SaveToBinary("battle/kernel/btlend_txt.bin"); err != nil {
		common.LogError("[FFX] Error saving battle end text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Battle end text salvos com sucesso")

	// monmagic1.bin
	monsterMagic1BinaryFile := objectsfile.ReadNameOnlyLocalizations("battle/kernel/monmagic1.bin")
	common.LogInfo("[FFX] ✓ Monster magic 1 text carregados: %d\n", monsterMagic1BinaryFile.GetObjects().Len())
	if err := monsterMagic1BinaryFile.ExportToJson("monster_magic1_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting monster magic 1 text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Monster magic 1 text exportados para JSON")
	if err := monsterMagic1BinaryFile.ImportFromJson("monster_magic1_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing monster magic 1 text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Monster magic 1 text editados e salvos com sucesso")
	if err := monsterMagic1BinaryFile.SaveToBinary("battle/kernel/monmagic1.bin"); err != nil {
		common.LogError("[FFX] Error saving monster magic 1 text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Monster magic 1 text salvos com sucesso")

	// monmagic2.bin
	monsterMagic2BinaryFile := objectsfile.ReadNameOnlyLocalizations("battle/kernel/monmagic2.bin")
	common.LogInfo("[FFX] ✓ Monster magic 2 text carregados: %d\n", monsterMagic2BinaryFile.GetObjects().Len())
	if err := monsterMagic2BinaryFile.ExportToJson("monster_magic2_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting monster magic 2 text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Monster magic 2 text exportados para JSON")
	if err := monsterMagic2BinaryFile.ImportFromJson("monster_magic2_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing monster magic 2 text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Monster magic 2 text editados e salvos com sucesso")
	if err := monsterMagic2BinaryFile.SaveToBinary("battle/kernel/monmagic2.bin"); err != nil {
		common.LogError("[FFX] Error saving monster magic 2 text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Monster magic 2 text salvos com sucesso")

	// monster1.bin
	monsterBinaryFile := objectsfile.ReadMonsterLocalizations("battle/kernel/monster1.bin")
	common.LogInfo("[FFX] monster 1 carregados: %d\n", monsterBinaryFile.GetObjects().Len())
	if err := monsterBinaryFile.ExportToJson("monster1_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting monsters 1 text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ monsters 1 text exportados para JSON")
	if err := monsterBinaryFile.ImportFromJson("monster1_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing monsters 1 text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ monsters 1 text editados e salvos com sucesso")
	if err := monsterBinaryFile.SaveToBinary("battle/kernel/monster1.bin"); err != nil {
		common.LogError("[FFX] Error saving monsters 1 text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ monsters 1 text salvos com sucesso")

	// monster2.bin
	monster2BinaryFile := objectsfile.ReadMonsterLocalizations("battle/kernel/monster2.bin")
	common.LogInfo("[FFX] monster 2 carregados: %d\n", monster2BinaryFile.GetObjects().Len())
	if err := monster2BinaryFile.ExportToJson("monster2_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting monsters 2 text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ monsters 2 text exportados para JSON")
	if err := monster2BinaryFile.ImportFromJson("monster2_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing monsters 2 text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ monsters 2 text editados e salvos com sucesso")
	if err := monster2BinaryFile.SaveToBinary("battle/kernel/monster2.bin"); err != nil {
		common.LogError("[FFX] Error saving monsters 2 text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ monsters 2 text salvos com sucesso")

	// monster3.bin
	monster3BinaryFile := objectsfile.ReadMonsterLocalizations("battle/kernel/monster3.bin")
	common.LogInfo("[FFX] monster 3 carregados: %d\n", monster3BinaryFile.GetObjects().Len())
	if err := monster3BinaryFile.ExportToJson("monster3_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting monsters 3 text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ monsters 3 text exportados para JSON")
	if err := monster3BinaryFile.ImportFromJson("monster3_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing monsters 3 text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ monsters 3 text editados e salvos com sucesso")
	if err := monster3BinaryFile.SaveToBinary("battle/kernel/monster3.bin"); err != nil {
		common.LogError("[FFX] Error saving monsters 3 text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ monsters 3 text salvos com sucesso")

	// build_txt.bin
	buildTextBinaryFile := objectsfile.ReadNameOnlyLocalizations("battle/kernel/build_txt.bin")
	common.LogInfo("[FFX] ✓ Build text carregados: %d\n", buildTextBinaryFile.GetObjects().Len())
	if err := buildTextBinaryFile.ExportToJson("build_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting build text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Build text exportados para JSON")
	if err := buildTextBinaryFile.ImportFromJson("build_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing build text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Build text editados e salvos com sucesso")
	if err := buildTextBinaryFile.SaveToBinary("battle/kernel/build_txt.bin"); err != nil {
		common.LogError("[FFX] Error saving build text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Build text salvos com sucesso")

	// config_txt.bin
	configTextBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/config_txt.bin")
	common.LogInfo("[FFX] config text carregados: %d\n", configTextBinaryFile.GetObjects().Len())
	if err := configTextBinaryFile.ExportToJson("config_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting config text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Config text exportados para JSON")
	if err := configTextBinaryFile.ImportFromJson("config_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing config text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Config text editados e salvos com sucesso")
	if err := configTextBinaryFile.SaveToBinary("battle/kernel/config_txt.bin"); err != nil {
		common.LogError("[FFX] Error saving config text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Config text salvos com sucesso")

	// item_txt.bin
	itemTextBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/item_txt.bin")
	common.LogInfo("[FFX] item text carregados: %d\n", itemTextBinaryFile.GetObjects().Len())
	if err := itemTextBinaryFile.ExportToJson("item_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting item text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Item text exportados para JSON")
	if err := itemTextBinaryFile.ImportFromJson("item_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing item text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Item text editados e salvos com sucesso")
	if err := itemTextBinaryFile.SaveToBinary("battle/kernel/item_txt.bin"); err != nil {
		common.LogError("[FFX] Error saving item text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Item text salvos com sucesso")

	// mmain_txt.bin
	mainMenuTextBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/mmain_txt.bin")
	common.LogInfo("[FFX] ✓ Main menu text carregados: %d\n", mainMenuTextBinaryFile.GetObjects().Len())
	if err := mainMenuTextBinaryFile.ExportToJson("mmain_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting main menu text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Main menu text exportados para JSON")
	if err := mainMenuTextBinaryFile.ImportFromJson("mmain_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing main menu text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Main menu text editados e salvos com sucesso")
	if err := mainMenuTextBinaryFile.SaveToBinary("battle/kernel/mmain_txt.bin"); err != nil {
		common.LogError("[FFX] Error saving main menu text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Main menu text salvos com sucesso")

	// panel.bin
	panelTextBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/panel.bin")
	common.LogInfo("[FFX] ✓ Panel text carregados: %d\n", panelTextBinaryFile.GetObjects().Len())
	if err := panelTextBinaryFile.ExportToJson("panel_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting panel text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Panel text exportados para JSON")
	if err := panelTextBinaryFile.ImportFromJson("panel_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing panel text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Panel text editados e salvos com sucesso")
	if err := panelTextBinaryFile.SaveToBinary("battle/kernel/panel.bin"); err != nil {
		common.LogError("[FFX] Error saving panel text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Panel text salvos com sucesso")

	// ply_rom.bin
	playerRoomBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/ply_rom.bin")
	common.LogInfo("[FFX] ✓ Player room text carregados: %d\n", playerRoomBinaryFile.GetObjects().Len())
	if err := playerRoomBinaryFile.ExportToJson("player_rom_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting player room text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Player room text exportados para JSON")
	if err := playerRoomBinaryFile.ImportFromJson("player_rom_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing player room text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Player room text editados e salvos com sucesso")
	if err := playerRoomBinaryFile.SaveToBinary("battle/kernel/ply_rom.bin"); err != nil {
		common.LogError("[FFX] Error saving player room text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Player room text salvos com sucesso")

	// ply_save.bin
	playerSaveRoomBinaryFile := objectsfile.ReadNameOnlyV2Localizations("battle/kernel/ply_save.bin")
	common.LogInfo("[FFX] ✓ Player save text carregados: %d\n", playerSaveRoomBinaryFile.GetObjects().Len())
	if err := playerSaveRoomBinaryFile.ExportToJson("player_save_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting player save text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Player save text exportados para JSON")
	if err := playerSaveRoomBinaryFile.ImportFromJson("player_save_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing player save text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Player save text editados e salvos com sucesso")
	if err := playerSaveRoomBinaryFile.SaveToBinary("battle/kernel/ply_save.bin"); err != nil {
		common.LogError("[FFX] Error saving player save text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Player save text salvos com sucesso")

	// sphere.bin
	sphereBinaryFile := objectsfile.ReadNameOnlyLocalizations("battle/kernel/sphere.bin")
	common.LogInfo("[FFX] ✓ Sphere text carregados: %d\n", sphereBinaryFile.GetObjects().Len())
	if err := sphereBinaryFile.ExportToJson("sphere_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting sphere text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Sphere text exportados para JSON")
	if err := sphereBinaryFile.ImportFromJson("sphere_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing sphere text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Sphere text editados e salvos com sucesso")
	if err := sphereBinaryFile.SaveToBinary("battle/kernel/sphere.bin"); err != nil {
		common.LogError("[FFX] Error saving sphere text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Sphere text salvos com sucesso")

	// save_txt.bin
	saveTextBinaryFile := objectsfile.ReadNameOnlyLocalizations("battle/kernel/save_txt.bin")
	common.LogInfo("[FFX] ✓ Save text carregados: %d\n", saveTextBinaryFile.GetObjects().Len())
	if err := saveTextBinaryFile.ExportToJson("save_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting save text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Save text exportados para JSON")
	if err := saveTextBinaryFile.ImportFromJson("save_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing save text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Save text editados e salvos com sucesso")
	if err := saveTextBinaryFile.SaveToBinary("battle/kernel/save_txt.bin"); err != nil {
		common.LogError("[FFX] Error saving save text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Save text salvos com sucesso")

	// status_txt.bin
	statusTextBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/status_txt.bin")
	common.LogInfo("[FFX] ✓ Status text carregados: %d\n", statusTextBinaryFile.GetObjects().Len())
	if err := statusTextBinaryFile.ExportToJson("status_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting status text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Status text exportados para JSON")
	if err := statusTextBinaryFile.ImportFromJson("status_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing status text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Status text editados e salvos com sucesso")
	if err := statusTextBinaryFile.SaveToBinary("battle/kernel/status_txt.bin"); err != nil {
		common.LogError("[FFX] Error saving status text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Status text salvos com sucesso")

	// summon_txt.bin
	summonTextBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/summon_txt.bin")
	common.LogInfo("[FFX] ✓ Summon text carregados: %d\n", summonTextBinaryFile.GetObjects().Len())
	if err := summonTextBinaryFile.ExportToJson("summon_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting summon text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Summon text exportados para JSON")
	if err := summonTextBinaryFile.ImportFromJson("summon_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing summon text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Summon text editados e salvos com sucesso")
	if err := summonTextBinaryFile.SaveToBinary("battle/kernel/summon_txt.bin"); err != nil {
		common.LogError("[FFX] Error saving summon text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Summon text salvos com sucesso")

	// w_name.bin
	weaponNamesBinaryFile := objectsfile.ReadWeaponNamesLocalizations("battle/kernel/w_name.bin")
	common.LogInfo("[FFX] ✓ Weapon names carregados: %d\n", weaponNamesBinaryFile.GetObjects().Len())
	if err := weaponNamesBinaryFile.ExportToJson("weapon_names_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting weapon names to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Weapon names exportados para JSON")
	if err := weaponNamesBinaryFile.ImportFromJson("weapon_names_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing weapon names from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Weapon names editados e salvos com sucesso")
	if err := weaponNamesBinaryFile.SaveToBinary("battle/kernel/w_name.bin"); err != nil {
		common.LogError("[FFX] Error saving weapon names to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Weapon names salvos com sucesso")

	// name_txt.bin
	nameTextBinaryFile := objectsfile.ReadNameOnlyLocalizations("battle/kernel/name_txt.bin")
	common.LogInfo("[FFX] ✓ Name text carregados: %d\n", nameTextBinaryFile.GetObjects().Len())
	if err := nameTextBinaryFile.ExportToJson("name_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error exporting name text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Name text exportados para JSON")
	if err := nameTextBinaryFile.ImportFromJson("name_txt_all_localizations.json"); err != nil {
		common.LogError("[FFX] Error importing name text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Name text editados e salvos com sucesso")
	if err := nameTextBinaryFile.SaveToBinary("battle/kernel/name_txt.bin"); err != nil {
		common.LogError("[FFX] Error saving name text to binary: %v\n", err)
	}
	common.LogInfo("[FFX] ✓ Name text salvos com sucesso")

	BinaryReconstructionExample()

	// Carregar eventos primeiro
	if err := readEvents(); err != nil {
		fmt.Printf("Erro ao carregar eventos: %v\n", err)
		return
	}

	//converter.ExportAllEventsToJSON()
	fmt.Println("✓ Eventos exportados para JSON")
	gameVersion := interactions.CurrentGameVersion()
	event.ImportEventsDataFromJsonFile(gameVersion)
	fmt.Println("✓ Eventos editados e salvos com sucesso")
	event.ExportAllEventsForLocalizations(gameVersion)
	fmt.Println("✓ Eventos salvos com sucesso")

	showMainMenu()
}

func runFFX2Examples() {
	common.LogInfo("Extraindo arquivos de name+description do FFX-2 (v2) para JSON...")

	// Arquivos em comum com a v1 (mesmos nomes existem no FFX-2).
	// important.bin
	keyItemsBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/important.bin")
	common.LogInfo("[FFX-2] ✓ Itens-chave carregados: %d\n", keyItemsBinaryFile.GetObjects().Len())
	keyItemsBinaryFile.ExportToJson("key_items_all_localizations.json")
	common.LogInfo("[FFX-2] ✓ Itens-chave exportados para JSON")
	keyItemsBinaryFile.ImportFromJson("key_items_all_localizations.json")
	common.LogInfo("[FFX-2] ✓ Itens-chave editados e salvos com sucesso")
	keyItemsBinaryFile.SaveToBinary("battle/kernel/important.bin")
	common.LogInfo("[FFX-2] ✓ Itens-chave salvos com sucesso")

	// commands.bin
	commandsBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/command.bin")
	common.LogInfo("[FFX-2] commands carregados: %d\n", commandsBinaryFile.GetObjects().Len())
	if err := commandsBinaryFile.ExportToJson("commands_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting commands to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Comandos exportados para JSON")
	if err := commandsBinaryFile.ImportFromJson("commands_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing commands from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Comandos editados e salvos com sucesso")
	if err := commandsBinaryFile.SaveToBinary("battle/kernel/command.bin"); err != nil {
		common.LogError("[FFX-2] Error saving commands to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Comandos salvos com sucesso")

	// item.bin
	itemsBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/item.bin")
	common.LogInfo("[FFX-2] items carregados: %d\n", datastore.Items.Len())
	if err := itemsBinaryFile.ExportToJson("items_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting items to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Itens exportados para JSON")
	if err := itemsBinaryFile.ImportFromJson("items_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing items from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Itens editados e salvos com sucesso")
	if err := itemsBinaryFile.SaveToBinary("battle/kernel/item.bin"); err != nil {
		common.LogError("[FFX-2] Error saving items to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Itens salvos com sucesso")

	// menu_txt.bin
	menuTextBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/menu_txt.bin")
	common.LogInfo("[FFX-2] menu text carregados: %d\n", menuTextBinaryFile.GetObjects().Len())
	if err := menuTextBinaryFile.ExportToJson("menu_text_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting menu text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Menu text exportados para JSON")
	if err := menuTextBinaryFile.ImportFromJson("menu_text_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing menu text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Menu text editados e salvos com sucesso")
	if err := menuTextBinaryFile.SaveToBinary("battle/kernel/menu_txt.bin"); err != nil {
		common.LogError("[FFX-2] Error saving menu text to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Menu text salvos com sucesso")

	// oversoul.bin
	oversoulBinaryFile := objectsfile.ReadNameOnlyV2Localizations("battle/kernel/oversoul.bin")
	common.LogInfo("[FFX-2] oversoul carregados: %d\n", oversoulBinaryFile.GetObjects().Len())
	if err := oversoulBinaryFile.ExportToJson("oversoul_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting oversoul text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Oversoul text exportados para JSON")
	if err := oversoulBinaryFile.ImportFromJson("oversoul_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing oversoul text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Oversoul text editados e salvos com sucesso")
	if err := oversoulBinaryFile.SaveToBinary("battle/kernel/oversoul.bin"); err != nil {
		common.LogError("[FFX-2] Error saving oversoul text to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Oversoul text salvos com sucesso")

	// btl_txt.bin
	battleTextBinaryFile := objectsfile.ReadNameOnlyV2Localizations("battle/kernel/btl_txt.bin")
	common.LogInfo("[FFX-2] battle text carregados: %d\n", battleTextBinaryFile.GetObjects().Len())
	if err := battleTextBinaryFile.ExportToJson("battle_text_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting battle text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Battle text exportados para JSON")
	if err := battleTextBinaryFile.ImportFromJson("battle_text_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing battle text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Battle text editados e salvos com sucesso")
	if err := battleTextBinaryFile.SaveToBinary("battle/kernel/btl_txt.bin"); err != nil {
		common.LogError("[FFX-2] Error saving battle text to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Battle text salvos com sucesso")

	// btlend_txt.bin
	battleEndTextBinaryFile := objectsfile.ReadNameOnlyV2Localizations("battle/kernel/btlend_txt.bin")
	common.LogInfo("[FFX-2] battle end text: %d\n", datastore.BattleEndTxt.Len())
	if err := battleEndTextBinaryFile.ExportToJson("battle_end_text_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting battle end text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Battle end text exportados para JSON")
	if err := battleEndTextBinaryFile.ImportFromJson("battle_text_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing battle end text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Battle end text editados e salvos com sucesso")
	if err := battleEndTextBinaryFile.SaveToBinary("battle/kernel/btlend_txt.bin"); err != nil {
		common.LogError("[FFX-2] Error saving battle end text to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Battle end text salvos com sucesso")

	// a_ability.bin
	aAbilityBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/a_ability.bin")
	common.LogInfo("[FFX-2] a_ability carregados: %d\n", aAbilityBinaryFile.GetObjects().Len())
	if err := aAbilityBinaryFile.ExportToJson("a_ability_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting a_ability text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ a_ability text exportados para JSON")
	if err := aAbilityBinaryFile.ImportFromJson("a_ability_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing a_ability text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ a_ability text editados e salvos com sucesso")
	if err := aAbilityBinaryFile.SaveToBinary("battle/kernel/a_ability.bin"); err != nil {
		common.LogError("[FFX-2] Error saving a_ability text to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ a_ability text salvos com sucesso")

	// Exclusivos do FFX-2 (retornam cedo se a versão não for FFX-2).
	// accessory.bin
	accessoriesBinaryFile := objectsfile.ReadJobLocalizations("battle/kernel/accessory.bin", objectsfile.FFx2AccessoryEffectSegmentDefaultPosition)
	common.LogInfo("[FFX-2] accessory carregados: %d\n", accessoriesBinaryFile.GetObjects().Len())
	if err := accessoriesBinaryFile.ExportToJson("accessory_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting accessories text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ accessories text exportados para JSON")
	if err := accessoriesBinaryFile.ImportFromJson("accessory_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing accessories text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ accessories text editados e salvos com sucesso")
	if err := accessoriesBinaryFile.SaveToBinary("battle/kernel/accessory.bin"); err != nil {
		common.LogError("[FFX-2] Error saving accessories text to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ accessories text salvos com sucesso")

	// job.bin
	jobsBinaryFile := objectsfile.ReadJobLocalizations("battle/kernel/job.bin", objectsfile.FFx2JobEffectSegmentDefaultPosition)
	common.LogInfo("[FFX-2] job carregados: %d\n", jobsBinaryFile.GetObjects().Len())
	if err := jobsBinaryFile.ExportToJson("job_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting jobs text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ jobs text exportados para JSON")
	if err := jobsBinaryFile.ImportFromJson("job_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing jobs text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ jobs text editados e salvos com sucesso")
	if err := jobsBinaryFile.SaveToBinary("battle/kernel/job.bin"); err != nil {
		common.LogError("[FFX-2] Error saving jobs text to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ jobs text salvos com sucesso")

	// monmagic.bin
	monMagicBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/monmagic.bin")
	common.LogInfo("[FFX-2] monster magic carregados: %d\n", monMagicBinaryFile.GetObjects().Len())
	if err := monMagicBinaryFile.ExportToJson("monster_magic_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting monster magic text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ monster magic text exportados para JSON")
	if err := monMagicBinaryFile.ImportFromJson("monster_magic_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing monster magic text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ monster magic text editados e salvos com sucesso")
	if err := monMagicBinaryFile.SaveToBinary("battle/kernel/monmagic.bin"); err != nil {
		common.LogError("[FFX-2] Error saving monster magic text to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ monster magic text salvos com sucesso")

	// monster.bin
	monsterBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/monster.bin")
	common.LogInfo("[FFX-2] monster carregados: %d\n", monsterBinaryFile.GetObjects().Len())
	if err := monsterBinaryFile.ExportToJson("monster_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting monsters text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ monsters text exportados para JSON")
	if err := monsterBinaryFile.ImportFromJson("monster_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing monsters text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ monsters text editados e salvos com sucesso")
	if err := monsterBinaryFile.SaveToBinary("battle/kernel/monster.bin"); err != nil {
		common.LogError("[FFX-2] Error saving monsters text to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ monsters text salvos com sucesso")

	// monster2.bin
	monster2BinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/monster2.bin")
	common.LogInfo("[FFX-2] monster2 carregados: %d\n", monster2BinaryFile.GetObjects().Len())
	if err := monster2BinaryFile.ExportToJson("monster2_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting monsters 2 text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ monsters 2 text exportados para JSON")
	if err := monster2BinaryFile.ImportFromJson("monster2_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing monsters 2 text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ monsters 2 text editados e salvos com sucesso")
	if err := monster2BinaryFile.SaveToBinary("battle/kernel/monster2.bin"); err != nil {
		common.LogError("[FFX-2] Error saving monsters 2 text to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ monsters 2 text salvos com sucesso")

	// plate.bin
	plateBinaryFile := objectsfile.ReadNameDescriptionEffectAbilitiesLocalizations("battle/kernel/plate.bin", objectsfile.FFx2PlateAbilitiesCount, objectsfile.FFx2PlateEffectSegmentDefaultPosition)
	common.LogInfo("[FFX-2] plate carregados: %d\n", plateBinaryFile.GetObjects().Len())
	if err := plateBinaryFile.ExportToJson("plate_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting plate text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ plate text exportados para JSON")
	if err := plateBinaryFile.ImportFromJson("plate_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing plate text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ plate text editados e salvos com sucesso")
	if err := plateBinaryFile.SaveToBinary("battle/kernel/plate.bin"); err != nil {
		common.LogError("[FFX-2] Error saving plate text to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ plate text salvos com sucesso")

	// ply_save.bin
	playerSaveBinaryFile := objectsfile.ReadNameOnlyV2Localizations("battle/kernel/ply_save.bin")
	common.LogInfo("[FFX-2] player save: %d\n", playerSaveBinaryFile.GetObjects().Len())
	if err := playerSaveBinaryFile.ExportToJson("player_save_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting player save text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ player save text exportados para JSON")
	if err := playerSaveBinaryFile.ImportFromJson("player_save_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing player save text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ player save text editados e salvos com sucesso")
	if err := playerSaveBinaryFile.SaveToBinary("battle/kernel/ply_save.bin"); err != nil {
		common.LogError("[FFX-2] Error saving player save text to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ player save text salvos com sucesso")

	// ply_rom.bin
	playerRoomTextBinaryFile := objectsfile.ReadNameOnlyV2Localizations("battle/kernel/ply_rom.bin")
	common.LogInfo("[FFX-2] player room carregados: %d\n", playerRoomTextBinaryFile.GetObjects().Len())
	if err := playerRoomTextBinaryFile.ExportToJson("player_rom_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting player room text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Player room text exportados para JSON")
	if err := playerRoomTextBinaryFile.ImportFromJson("player_rom_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing player room text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Player room text editados e salvos com sucesso")
	if err := playerRoomTextBinaryFile.SaveToBinary("battle/kernel/ply_rom.bin"); err != nil {
		common.LogError("[FFX-2] Error saving player room text to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Player room text salvos com sucesso")

	// save_txt.bin
	saveTextBinaryFile := objectsfile.ReadCommandLocalizations("battle/kernel/save_txt.bin")
	fmt.Printf("[FFX-2] save text carregados: %d\n", saveTextBinaryFile.GetObjects().Len())
	if err := saveTextBinaryFile.ExportToJson("save_text_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting save text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Save text exportados para JSON")
	if err := saveTextBinaryFile.ImportFromJson("save_text_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing save text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Save text editados e salvos com sucesso")
	if err := saveTextBinaryFile.SaveToBinary("battle/kernel/save_txt.bin"); err != nil {
		common.LogError("[FFX-2] Error saving save text to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Save text salvos com sucesso")

	// lm_accesary.bin
	lmAccesaryTextBinaryFile := objectsfile.ReadCommandLocalizations("lastmiss/kernel/lm_accesary.bin")
	common.LogInfo("[FFX-2] Last mission accesary text carregados: %d\n", lmAccesaryTextBinaryFile.GetObjects().Len())
	if err := lmAccesaryTextBinaryFile.ExportToJson("lm_accesary_text_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error exporting last mission accesary text to JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Last mission accesary text exportados para JSON")
	if err := lmAccesaryTextBinaryFile.ImportFromJson("lm_accesary_text_all_localizations.json"); err != nil {
		common.LogError("[FFX-2] Error importing last mission accesary text from JSON: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Last mission accesary text editados e salvos com sucesso")
	if err := lmAccesaryTextBinaryFile.SaveToBinary("lastmiss/kernel/lm_accesary.bin"); err != nil {
		common.LogError("[FFX-2] Error saving last mission accesary text to binary: %v\n", err)
	}
	common.LogInfo("[FFX-2] ✓ Last mission accesary text salvos com sucesso")

	// Carregar eventos
	if err := readEvents(); err != nil {
		fmt.Printf("Erro ao carregar eventos: %v\n", err)
		return
	}

	//converter.ExportAllEventsToJSON()
	fmt.Println("✓ Eventos exportados para JSON")
	gameVersion := interactions.CurrentGameVersion()
	event.ImportEventsDataFromJsonFile(gameVersion)
	fmt.Println("✓ Eventos editados e salvos com sucesso")
	event.ExportAllEventsForLocalizations(gameVersion)
	fmt.Println("✓ Eventos salvos com sucesso")

	fmt.Println("✓ FFX-2 (v2) extraído para JSON (arquivos *_v2_*.json)")

	showMainMenu()
}

func readEvents() error {
	fmt.Println("Carregando eventos...")
	eventsFolder, err := common.NewFileAccessor(common.GetPathOriginalsEvent())
	if err != nil {
		return fmt.Errorf("failed to resolve events directory: %w", err)
	}
	version := interactions.CurrentGameVersion()
	if err := event.ReadAllEventFiles(eventsFolder, string(version)); err != nil {
		fmt.Printf("Erro ao carregar eventos: %v\n", err)
		return err
	}

	// Contar eventos carregados no datastore
	eventIDs := event.GetAllEventIDs(models.GameVersion(version))
	fmt.Printf("✓ Carregados %d eventos no datastore\n", len(eventIDs))
	return nil
}

func showMainMenu() {
	// ===== EXEMPLOS BÁSICOS =====
	ExportMacroDictionaryExample() // Exemplo de exportação de dicionário de macros (JSON com metadados)
	// ExampleJSONEditorUsage()

	// ===== WORKFLOWS COMPLETOS =====
	// ExampleCompleteJSONWorkflow()
	// ===== EDIÇÃO DIRECIONADA =====
	//writer.WriteStringsEventForAllLocalizationsJSON("znkd1500", true)
	//ExportMacroDictionaryExample()           // Exemplo de exportação de dicionário de macros
	writer.ExportAllLocalizationsToJSON() // Exporta todos os eventos para JSON
	//writer.WriteStringsEventForAllLocalizationsJSON("akagi0100", true) // Exporta o evento "akagi0100" para JSON
	writer.ExportMacroDictionaryToJSON()
	//reader.EditAndSaveMacroDictJSONFiles(true) // Exemplo de fluxo completo de exportação/importação
	reader.EditAndSaveSpecificEventFromJSON("znkd1500") // Edita e salva o evento "znkd1500" do JSON
	// ===== WORKFLOW JSON =====
	// demoWorkflowJSON()
	// exemploEditorEspecificoDemo()	// ===== EXEMPLO ATIVO (descomente para testar) =====
	// Exemplo básico ativo para teste:
	// ExampleSpecificEventJSONEdit("ev001") // Testando a nova função
}
