package main

import (
	encoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/core/writer"
	"ffxresources/backend/fileFormats/macrodic"
	"ffxresources/backend/interactions"
	"fmt"
)

// TestMacroStringRebuild tests the container-based rebuild roundtrip:
// binary -> TextFile -> ToBytes -> reparse, comparing texts.
func TestMacroStringRebuild() {
	fmt.Println("=== Teste de Reconstrução de Macro Dictionary ===")

	// Initialize data first
	fmt.Println("Inicializando dados...")
	if err := reader.InitializeInternals(); err != nil {
		fmt.Printf("Erro ao inicializar: %v\n", err)
		return
	}

	// Get test data from available containers
	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	containers, err := macrodic.ReadMacroDictionaryContainers(version)
	if err != nil {
		fmt.Printf("Erro ao ler containers: %v\n", err)
		return
	}
	testLocalization := "us"
	c, ok := containers[testLocalization]
	if !ok {
		fmt.Printf("Nenhum dado encontrado para localização %s\n", testLocalization)
		return
	}

	// Use the first non-empty chunk
	var chunkIndex int = -1
	for i := range c.ChunkOffsets {
		f, err := c.FileAt(i)
		if err != nil || len(f.Segments) == 0 {
			continue
		}
		chunkIndex = i
		break
	}
	if chunkIndex < 0 {
		fmt.Printf("Nenhum chunk com dados para localização %s\n", testLocalization)
		return
	}

	f, _ := c.FileAt(chunkIndex)
	fmt.Printf("Testando com chunk %d da localização %s (%d segmentos)\n", chunkIndex, testLocalization, len(f.Segments))

	// Rebuild and reparse
	rebuilt, err := f.ToBytes()
	if err != nil {
		fmt.Printf("Erro ao reconstruir: %v\n", err)
		return
	}
	fmt.Printf("Dados reconstruídos: %d bytes\n", len(rebuilt))

	charset := encoding.GetCharsetForLanguage(testLocalization)
	reparsed, err := macrodic.NewMacroDictionaryTextFile(rebuilt, charset, c.Version)
	if err != nil {
		fmt.Printf("Erro ao reler: %v\n", err)
		return
	}
	fmt.Printf("Segmentos analisados de volta: %d\n", len(reparsed.Segments))

	// Compare first few strings
	fmt.Println("Comparando primeiras strings:")
	compareCount := 3
	if len(f.Segments) < compareCount {
		compareCount = len(f.Segments)
	}
	if len(reparsed.Segments) < compareCount {
		compareCount = len(reparsed.Segments)
	}

	origStrings := f.ToMacroStrings()
	newStrings := reparsed.ToMacroStrings()
	for i := 0; i < compareCount; i++ {
		if origStrings[i] != nil && newStrings[i] != nil {
			original := origStrings[i].GetRegularString()
			parsed := newStrings[i].GetRegularString()
			if original == parsed {
				fmt.Printf("  ✓ String %d: '%s'\n", i, original)
			} else {
				fmt.Printf("  ❌ String %d: '%s' != '%s'\n", i, original, parsed)
			}
		}
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
	writer.ExampleMacroDictionaryUsage()

	fmt.Println("=== Todos os Testes Concluídos ===")
}

func RunMacroRebuildTest() {
	TestFullWorkflow()
}
