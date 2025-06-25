package main

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/converter"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/core/writer"
	"fmt"
)

/*
MENU PRINCIPAL DE EXEMPLOS
==========================

Este arquivo contém o main() principal que organiza e executa todas as funções de exemplo
dos editores de eventos CSV e JSON.
*/

func main() {
	fmt.Println("=== EXEMPLOS DO SISTEMA DE EVENTOS FFX ===")
	fmt.Println()

	common.SetGameVersion(1)
	common.SetVerboseMode(true) // Ativa o modo verboso para depuração

	// Inicialização obrigatória
	fmt.Println("Inicializando sistema...")
	if err := reader.InitializeInternals(); err != nil {
		fmt.Printf("Erro ao inicializar sistema: %v\n", err)
		return
	}

	// key items
	// ok
	converter.ReadKeyItemsWithAllLocalizations()
	fmt.Printf("✓ Itens-chave carregados: %d\n", components.KEY_ITEMS.GetLength())
	converter.ExportKeyItemsToJSON()
	fmt.Println("✓ Itens-chave exportados para JSON")
	converter.ProcessKeyItemsJsonFile()
	fmt.Println("✓ Itens-chave editados e salvos com sucesso")
	converter.WriteAllKeyItemsData()
	fmt.Println("✓ Itens-chave salvos com sucesso")

	// commands
	fmt.Println("Carregando comandos...")
	converter.ReadCommandsWithAllLocalizations()
	fmt.Printf("✓ Habilidades carregadas: %d\n", components.COMMANDS.GetLength())
	converter.ExportCommandsToJSON()
	fmt.Println("✓ Habilidades exportadas para JSON")
	converter.ProcessCommandsJsonFile()
	fmt.Println("✓ Comandos editados e salvos com sucesso")
	converter.WriteAllCommandsData()
	fmt.Println("✓ Comandos salvos com sucesso")

	// items
	// ok
	fmt.Println("Carregando itens...")
	converter.ReadItemsWithAllLocalizations()
	fmt.Printf("✓ Itens carregados: %d\n", components.ITEMS.GetLength())
	converter.ExportItemsToJSON()
	fmt.Println("✓ Itens exportados para JSON")
	converter.ProcessItemsJsonFile()
	fmt.Println("✓ Itens editados e salvos com sucesso")
	converter.WriteAllItemsData()
	fmt.Println("✓ Itens salvos com sucesso")

	// arms text
	// ok
	fmt.Println("Carregando textos de armas...")
	converter.ReadArmsTextWithAllLocalizations()
	fmt.Printf("✓ Textos de armas carregados: %d\n", components.ARMS_TEXT.GetLength())
	converter.ExportArmsToJSON()
	fmt.Println("✓ Textos de armas exportados para JSON")
	converter.ProcessArmsJsonFile()
	fmt.Println("✓ Textos de armas editados e salvos com sucesso")
	converter.WriteAllArmsTextData()
	fmt.Println("✓ Textos de armas salvos com sucesso")

	// battle text
	// ok
	converter.ReadBattleTextWithAllLocalizations()
	fmt.Printf("✓ Textos de batalha carregados: %d\n", components.BTL_TEXT.GetLength())
	converter.ExportBattleToJSON()
	fmt.Println("✓ Textos de batalha exportados para JSON")
	converter.ProcessBattleTextJsonFile()
	fmt.Println("✓ Textos de batalha editados e salvos com sucesso")
	converter.WriteAllBattleTextData()
	fmt.Println("✓ Textos de batalha salvos com sucesso")

	// battle end text
	// ok
	converter.ReadBattleEndTextWithAllLocalizations()
	fmt.Printf("✓ Textos de fim de batalha carregados: %d\n", components.BTLEND_TEXT.GetLength())
	converter.ExportBattleEndToJSON()
	fmt.Println("✓ Textos de fim de batalha exportados para JSON")
	converter.ProcessBattleEndTextJsonFile()
	fmt.Println("✓ Textos de fim de batalha editados e salvos com sucesso")
	converter.WriteAllBattleEndTextData()
	fmt.Println("✓ Textos de fim de batalha salvos com sucesso")

	// monster magic 1
	// ok
	fmt.Println("Carregando itens mágicos...")
	converter.ReadMonsterMagic1WithAllLocalizations()
	fmt.Printf("✓ Itens carregados: %d\n", components.MONMAGIC1.GetLength())
	converter.ExportMonsterMagic1ToJSON()
	fmt.Println("✓ Itens exportados para JSON")
	converter.ProcessMonsterMagic1JsonFile()
	fmt.Println("✓ Comandos editados e salvos com sucesso")
	converter.WriteAllMonsterMagic1Data()
	fmt.Println("✓ Itens mágicos salvos com sucesso")

	// monster magic 2
	// ok
	converter.ReadMonsterMagic2WithAllLocalizations()
	fmt.Printf("✓ Itens carregados: %d\n", components.MONMAGIC2.GetLength())
	converter.ExportMonsterMagic2ToJSON()
	fmt.Println("✓ Itens exportados para JSON")
	converter.ProcessMonsterMagic2JsonFile()
	fmt.Println("✓ Comandos editados e salvos com sucesso")
	converter.WriteAllMonsterMagic2Data()
	fmt.Println("✓ Itens mágicos salvos com sucesso")

	// build text
	fmt.Println("Carregando textos de build...")
	converter.ReadBuildTextWithAllLocalizations()
	fmt.Printf("✓ Textos de build carregados: %d\n", components.BUILD_TEXT.GetLength())
	converter.ExportBuildToJSON()
	fmt.Println("✓ Textos de build exportados para JSON")
	converter.ProcessBuildTextJsonFile()
	fmt.Println("✓ Textos de build editados e salvos com sucesso")
	converter.WriteAllBuildTextData()
	fmt.Println("✓ Textos de build salvos com sucesso")

	// config text
	fmt.Println("Carregando textos de configuração...")
	converter.ReadConfigTextWithAllLocalizations()
	fmt.Printf("✓ Textos de configuração carregados: %d\n", components.CONFIG_TEXT.GetLength())
	converter.ExportConfigToJSON()
	fmt.Println("✓ Textos de configuração exportados para JSON")
	converter.ProcessConfigTextJsonFile()
	fmt.Println("✓ Textos de configuração editados e salvos com sucesso")
	converter.WriteAllConfigTextData()
	fmt.Println("✓ Textos de configuração salvos com sucesso")

	// item text
	fmt.Println("Carregando textos de itens...")
	converter.ReadItemCommandsWithAllLocalizations()
	fmt.Printf("✓ Textos de itens carregados: %d\n", components.ITEM_TEXT.GetLength())
	converter.ExportItemCommandsToJSON()
	fmt.Println("✓ Textos de itens exportados para JSON")
	converter.ProcessItemCommandsJsonFile()
	fmt.Println("✓ Textos de itens editados e salvos com sucesso")
	converter.WriteAllItemCommandsData()
	fmt.Println("✓ Textos de itens salvos com sucesso")

	// main menu text
	fmt.Println("Carregando textos do menu principal...")
	converter.ReadMainMenuTextWithAllLocalizations()
	fmt.Printf("✓ Textos do menu principal carregados: %d\n", components.MMAIN_TEXT.GetLength())
	converter.ExportMainMenuToJSON()
	fmt.Println("✓ Textos do menu principal exportados para JSON")
	converter.ProcessMainMenuTextJsonFile()
	fmt.Println("✓ Textos do menu principal editados e salvos com sucesso")
	converter.WriteAllMainMenuTextData()
	fmt.Println("✓ Textos do menu principal salvos com sucesso")

	// player blitsball text
	fmt.Println("Carregando textos de jogador de blitsball...")
	converter.ReadPlayerRomTextWithAllLocalizations()
	fmt.Printf("✓ Textos de jogador de blitsball carregados: %d\n", components.PLAYER_ROOM.GetLength())
	converter.ExportPlayerRoomToJSON()
	fmt.Println("✓ Textos de jogador de blitsball exportados para JSON")
	converter.ProcessPlayerRoomTextJsonFile()
	fmt.Println("✓ Textos de jogador de blitsball editados e salvos com sucesso")
	converter.WriteAllPlayerRoomTextData()
	fmt.Println("✓ Textos de jogador de blitsball salvos com sucesso")

	// name text
	fmt.Println("Carregando textos de nomes...")
	converter.ReadNameTextWithAllLocalizations()
	fmt.Printf("✓ Textos de nomes carregados: %d\n", components.NAME_TEXT.GetLength())
	converter.ExportNameToJSON()
	fmt.Println("✓ Textos de nomes exportados para JSON")
	converter.ProcessNameTextJsonFile()
	fmt.Println("✓ Textos de nomes editados e salvos com sucesso")
	converter.WriteAllNameTextData()
	fmt.Println("✓ Textos de nomes salvos com sucesso")

	// Carregar eventos primeiro
	if err := readEvents(); err != nil {
		fmt.Printf("Erro ao carregar eventos: %v\n", err)
		return
	}

	//converter.ExportAllEventsToJSON()
	fmt.Println("✓ Eventos exportados para JSON")
	converter.ImportEventsDataFromJsonFile()
	fmt.Println("✓ Eventos editados e salvos com sucesso")
	converter.ExportAllEventsForLocalizations()
	fmt.Println("✓ Eventos salvos com sucesso")

	showMainMenu()
}

func readEvents() error {
	fmt.Println("Carregando eventos...")
	eventsFolder, err := common.NewFileAccessor(common.GetPathOriginalsEvent())
	if err != nil {
		return fmt.Errorf("failed to resolve events directory: %w", err)
	}
	if err := converter.ReadAllEventFiles(eventsFolder); err != nil {
		fmt.Printf("Erro ao carregar eventos: %v\n", err)
		return err
	}
	fmt.Printf("✓ Carregados %d eventos\n", len(components.EVENTS))
	return nil
}

func showMainMenu() {
	// ===== EXEMPLOS BÁSICOS =====
	writer.ExampleWriteManagerUsage()
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
	//writer.WriteMacroDictionaryJSON(true)
	//reader.EditAndSaveMacroDictJSONFiles(true) // Exemplo de fluxo completo de exportação/importação
	reader.EditAndSaveSpecificEventFromJSON("znkd1500") // Edita e salva o evento "znkd1500" do JSON
	// ===== COMPARAÇÃO CSV vs JSON =====
	// demoWorkflowCSV()
	// demoWorkflowJSON()
	// exemploEditorEspecificoDemo()	// ===== EXEMPLO ATIVO (descomente para testar) =====
	// Exemplo básico ativo para teste:
	// ExampleSpecificEventJSONEdit("ev001") // Testando a nova função
}
