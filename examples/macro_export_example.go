package main

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/core/writer"
	"ffxresources/backend/fileFormats/macrodic"
	"ffxresources/backend/interactions"
	"fmt"
	"log"
	"path/filepath"
)

// ExportMacroDictionaryExample demonstrates how to load and export macro dictionary data
func ExportMacroDictionaryExample() {
	common.LogInfo("=== Exemplo de Exportação de Dicionário de Macros ===")

	// Step 1: Initialize internals to load character maps and macro dictionaries
	common.LogInfo("1. Inicializando dados internos (carregando mapas de caracteres e dicionários de macros)...")
	err := reader.InitializeInternals()
	if err != nil {
		log.Printf("Erro ao inicializar dados internos: %v", err)
		return
	}
	common.LogInfo("   ✓ Dados internos carregados com sucesso")

	// Step 2: Export macro dictionary for all localizations
	common.LogInfo("\n2. Exportando dicionário de macros para todas as localizações:")
	writer.ExportMacroDictionaryToJSON()

	// Step 3: Export macro dictionary for specific localizations
	common.LogInfo("\n3. Exportando dicionário de macros para localizações específicas:")

	// English (US)
	fmt.Println("   - Exportando para inglês (US)...")
	writer.WriteMacroDictionaryForLocalizationJSON("us")

	// Japanese
	fmt.Println("   - Exportando para japonês...")
	writer.WriteMacroDictionaryForLocalizationJSON("jp")

	// German
	fmt.Println("   - Exportando para alemão...")
	writer.WriteMacroDictionaryForLocalizationJSON("de")

	// French
	fmt.Println("   - Exportando para francês...")
	writer.WriteMacroDictionaryForLocalizationJSON("fr")

	// Italian
	fmt.Println("   - Exportando para italiano...")
	writer.WriteMacroDictionaryForLocalizationJSON("it")

	// Spanish
	fmt.Println("   - Exportando para espanhol...")
	writer.WriteMacroDictionaryForLocalizationJSON("sp")

	fmt.Println("\n=== Exportação de dicionário de macros concluída ===")
	fmt.Println("\nOs arquivos JSON foram salvos no diretório: edits/macrodic/")
	fmt.Println("- macro_dictionary_all_localizations.json (todas as localizações)")
	fmt.Println("- macro_dictionary_[código].json (localizações específicas)")
}

// ImportMacroDictionaryExample demonstrates how to load macro dictionary from JSON files
func ImportMacroDictionaryExample() {
	fmt.Println("=== Exemplo de Importação de Dicionário de Macros ===")

	// Example 1: Load all localizations from the combined JSON file
	fmt.Println("1. Carregando todas as localizações do arquivo JSON combinado:")
	macrodicPath := filepath.Join("edits", "macrodic")
	allLocalizationsFile := filepath.Join(macrodicPath, "macro_dictionary_all_localizations.json")

	if err := writer.EditAndSaveMacrodicFromJson(allLocalizationsFile); err != nil {
		fmt.Printf("Erro ao carregar arquivo de todas as localizações: %v\n", err)
	} else {
		fmt.Println("   ✓ Todas as localizações carregadas com sucesso")
	}

	// Example 2: Load specific localization files
	fmt.Println("\n2. Carregando localizações específicas:")
	localizations := []string{"us", "jp", "de", "fr", "it", "sp"}

	for _, loc := range localizations {
		fileName := fmt.Sprintf("macro_dictionary_%s.json", loc)
		filePath := filepath.Join(macrodicPath, fileName)

		fmt.Printf("   - Carregando %s...\n", loc)
		if err := writer.EditAndSaveMacrodicFromJson(filePath); err != nil {
			fmt.Printf("     Erro ao carregar %s: %v\n", loc, err)
		} else {
			fmt.Printf("     ✓ %s carregado com sucesso\n", loc)
		}
	}

	fmt.Println("\n=== Importação de dicionário de macros concluída ===")
}

// CompleteWorkflowExample demonstrates the complete export/import workflow
func CompleteWorkflowExample() {
	fmt.Println("=== Exemplo de Fluxo Completo: Exportação → Modificação → Importação ===")

	// Step 1: Initialize data
	fmt.Println("1. Inicializando dados do jogo...")
	if err := reader.InitializeInternals(); err != nil {
		fmt.Printf("Erro ao inicializar dados: %v\n", err)
		return
	}
	fmt.Println("   ✓ Dados inicializados")

	// Step 2: Export to JSON
	fmt.Println("\n2. Exportando para JSON...")
	ExportMacroDictionaryExample()

	// Step 3: Simulate modification by clearing and reloading
	fmt.Println("\n3. Simulando modificação (limpando dados atuais)...")
	// In real usage, user would edit the JSON files here

	// Step 4: Import from JSON
	fmt.Println("\n4. Reimportando do JSON...")
	ImportMacroDictionaryExample()

	fmt.Println("\n✓ Fluxo completo concluído com sucesso!")
}

// BinaryReconstructionExample demonstrates the complete workflow including binary reconstruction
func BinaryReconstructionExample() {
	common.LogInfo("=== Exemplo de Reconstrução Binária de Macro Dictionary ===")

	// Step 1: Initialize data
	common.LogInfo("1. Inicializando dados do jogo...")
	if err := reader.InitializeInternals(); err != nil {
		common.LogError("Erro ao inicializar dados: %v\n", err)
		return
	}
	common.LogInfo("   ✓ Dados inicializados")

	// Step 2: Export to JSON
	common.LogInfo("\n2. Exportando para JSON...")
	writer.ExportMacroDictionaryToJSON()

	// Step 3: Test round-trip conversion
	common.LogInfo("\n3. Testando conversão completa (JSON → Binary → parse)...")
	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	containers, err := macrodic.ReadMacroDictionaryContainers(version)
	if err != nil {
		common.LogError("Erro ao ler containers: %v\n", err)
		return
	}
	js, err := macrodic.MarshalToJson(macrodic.ExportToJson(containers))
	if err != nil {
		common.LogError("Erro ao exportar JSON: %v\n", err)
		return
	}
	imp, err := macrodic.UnmarshalJson(js)
	if err != nil {
		common.LogError("Erro ao ler JSON: %v\n", err)
		return
	}
	back, err := macrodic.ImportFromJson(imp, version)
	if err != nil {
		common.LogError("Erro ao importar JSON: %v\n", err)
		return
	}
	common.LogInfo("   ✓ %d localizações reconstruídas\n", len(back))

	// Step 4: Write binary files
	common.LogInfo("\n4. Criando arquivos binários reconstruídos...")
	if err := macrodic.SaveMacroDictionaryBinaries(back); err != nil {
		common.LogError("Erro ao salvar binários: %v\n", err)
		return
	}

	common.LogInfo("\n✓ Exemplo de reconstrução binária concluído!")
	common.LogInfo("Arquivos criados em:")
	common.LogInfo("  - edits/macrodic/*.json (formatos JSON)")
	common.LogInfo("  - edits/macrodic/binary/*.dcp (formatos binários reconstruídos)")
}
