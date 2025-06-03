package main

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/reader"
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

	// Inicialização obrigatória
	fmt.Println("Inicializando sistema...")
	reader.InitializeInternals()

	// Carregar eventos primeiro
	fmt.Println("Carregando eventos...")
	err := reader.ReadAllEvents(false)
	if err != nil {
		fmt.Printf("Erro ao carregar eventos: %v\n", err)
		return
	}
	fmt.Printf("✓ Carregados %d eventos\n", len(components.EVENTS))
	fmt.Println()

	// Menu principal
	showMainMenu()
}

func showMainMenu() {
	fmt.Println("=== MENU DE EXEMPLOS ===")
	fmt.Println("Escolha uma das opções abaixo (descomente a linha correspondente):")
	fmt.Println()

	fmt.Println("--- EXEMPLOS BÁSICOS ---")
	fmt.Println("1. CSV Editor Básico:")
	fmt.Println("   // ExampleCsvEditorUsage()")
	fmt.Println()

	fmt.Println("2. JSON Editor Básico:")
	fmt.Println("   // ExampleJSONEditorUsage()")
	fmt.Println()

	fmt.Println("--- WORKFLOWS COMPLETOS ---")
	fmt.Println("3. Workflow CSV Completo (Exportar → Editar → Importar):")
	fmt.Println("   // ExampleCompleteWorkflow()")
	fmt.Println()

	fmt.Println("4. Workflow JSON Completo (Exportar → Editar → Importar):")
	fmt.Println("   // ExampleCompleteJSONWorkflow()")
	fmt.Println()
	fmt.Println("--- EDIÇÃO DIRECIONADA ---")
	fmt.Println("5. Editar Evento Específico via CSV:")
	fmt.Println("   // ExampleTargetedEventEdit(\"ev001\")")
	fmt.Println()

	fmt.Println("6. Editar Evento Específico via JSON:")
	fmt.Println("   // ExampleTargetedEventJSONEdit(\"ev001\")")
	fmt.Println()
	fmt.Println("7. Editar Evento Específico via JSON (Nova Função):")
	fmt.Println("   // ExampleSpecificEventJSONEdit(\"ev001\")")
	fmt.Println()
	fmt.Println("--- COMPARAÇÃO CSV vs JSON ---")
	fmt.Println("8. Demonstração CSV vs JSON:")
	fmt.Println("   // demoWorkflowCSV()")
	fmt.Println("   // demoWorkflowJSON()")
	fmt.Println()

	fmt.Println("9. Edição Específica Comparativa:")
	fmt.Println("   // exemploEditorEspecificoDemo()")
	fmt.Println()

	fmt.Println("--- EXEMPLOS PRONTOS PARA EXECUTAR ---")
	fmt.Println("Descomente uma das linhas abaixo para executar:")

	// ===== EXEMPLOS BÁSICOS =====
	// ExampleCsvEditorUsage()
	// ExampleJSONEditorUsage()

	// ===== WORKFLOWS COMPLETOS =====
	// ExampleCompleteWorkflow()
	// ExampleCompleteJSONWorkflow()
	// ===== EDIÇÃO DIRECIONADA =====
	// ExampleTargetedEventEdit("ev001")  // Substitua "ev001" por um ID real
	// ExampleTargetedEventJSONEdit("ev001")  // Substitua "ev001" por um ID real
	ExampleSpecificEventJSONEdit("azit0300")  // Nova função - Substitua "ev001" por um ID real

	// ===== COMPARAÇÃO CSV vs JSON =====
	// demoWorkflowCSV()
	// demoWorkflowJSON()
	// exemploEditorEspecificoDemo()	// ===== EXEMPLO ATIVO (descomente para testar) =====
	// Exemplo básico ativo para teste:
	// ExampleSpecificEventJSONEdit("ev001") // Testando a nova função
	showEventStats()
}

// showEventStats exibe estatísticas básicas dos eventos carregados
func showEventStats() {
	fmt.Println("=== ESTATÍSTICAS DOS EVENTOS ===")
	fmt.Printf("Total de eventos carregados: %d\n", len(components.EVENTS))

	if len(components.EVENTS) > 0 {
		fmt.Println("\nPrimeiros 5 eventos encontrados:")
		count := 0
		for eventID, eventFile := range components.EVENTS {
			if count >= 5 {
				break
			}
			if eventFile != nil {
				fmt.Printf("  %s (%d strings)\n", eventID, len(eventFile.Strings))
			}
			count++
		}

		fmt.Println("\nPara executar exemplos específicos, edite este arquivo e descomente")
		fmt.Println("as funções que deseja testar na função main().")
	}
}
