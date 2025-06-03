package main

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/core/writer"
	"fmt"
	"path/filepath"
)

// ExampleCompleteWorkflow demonstrates the complete CSV export -> edit -> import workflow
func ExampleCompleteWorkflow() {
	fmt.Println("=== Exemplo de Workflow Completo: Exportar -> Editar -> Importar ===")

	// Step 1: Load all events
	fmt.Println("1. Carregando todos os eventos...")
	err := reader.ReadAllEvents(false)
	if err != nil {
		fmt.Printf("Erro ao carregar eventos: %v\n", err)
		return
	}
	fmt.Printf("Carregados %d eventos\n", len(components.EVENTS))

	// Step 2: Export events to CSV using writer package
	fmt.Println("\n2. Exportando eventos para CSV...")
	fmt.Println("Use: writer.WriteEventFileForAllLocalizations(true) para exportar")

	// Step 3: User would manually edit CSV files here
	fmt.Println("\n3. [Etapa Manual] Edite os arquivos CSV em:")
	csvPath := filepath.Join(common.PathFfxRoot, "edits", "events")
	fmt.Printf("   %s\n", csvPath)
	fmt.Println("   - Modifique as traduções nas colunas de idioma")
	fmt.Println("   - Mantenha as colunas 'id' e 'string index' inalteradas")

	// Step 4: Import changes back from CSV
	fmt.Println("\n4. Importando mudanças dos arquivos CSV...")
	err = reader.EditAndSaveEventCSVFiles(true)
	if err != nil {
		fmt.Printf("Erro ao processar CSV: %v\n", err)
		return
	}

	fmt.Println("\n=== Workflow Completo Concluído ===")
	fmt.Println("Os arquivos de evento foram atualizados com as mudanças do CSV")
}

// ExampleSpecificEventJSONEdit demonstrates how to edit a specific event using JSON
func ExampleSpecificEventJSONEdit(eventID string) {
	fmt.Printf("=== Exemplo de Edição de Evento Específico via JSON: %s ===\n", eventID)

	// Step 1: Load all events
	fmt.Println("1. Carregando todos os eventos...")
	err := reader.ReadAllEvents(false)
	if err != nil {
		fmt.Printf("Erro ao carregar eventos: %v\n", err)
		return
	}
	fmt.Printf("Carregados %d eventos\n", len(components.EVENTS))

	if _, exists := components.EVENTS[eventID]; !exists {
		fmt.Printf("Erro: Evento '%s' não encontrado nos eventos carregados\n", eventID)
		fmt.Println("Eventos disponíveis (primeiros 10):")
		count := 0
		for id := range components.EVENTS {
			if count >= 10 {
				fmt.Println("...")
				break
			}
			fmt.Printf("  - %s\n", id)
			count++
		}
		return
	}

	fmt.Println("\n2. Exportando todos os eventos para JSON...")
	fmt.Println("Use: writer.WriteEventFileForAllLocalizationsJSON(true) para exportar")
	writer.WriteEventFileForAllLocalizationsJSON(true)
	jsonPath := filepath.Join(components.GameFilesRoot, components.ModsFolder, "edits", "events", "events_all_localizations.json")
	fmt.Printf("   Arquivo JSON será criado em: %s\n", jsonPath)

	// Step 4: Manual editing step
	fmt.Printf("\n3. [Etapa Manual] Edite o evento '%s' no arquivo JSON:\n", eventID)
	fmt.Printf("   - Abra o arquivo: %s\n", jsonPath)
	fmt.Printf("   - Procure pelo evento com \"id\": \"%s\"\n", eventID)
	fmt.Println("   - Modifique os textos no campo \"text\" para cada idioma")
	fmt.Println("   - Salve o arquivo JSON (certifique-se de que está válido)")

	// Step 5: Import changes for the specific event
	fmt.Printf("\n4. Importando mudanças apenas para o evento '%s'...\n", eventID)
	err = reader.EditAndSaveSpecificEventFromJSON(eventID, true)
	if err != nil {
		fmt.Printf("Erro ao processar evento específico: %v\n", err)
		return
	}

	fmt.Printf("\n=== Edição do Evento '%s' Concluída ===\n", eventID)
	fmt.Println("O evento foi atualizado com as mudanças do JSON")
}
