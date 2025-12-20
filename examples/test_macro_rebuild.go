package main

import (
	"ffxresources/backend/core/reader"
	"ffxresources/backend/core/writer"
	"ffxresources/backend/fileFormats/macrodic"
	"ffxresources/backend/sharedutils"
	"fmt"
)

// TestMacroStringRebuild tests the RebuildMacroStrings function
func TestMacroStringRebuild() {
	fmt.Println("=== Teste da Função RebuildMacroStrings ===")

	// Initialize data first
	fmt.Println("Inicializando dados...")
	if err := reader.InitializeInternals(); err != nil {
		fmt.Printf("Erro ao inicializar: %v\n", err)
		return
	}

	// Get test data from MACRODICTFILE
	testLocalization := "us"
	if chunks, exists := macrodic.MACRODICTFILE[testLocalization]; exists && len(chunks) > 0 {
		chunk := chunks[0]
		if len(chunk) > 0 {
			fmt.Printf("Testando com chunk 0 da localização %s (%d strings)\n", testLocalization, len(chunk))

			// Test the rebuild function
			charset := sharedutils.GetCharsetForLanguage(testLocalization)
			rebuiltData := macrodic.GenerateMacroStringData(chunk, charset, true)

			fmt.Printf("Dados reconstruídos: %d bytes\n", len(rebuiltData))

			// Test complete conversion
			completeData := macrodic.MacroStringsToBytes(chunk, charset, true)
			fmt.Printf("Dados completos (com cabeçalho): %d bytes\n", len(completeData))

			// Verify by parsing back
			parsedStrings := macrodic.FromStringData(completeData[2:], charset)
			fmt.Printf("Strings analisadas de volta: %d\n", len(parsedStrings))

			// Compare first few strings
			fmt.Println("Comparando primeiras strings:")
			compareCount := 3
			if len(chunk) < compareCount {
				compareCount = len(chunk)
			}
			if len(parsedStrings) < compareCount {
				compareCount = len(parsedStrings)
			}

			for i := 0; i < compareCount; i++ {
				if chunk[i] != nil && parsedStrings[i] != nil {
					original := chunk[i].GetRegularString()
					parsed := parsedStrings[i].GetRegularString()
					if original == parsed {
						fmt.Printf("  ✓ String %d: '%s'\n", i, original)
					} else {
						fmt.Printf("  ❌ String %d: '%s' != '%s'\n", i, original, parsed)
					}
				}
			}
		}
	} else {
		fmt.Printf("Nenhum dado encontrado para localização %s\n", testLocalization)
	}

	fmt.Println("=== Teste Concluído ===")
}

// TestFullWorkflow tests the complete workflow
func TestFullWorkflow() {
	fmt.Println("=== Teste de Fluxo Completo ===")

	// Test basic rebuild
	TestMacroStringRebuild()

	fmt.Println("\n=== Testando Funcionalidades de Writer ===")

	// Test writer functions
	writer.TestMacroStringReconstruction()

	fmt.Println("=== Todos os Testes Concluídos ===")
}

func RunMacroRebuildTest() {
	TestFullWorkflow()
}
