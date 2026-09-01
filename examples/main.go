package main

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/exporters"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/core/writer"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/objectsfile"
	"fmt"
)

/*
MENU PRINCIPAL DE EXEMPLOS
=========================

Este arquivo contém o main() principal que organiza e executa todas as funções de exemplo
dos editores de eventos CSV e JSON.
*/
func main() {
	fmt.Println("=== EXEMPLOS DO SISTEMA DE EVENTOS FFX ===")
	fmt.Println()

	// ===== FFX (v1) =====
	common.SetGameVersion(1)
	common.SetGameFilesRoot("/home/mestre/FFX_Resources/build/bin/data/") // Defina o caminho correto para os arquivos do jogo
	common.SetVerboseMode(true)                                           // Ativa o modo verboso para depuração

	// Inicialização obrigatória
	fmt.Println("=== FFX (v1) ===")
	fmt.Println("Inicializando sistema (FFX v1)...")
	if err := reader.InitializeInternals(); err != nil {
		fmt.Printf("Erro ao inicializar sistema: %v\n", err)
		return
	}
	// Executa os exemplos para FFX v1
	//runFFXv1Examples()

	// ===== FFX-2 (v2) =====
	common.SetGameVersion(2)
	common.SetGameFilesRoot("/home/mestre/FFX_Resources/build/bin/data/") // Defina o caminho correto para os arquivos do jogo

	fmt.Println("\n=== FFX-2 (v2) ===")
	fmt.Println("Reinicializando dicionários para FFX-2...")
	if err := reader.InitializeInternals(); err != nil {
		fmt.Printf("Erro ao inicializar sistema: %v\n", err)
		return
	}
	runFFX2Examples()
}

func runFFXv1Examples() {

	// key items
	// ok
	objectsfile.ReadKeyItemsWithAllLocalizations()
	fmt.Printf("✓ Itens-chave carregados: %d\n", datastore.KeyItems.Len())
	exporters.ExportKeyItemsToJSON()
	fmt.Println("✓ Itens-chave exportados para JSON")
	objectsfile.ProcessKeyItemsJsonFile()
	fmt.Println("✓ Itens-chave editados e salvos com sucesso")
	objectsfile.WriteAllKeyItemsData()
	fmt.Println("✓ Itens-chave salvos com sucesso")

	// commands
	fmt.Println("Carregando comandos...")
	objectsfile.ReadCommandsWithAllLocalizations()
	fmt.Printf("✓ Habilidades carregadas: %d\n", datastore.Commands.Len())
	exporters.ExportCommandsToJSON()
	fmt.Println("✓ Habilidades exportadas para JSON")
	objectsfile.ProcessCommandsJsonFile()
	fmt.Println("✓ Comandos editados e salvos com sucesso")
	objectsfile.WriteAllCommandsData()
	fmt.Println("✓ Comandos salvos com sucesso")

	// items
	// ok
	fmt.Println("Carregando itens...")
	objectsfile.ReadItemsWithAllLocalizations()
	fmt.Printf("✓ Itens carregados: %d\n", datastore.Items.Len())
	exporters.ExportItemsToJSON()
	fmt.Println("✓ Itens exportados para JSON")
	objectsfile.ProcessItemsJsonFile()
	fmt.Println("✓ Itens editados e salvos com sucesso")
	objectsfile.WriteAllItemsData()
	fmt.Println("✓ Itens salvos com sucesso")

	// arms text
	// ok
	fmt.Println("Carregando textos de armas...")
	objectsfile.ReadArmsTextWithAllLocalizations()
	fmt.Printf("✓ Textos de armas carregados: %d\n", datastore.ArmsTxt.Len())
	exporters.ExportArmsToJSON()
	fmt.Println("✓ Textos de armas exportados para JSON")
	objectsfile.ProcessArmsJsonFile()
	fmt.Println("✓ Textos de armas editados e salvos com sucesso")
	objectsfile.WriteAllArmsTextData()
	fmt.Println("✓ Textos de armas salvos com sucesso")

	// battle text
	// ok
	objectsfile.ReadBattleTextWithAllLocalizations()
	fmt.Printf("✓ Textos de batalha carregados: %d\n", datastore.BattleTxt.Len())
	exporters.ExportBattleToJSON()
	fmt.Println("✓ Textos de batalha exportados para JSON")
	objectsfile.ProcessBattleTextJsonFile()
	fmt.Println("✓ Textos de batalha editados e salvos com sucesso")
	objectsfile.WriteAllBattleTextData()
	fmt.Println("✓ Textos de batalha salvos com sucesso")

	// battle end text
	// ok
	objectsfile.ReadBattleEndTextWithAllLocalizations()
	fmt.Printf("✓ Textos de fim de batalha carregados: %d\n", datastore.BattleEndTxt.Len())
	exporters.ExportBattleEndToJSON()
	fmt.Println("✓ Textos de fim de batalha exportados para JSON")
	objectsfile.ProcessBattleEndTextJsonFile()
	fmt.Println("✓ Textos de fim de batalha editados e salvos com sucesso")
	objectsfile.WriteAllBattleEndTextData()
	fmt.Println("✓ Textos de fim de batalha salvos com sucesso")

	// monster magic 1
	// ok
	fmt.Println("Carregando itens mágicos...")
	objectsfile.ReadMonsterMagic1WithAllLocalizations()
	fmt.Printf("✓ Itens carregados: %d\n", objectsfile.MONMAGIC1.Len())
	exporters.ExportMonsterMagic1ToJSON()
	fmt.Println("✓ Itens exportados para JSON")
	objectsfile.ProcessMonsterMagic1JsonFile()
	fmt.Println("✓ Comandos editados e salvos com sucesso")
	objectsfile.WriteAllMonsterMagic1Data()
	fmt.Println("✓ Itens mágicos salvos com sucesso")

	// monster magic 2
	// ok
	objectsfile.ReadMonsterMagic2WithAllLocalizations()
	fmt.Printf("✓ Itens carregados: %d\n", objectsfile.MONMAGIC2.Len())
	exporters.ExportMonsterMagic2ToJSON()
	fmt.Println("✓ Itens exportados para JSON")
	objectsfile.ProcessMonsterMagic2JsonFile()
	fmt.Println("✓ Comandos editados e salvos com sucesso")
	objectsfile.WriteAllMonsterMagic2Data()
	fmt.Println("✓ Itens mágicos salvos com sucesso")

	// build text
	fmt.Println("Carregando textos de build...")
	objectsfile.ReadBuildTextWithAllLocalizations()
	fmt.Printf("✓ Textos de build carregados: %d\n", objectsfile.BUILD_TEXT.Len())
	exporters.ExportBuildToJSON()
	fmt.Println("✓ Textos de build exportados para JSON")
	objectsfile.ProcessBuildTextJsonFile()
	fmt.Println("✓ Textos de build editados e salvos com sucesso")
	objectsfile.WriteAllBuildTextData()
	fmt.Println("✓ Textos de build salvos com sucesso")

	// config text
	fmt.Println("Carregando textos de configuração...")
	objectsfile.ReadConfigTextWithAllLocalizations()
	fmt.Printf("✓ Textos de configuração carregados: %d\n", objectsfile.CONFIG_TEXT.Len())
	exporters.ExportConfigToJSON()
	fmt.Println("✓ Textos de configuração exportados para JSON")
	objectsfile.ProcessConfigTextJsonFile()
	fmt.Println("✓ Textos de configuração editados e salvos com sucesso")
	objectsfile.WriteAllConfigTextData()
	fmt.Println("✓ Textos de configuração salvos com sucesso")

	// item text
	fmt.Println("Carregando textos de itens...")
	objectsfile.ReadItemCommandsWithAllLocalizations()
	fmt.Printf("✓ Textos de itens carregados: %d\n", objectsfile.ITEM_TEXT.Len())
	exporters.ExportItemCommandsToJSON()
	fmt.Println("✓ Textos de itens exportados para JSON")
	objectsfile.ProcessItemCommandsJsonFile()
	fmt.Println("✓ Textos de itens editados e salvos com sucesso")
	objectsfile.WriteAllItemCommandsData()
	fmt.Println("✓ Textos de itens salvos com sucesso")

	// main menu text
	fmt.Println("Carregando textos do menu principal...")
	objectsfile.ReadMainMenuTextWithAllLocalizations()
	fmt.Printf("✓ Textos do menu principal carregados: %d\n", objectsfile.MMAIN_TEXT.Len())
	exporters.ExportMainMenuToJSON()
	fmt.Println("✓ Textos do menu principal exportados para JSON")
	objectsfile.ProcessMainMenuTextJsonFile()
	fmt.Println("✓ Textos do menu principal editados e salvos com sucesso")
	objectsfile.WriteAllMainMenuTextData()
	fmt.Println("✓ Textos do menu principal salvos com sucesso")

	// player blitsball text
	fmt.Println("Carregando textos de jogador de blitsball...")
	objectsfile.ReadPlayerRomTextWithAllLocalizations()
	fmt.Printf("✓ Textos de jogador de blitsball carregados: %d\n", objectsfile.PLAYER_ROOM.Len())
	exporters.ExportPlayerRoomToJSON()
	fmt.Println("✓ Textos de jogador de blitsball exportados para JSON")
	objectsfile.ProcessPlayerRoomTextJsonFile()
	fmt.Println("✓ Textos de jogador de blitsball editados e salvos com sucesso")
	objectsfile.WriteAllPlayerRoomTextData()
	fmt.Println("✓ Textos de jogador de blitsball salvos com sucesso")

	// name text
	fmt.Println("Carregando textos de nomes...")
	objectsfile.ReadNameTextWithAllLocalizations()
	fmt.Printf("✓ Textos de nomes carregados: %d\n", objectsfile.NAME_TEXT.Len())
	exporters.ExportNameToJSON()
	fmt.Println("✓ Textos de nomes exportados para JSON")
	objectsfile.ProcessNameTextJsonFile()
	fmt.Println("✓ Textos de nomes editados e salvos com sucesso")
	objectsfile.WriteAllNameTextData()
	fmt.Println("✓ Textos de nomes salvos com sucesso")

	// Carregar eventos primeiro
	if err := readEvents(); err != nil {
		fmt.Printf("Erro ao carregar eventos: %v\n", err)
		return
	}

	//converter.ExportAllEventsToJSON()
	fmt.Println("✓ Eventos exportados para JSON")
	event.ImportEventsDataFromJsonFile()
	fmt.Println("✓ Eventos editados e salvos com sucesso")
	event.ExportAllEventsForLocalizations()
	fmt.Println("✓ Eventos salvos com sucesso")

	showMainMenu()
}

func runFFX2Examples() {
	fmt.Println("Extraindo arquivos de name+description do FFX-2 (v2) para JSON...")

	// Arquivos em comum com a v1 (mesmos nomes existem no FFX-2).
	objectsfile.ReadKeyItemsWithAllLocalizations()
	fmt.Printf("[FFX-2] key items: %d\n", datastore.KeyItems.Len())
	if datastore.KeyItems.Len() > 0 {
		exporters.ExportKeyItemsToJSON()
	}

	objectsfile.ReadCommandsWithAllLocalizations()
	fmt.Printf("[FFX-2] commands: %d\n", datastore.Commands.Len())
	if datastore.Commands.Len() > 0 {
		exporters.ExportCommandsToJSON()
	}

	objectsfile.ReadItemsWithAllLocalizations()
	fmt.Printf("[FFX-2] items: %d\n", datastore.Items.Len())
	if datastore.Items.Len() > 0 {
		exporters.ExportItemsToJSON()
	}

	objectsfile.ReadMenuTextWithAllLocalizations()
	fmt.Printf("[FFX-2] menu text: %d\n", objectsfile.MENU_TEXT.Len())
	if datastore.MenuTxt.Len() > 0 {
		exporters.ExportItemsToJSON()
	}

	objectsfile.ReadOversoulWithAllLocalizations()
	fmt.Printf("[FFX-2] oversoul: %d\n", objectsfile.OVERSOUL.Len())
	if objectsfile.OVERSOUL.Len() > 0 {
		exporters.ExportOversoulToJSON()
	}

	objectsfile.ReadBattleTextWithAllLocalizations()
	fmt.Printf("[FFX-2] battle text: %d\n", datastore.BattleTxt.Len())
	if datastore.BattleTxt.Len() > 0 {
		exporters.ExportBattleToJSON()
	}

	objectsfile.ReadBattleEndTextWithAllLocalizations()
	fmt.Printf("[FFX-2] battle end text: %d\n", datastore.BattleEndTxt.Len())
	if datastore.BattleEndTxt.Len() > 0 {
		exporters.ExportBattleEndToJSON()
	}

	objectsfile.ReadPlayerRomTextWithAllLocalizations()
	fmt.Printf("[FFX-2] player room: %d\n", objectsfile.PLAYER_ROOM.Len())
	if objectsfile.PLAYER_ROOM.Len() > 0 {
		exporters.ExportPlayerRoomToJSON()
	}

	// Exclusivos do FFX-2 (retornam cedo se a versão não for FFX-2).
	objectsfile.ReadAAbilityWithAllLocalizations()
	fmt.Printf("[FFX-2] a-ability: %d\n", objectsfile.A_ABILITY.Len())
	if objectsfile.A_ABILITY.Len() > 0 {
		exporters.ExportAAbilityToJSON()
	}

	objectsfile.ReadAccessoriesWithAllLocalizations()
	fmt.Printf("[FFX-2] accessory: %d\n", objectsfile.ACCESSORY.Len())
	if objectsfile.ACCESSORY.Len() > 0 {
		exporters.ExportAccessoriesToJSON()
	}

	objectsfile.ReadJobsWithAllLocalizations()
	fmt.Printf("[FFX-2] job: %d\n", objectsfile.JOB.Len())
	if objectsfile.JOB.Len() > 0 {
		exporters.ExportJobsToJSON()
	}

	objectsfile.ReadMenuTextWithAllLocalizations()
	fmt.Printf("[FFX-2] menu text: %d\n", objectsfile.MENU_TEXT.Len())
	if objectsfile.MENU_TEXT.Len() > 0 {
		exporters.ExportMenuTextToJSON()
	}

	objectsfile.ReadMonsterMagicWithAllLocalizations()
	fmt.Printf("[FFX-2] monster magic: %d\n", objectsfile.MONMAGIC.Len())
	if objectsfile.MONMAGIC.Len() > 0 {
		exporters.ExportMonsterMagicToJSON()
	}

	objectsfile.ReadMonstersWithAllLocalizations()
	fmt.Printf("[FFX-2] monster: %d\n", objectsfile.MONSTER.Len())
	if objectsfile.MONSTER.Len() > 0 {
		exporters.ExportMonstersToJSON()
	}

	objectsfile.ReadMonsters2WithAllLocalizations()
	fmt.Printf("[FFX-2] monster2: %d\n", objectsfile.MONSTER2.Len())
	if objectsfile.MONSTER2.Len() > 0 {
		exporters.ExportMonsters2ToJSON()
	}

	objectsfile.ReadPlateWithAllLocalizations()
	fmt.Printf("[FFX-2] plate: %d\n", objectsfile.PLATE.Len())
	if objectsfile.PLATE.Len() > 0 {
		exporters.ExportPlateToJSON()
	}

	objectsfile.ReadPlayerSaveWithAllLocalizations()
	fmt.Printf("[FFX-2] player save: %d\n", objectsfile.PLAYER_SAVE.Len())
	if objectsfile.PLAYER_SAVE.Len() > 0 {
		exporters.ExportPlayerSaveToJSON()
	}

	objectsfile.ReadSaveTextWithAllLocalizations()
	fmt.Printf("[FFX-2] save text: %d\n", objectsfile.SAVE_TEXT.Len())
	if objectsfile.SAVE_TEXT.Len() > 0 {
		exporters.ExportSaveTextToJSON()
	}

	// Carregar eventos
	if err := readEvents(); err != nil {
		fmt.Printf("Erro ao carregar eventos: %v\n", err)
		return
	}

	//converter.ExportAllEventsToJSON()
	fmt.Println("✓ Eventos exportados para JSON")
	event.ImportEventsDataFromJsonFile()
	fmt.Println("✓ Eventos editados e salvos com sucesso")
	event.ExportAllEventsForLocalizations()
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
	if err := event.ReadAllEventFiles(eventsFolder); err != nil {
		fmt.Printf("Erro ao carregar eventos: %v\n", err)
		return err
	}

	// Contar eventos carregados no datastore
	eventIDs := event.GetAllEventIDs()
	fmt.Printf("✓ Carregados %d eventos no datastore\n", len(eventIDs))
	return nil
}

func showMainMenu() {
	// ===== EXEMPLOS BÁSICOS =====
	writer.ExampleWriteManagerUsage()
	ExportMacroDictionaryExample() // Exemplo de exportação de dicionário de macros (JSON com metadados)
	// ExampleCsvEditorUsage()
	// ExampleJSONEditorUsage()

	// ===== WORKFLOWS COMPLETOS =====
	// ExampleCompleteWorkflow()
	// ExampleCompleteJSONWorkflow()
	// ===== EDIÇÃO DIRECIONADA =====
	//writer.WriteStringsEventForAllLocalizationsJSON("znkd1500", true)
	//ExportMacroDictionaryExample()           // Exemplo de exportação de dicionário de macros
	writer.ExportAllLocalizationsToJSON() // Exporta todos os eventos para JSON
	//writer.WriteStringsEventForAllLocalizationsJSON("akagi0100", true) // Exporta o evento "akagi0100" para JSON
	writer.WriteMacroDictionaryJSON(true)
	//reader.EditAndSaveMacroDictJSONFiles(true) // Exemplo de fluxo completo de exportação/importação
	reader.EditAndSaveSpecificEventFromJSON("znkd1500") // Edita e salva o evento "znkd1500" do JSON
	// ===== COMPARAÇÃO CSV vs JSON =====
	// demoWorkflowCSV()
	// demoWorkflowJSON()
	// exemploEditorEspecificoDemo()	// ===== EXEMPLO ATIVO (descomente para testar) =====
	// Exemplo básico ativo para teste:
	// ExampleSpecificEventJSONEdit("ev001") // Testando a nova função
}
