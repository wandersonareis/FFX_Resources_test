package main

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/core/writer"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/interactions"
	"fmt"
)

/*
MENU PRINCIPAL DE EXEMPLOS
=========================

Este arquivo contém o main() principal que organiza e executa todas as funções de exemplo
dos editores de eventos JSON.
*/
func main() {
	common.LogInfo("=== EXEMPLOS DO SISTEMA DE EVENTOS FFX ===")

	// ===== FFX (v1) =====
	interactions.NewInteractionService().FFXAppConfig().SetGameVersion(common.GameVersionFFX)
	interactions.NewInteractionService().GameLocation.SetTargetDirectory("/home/mestre/FFX_Resources/build/bin/data/") // Defina o caminho correto para os arquivos do jogo
	common.SetVerboseMode(false)                                                                                    // Ativa o modo verboso para depuração

	// Inicialização obrigatória
	common.LogInfo("=== FFX (v1) ===")
	common.LogInfo("Inicializando sistema (FFX v1)...")
	if err := reader.InitializeInternals(); err != nil {
		common.LogError("Erro ao inicializar sistema: %v\n", err)
		return
	}
	// Executa os exemplos para FFX v1
	runFFXExamples()

	// ===== FFX-2 (v2) =====
	interactions.NewInteractionService().FFXAppConfig().SetGameVersion(common.GameVersionFFX2)
	interactions.NewInteractionService().GameLocation.SetTargetDirectory("/home/mestre/FFX_Resources/build/bin/data/") // Defina o caminho correto para os arquivos do jogo

	common.LogInfo("\n=== FFX-2 (v2) ===")
	common.LogInfo("Reinicializando dicionários para FFX-2...")
	if err := reader.InitializeInternals(); err != nil {
		common.LogError("Erro ao inicializar sistema: %v\n", err)
		return
	}
	runFFX2Examples()

	// ===== LastMiss (lastmiss) =====
	interactions.NewInteractionService().FFXAppConfig().SetGameVersion(common.GameVersionLastMiss)

	common.LogInfo("\n=== LastMiss (lastmiss) ===")
	common.LogInfo("Reinicializando dicionários para LastMiss...")
	if err := reader.InitializeInternals(); err != nil {
		common.LogError("Erro ao inicializar sistema: %v\n", err)
		return
	}
	runLastMissExamples()
}

func readEvents() error {
	common.LogInfo("Carregando eventos...")
	eventsFolder, err := common.NewFileAccessor(common.GetPathOriginalsEvent())
	if err != nil {
		return fmt.Errorf("failed to resolve events directory: %w", err)
	}
	version := interactions.CurrentGameVersion()
	if err := event.ReadAllEventFiles(eventsFolder, version); err != nil {
		common.LogVerbose("Erro ao carregar eventos: %v\n", err)
		return err
	}

	// Contar eventos carregados no datastore
	eventIDs := event.GetAllEventIDs(version)
	common.LogInfo("✓ Carregados %d eventos no datastore\n", len(eventIDs))
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