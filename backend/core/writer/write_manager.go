package writer

import (
	"encoding/json"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/reader"
	"fmt"
	"path/filepath"
	"sort"
)

// getLocalizationKeys returns all available localization keys
// This function returns the localization keys from the common package
func getLocalizationKeys() []string {
	var keys []string
	for key := range common.Localizations {
		keys = append(keys, key)
	}
	return keys
}

func ExampleWriteManagerUsage() {
	fmt.Println("=== Write Manager Usage Example ===")

	// Ensure events are loaded first
	err := reader.ReadAllEvents(false) // false = don't skip blitzball
	if err != nil {
		fmt.Printf("Error loading events: %v\n", err)
		return
	}

	// Example 1: Write CSV files for all localizations
	fmt.Println("\n1. Writing CSV files for all localizations:")
	WriteEventFileForAllLocalizations(true)

	// Example 2: Write JSON files for all localizations
	fmt.Println("\n2. Writing JSON files for all localizations:")
	WriteEventFileForAllLocalizationsJSON(true)

	// Example 3: Write CSV files for a specific localization
	fmt.Println("\n3. Writing CSV files for Japanese localization:")
	WriteEventFileForLocalization("jp", true)

	// Example 4: Write JSON files for a specific localization
	fmt.Println("\n4. Writing JSON files for Japanese localization:")
	WriteEventFileForLocalizationJSON("jp", true)

	// Example 5: Write files for English localization (both formats)
	fmt.Println("\n5. Writing files for English localization:")
	WriteEventFileForLocalization("us", true)
	WriteEventFileForLocalizationJSON("us", true)

	fmt.Println("\n=== Write operations completed ===")
}

type MacroStringData struct {
	Index          int    `json:"index"`
	RegularText    string `json:"regular_text"`
	SimplifiedText string `json:"simplified_text"`
	HasDistinct    bool   `json:"has_distinct_simplified"`
}

type MacroChunkData struct {
	ChunkIndex int               `json:"chunk_index"`
	Strings    []MacroStringData `json:"strings"`
}

type MacroLocalizationData struct {
	Localization string           `json:"localization"`
	Chunks       []MacroChunkData `json:"chunks"`
}

// WriteMacroDictionaryJSON writes macro dictionary files as JSON for all localizations
// Creates JSON files with macro strings for each language in the edits/macrodic/ directory
//
// Parameters:
//   - print: If true, prints the exported file paths for debugging
//
// JSON Format:
//   - Array of localization objects, each containing chunks with macro strings
//   - Each macro string has regular and simplified text variations
//   - Only exports localizations that have macro data (skips empty ones)
func WriteMacroDictionaryJSON(print bool) {
	path := filepath.Join(components.GameFilesRoot, components.ModsFolder, "edits", "macrodic")

	// Ensure the output directory exists
	if err := common.EnsurePathExists(path); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Get all localization keys and sort them for consistent output
	localizationKeys := getLocalizationKeys()
	sort.Strings(localizationKeys)

	var allLocalizations []MacroLocalizationData

	// Process each localization
	for _, localizationKey := range localizationKeys {
		if print {
			fmt.Printf("Processing macro dictionary for localization: %s\n", localizationKey)
		}

		// Check if this localization has macro data
		chunks, exists := components.MACRODICTFILE[localizationKey]
		if !exists || len(chunks) == 0 {
			if print {
				fmt.Printf("No macro data found for localization: %s\n", localizationKey)
			}
			continue
		}

		// Create localization data structure
		localizationData := MacroLocalizationData{
			Localization: localizationKey,
			Chunks:       make([]MacroChunkData, 0),
		}

		// Process each chunk
		for chunkIndex, chunk := range chunks {
			if len(chunk) == 0 {
				continue
			}

			// Create chunk data structure
			chunkData := MacroChunkData{
				ChunkIndex: chunkIndex,
				Strings:    make([]MacroStringData, 0, len(chunk)),
			}

			for stringIndex, macroString := range chunk {
				if macroString == nil {
					continue
				}

				stringData := MacroStringData{
					Index:       stringIndex,
					RegularText: macroString.GetRegularString(),
					HasDistinct: macroString.HasDistinctSimplified(),
				}

				if stringData.HasDistinct {
					stringData.SimplifiedText = macroString.GetSimplifiedString()
				}

				chunkData.Strings = append(chunkData.Strings, stringData)
			}

			if len(chunkData.Strings) > 0 {
				localizationData.Chunks = append(localizationData.Chunks, chunkData)
			}
		}

		// Only add localization if it has chunks
		if len(localizationData.Chunks) > 0 {
			allLocalizations = append(allLocalizations, localizationData)
		}
	}

	// Skip if no localizations were processed
	if len(allLocalizations) == 0 {
		fmt.Println("No macro dictionary data found to export to JSON")
		return
	}

	// Marshal to JSON with proper formatting
	jsonData, err := json.MarshalIndent(allLocalizations, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling macro dictionary to JSON: %v\n", err)
		return
	}

	// Write JSON file
	fileName := "macro_dictionary_all_localizations.json"
	filePath := filepath.Join(path, fileName)

	err = components.WriteStringToFile(filePath, string(jsonData))
	if err != nil {
		fmt.Printf("Error writing JSON file %s: %v\n", filePath, err)
		return
	}

	if print {
		fmt.Printf("Arquivo JSON de dicionário de macros exportado: %s\n", filePath)
		fmt.Printf("Total de localizações exportadas: %d\n", len(allLocalizations))

		// Print summary for each localization
		for _, locData := range allLocalizations {
			fmt.Printf("  - %s: %d chunks\n", locData.Localization, len(locData.Chunks))
		}
	}
}

func WriteMacroDictionaryForLocalizationJSON(localization string, print bool) {
	path := filepath.Join(components.GameFilesRoot, components.ModsFolder, "edits", "macrodic")

	// Ensure the output directory exists
	if err := common.EnsurePathExists(path); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Check if this localization has macro data
	chunks, exists := components.MACRODICTFILE[localization]
	if !exists || len(chunks) == 0 {
		fmt.Printf("No macro data found for localization: %s\n", localization)
		return
	}

	if print {
		fmt.Printf("Processing macro dictionary for localization: %s\n", localization)
	}

	// Create localization data structure
	localizationData := MacroLocalizationData{
		Localization: localization,
		Chunks:       make([]MacroChunkData, 0),
	}

	// Process each chunk
	for chunkIndex, chunk := range chunks {
		if len(chunk) == 0 {
			continue // Skip empty chunks
		}

		// Create chunk data structure
		chunkData := MacroChunkData{
			ChunkIndex: chunkIndex,
			Strings:    make([]MacroStringData, 0, len(chunk)),
		}

		// Process each macro string in the chunk
		for stringIndex, macroString := range chunk {
			if macroString == nil {
				continue
			}

			// Create string data structure
			stringData := MacroStringData{
				Index:       stringIndex,
				RegularText: macroString.GetRegularString(),
				HasDistinct: macroString.HasDistinctSimplified(),
			}

			if stringData.HasDistinct {
				stringData.SimplifiedText = macroString.GetSimplifiedString()
			}

			chunkData.Strings = append(chunkData.Strings, stringData)
		}

		if len(chunkData.Strings) > 0 {
			localizationData.Chunks = append(localizationData.Chunks, chunkData)
		}
	}

	// Skip if no chunks were processed
	if len(localizationData.Chunks) == 0 {
		fmt.Printf("No macro dictionary chunks found for localization: %s\n", localization)
		return
	}

	// Marshal to JSON with proper formatting
	jsonData, err := json.MarshalIndent(localizationData, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling macro dictionary to JSON: %v\n", err)
		return
	}

	// Write JSON file
	fileName := fmt.Sprintf("macro_dictionary_%s.json", localization)
	filePath := filepath.Join(path, fileName)

	err = components.WriteStringToFile(filePath, string(jsonData))
	if err != nil {
		fmt.Printf("Error writing JSON file %s: %v\n", filePath, err)
		return
	}

	if print {
		fmt.Printf("Arquivo JSON de dicionário de macros exportado (%s): %s\n", localization, filePath)
		fmt.Printf("Total de chunks exportados: %d\n", len(localizationData.Chunks))
	}
}

// ExampleMacroDictionaryUsage demonstrates how to use the macro dictionary export functions
func ExampleMacroDictionaryUsage() {
	fmt.Println("=== Exemplo de Exportação de Dicionário de Macros ===")

	// Example 1: Export macro dictionary for all localizations
	fmt.Println("\n1. Exportando dicionário de macros para todas as localizações:")
	WriteMacroDictionaryJSON(true)

	// Example 2: Export macro dictionary for a specific localization
	fmt.Println("\n2. Exportando dicionário de macros para localização japonesa:")
	WriteMacroDictionaryForLocalizationJSON("jp", true)

	// Example 3: Export macro dictionary for English localization
	fmt.Println("\n3. Exportando dicionário de macros para localização inglesa:")
	WriteMacroDictionaryForLocalizationJSON("us", true)

	fmt.Println("\n=== Exportação de dicionário de macros concluída ===")
}

// EditAndSaveMacrodicFromJson reads JSON files and reconstructs the components.MACRODICTFILE
// This function can read both single localization files and all localizations file
//
// Parameters:
//   - jsonFilePath: Path to the JSON file to read
//   - print: If true, prints debug information during the process
//
// JSON Format Expected:
//   - Single localization: MacroLocalizationData object
//   - All localizations: Array of MacroLocalizationData objects
//
// The function will:
//  1. Read and parse the JSON file
//  2. Clear existing MACRODICTFILE data for the localizations being updated
//  3. Reconstruct MacroString objects from JSON data
//  4. Update the global MACRODICTFILE with the new data
func EditAndSaveMacrodicFromJson(jsonFilePath string, print bool) error {
	if print {
		fmt.Printf("Carregando dados do dicionário de macros do arquivo: %s\n", jsonFilePath)
	}
	// Read JSON file
	resolvedFile, err := components.ResolveFile(jsonFilePath, true)
	if err != nil {
		return fmt.Errorf("erro ao resolver caminho do arquivo JSON: %v", err)
	}
	jsonData, err := common.ReadFile(resolvedFile)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo JSON: %v", err)
	}
	// Try to parse as array of localizations first (all localizations file)
	var allLocalizations []MacroLocalizationData
	if err := json.Unmarshal(jsonData, &allLocalizations); err != nil {
		// If that fails, try to parse as single localization
		var singleLocalization MacroLocalizationData
		if err := json.Unmarshal(jsonData, &singleLocalization); err != nil {
			return fmt.Errorf("erro ao fazer parse do JSON: %v", err)
		}
		// Convert single localization to array
		allLocalizations = []MacroLocalizationData{singleLocalization}
	}

	if print {
		fmt.Printf("Encontradas %d localizações no arquivo JSON\n", len(allLocalizations))
	}

	MacroDic := map[string][][]*components.MacroString{}

	// Process each localization
	for _, locData := range allLocalizations {
		if print {
			fmt.Printf("Processando localização: %s (%d chunks)\n", locData.Localization, len(locData.Chunks))
		}

		// Clear existing data for this localization
		components.MACRODICTFILE[locData.Localization] = make([][]*components.MacroString, 0)

		// Find the maximum chunk index to properly size the array
		maxChunkIndex := 15

		// Initialize the chunks array with proper size
		chunks := make([][]*components.MacroString, maxChunkIndex+1)
		for i := range chunks {
			chunks[i] = make([]*components.MacroString, 0)
		}

		// Process each chunk
		for _, chunkData := range locData.Chunks {
			chunkIndex := chunkData.ChunkIndex

			if print {
				fmt.Printf("  Processando chunk %d (%d strings)\n", chunkIndex, len(chunkData.Strings))
			}

			// Find the maximum string index to properly size the chunk
			maxStringIndex := -1
			for _, stringData := range chunkData.Strings {
				if stringData.Index > maxStringIndex {
					maxStringIndex = stringData.Index
				}
			}

			// Initialize the strings array for this chunk
			if maxStringIndex >= 0 {
				chunks[chunkIndex] = make([]*components.MacroString, maxStringIndex+1)
			} // Process each string in the chunk
			for _, stringData := range chunkData.Strings {
				stringIndex := stringData.Index

				// Get the text for this localization
				regularText := stringData.RegularText
				simplifiedText := stringData.SimplifiedText

				// If no simplified text is provided, use regular text
				if simplifiedText == "" {
					simplifiedText = regularText
				}

				// Convert strings back to bytes using the localization's charset
				charset := components.LocalizationToCharset(locData.Localization)
				regularBytes := components.StringToBytes(regularText, charset)
				simplifiedBytes := components.StringToBytes(simplifiedText, charset)

				// Create MacroString object
				macroString := &components.MacroString{
					Charset:          charset,
					RegularOffset:    0, // These offsets are not relevant when reconstructing from JSON
					SimplifiedOffset: 0, // They are used during binary parsing only
					RegularBytes:     regularBytes,
					SimplifiedBytes:  simplifiedBytes,
				}

				// Add to chunk (ensure the array is large enough)
				if stringIndex >= len(chunks[chunkIndex]) {
					// Expand the array to accommodate this index
					newSize := stringIndex + 1
					newSlice := make([]*components.MacroString, newSize)
					copy(newSlice, chunks[chunkIndex])
					chunks[chunkIndex] = newSlice
				}
				chunks[chunkIndex][stringIndex] = macroString
			}
		}

		// Update the global MACRODICTFILE
		MacroDic[locData.Localization] = chunks
		components.MACRODICTFILE[locData.Localization] = chunks

		if print {
			fmt.Printf("  ✓ Localização %s atualizada com sucesso\n", locData.Localization)
		}
	}

	if print {
		fmt.Printf("Dicionário de macros reconstruído com sucesso!\n")
		fmt.Printf("MACRODICTFILE agora contém %d localizações\n", len(MacroDic))
	}

	return nil
}

// LoadMacrodicFromJsonExample demonstrates how to load macro dictionary from JSON
func LoadMacrodicFromJsonExample() {
	fmt.Println("=== Exemplo de Carregamento de Dicionário de Macros do JSON ===")

	// Path to the JSON files
	macrodicPath := filepath.Join(components.GameFilesRoot, components.ModsFolder, "edits", "macrodic")

	// Example 1: Load all localizations from the combined JSON file
	fmt.Println("\n1. Carregando todas as localizações do arquivo JSON combinado:")
	allLocalizationsFile := filepath.Join(macrodicPath, "macro_dictionary_all_localizations.json")
	if err := EditAndSaveMacrodicFromJson(allLocalizationsFile, true); err != nil {
		fmt.Printf("Erro ao carregar arquivo de todas as localizações: %v\n", err)
	}

	// Example 2: Load specific localization files
	fmt.Println("\n2. Carregando localizações específicas:")

	localizations := []string{"us", "jp", "de", "fr", "it", "sp"}
	for _, loc := range localizations {
		fileName := fmt.Sprintf("macro_dictionary_%s.json", loc)
		filePath := filepath.Join(macrodicPath, fileName)

		fmt.Printf("   - Carregando %s...\n", loc)
		if err := EditAndSaveMacrodicFromJson(filePath, false); err != nil {
			fmt.Printf("     Erro ao carregar %s: %v\n", loc, err)
		} else {
			fmt.Printf("     ✓ %s carregado com sucesso\n", loc)
		}
	}

	fmt.Println("\n=== Carregamento de dicionário de macros concluído ===")
}

// CompleteMacroDictionaryWorkflowExample demonstrates the complete workflow:
// 1. Initialize and load macro data from game files
// 2. Export to JSON
// 3. Modify JSON (simulated)
// 4. Load back from JSON
func CompleteMacroDictionaryWorkflowExample() {
	fmt.Println("=== Exemplo de Fluxo Completo de Dicionário de Macros ===")

	// Step 1: Initialize macro data from game files
	fmt.Println("\n1. Inicializando dados do jogo...")
	if err := reader.InitializeInternals(); err != nil {
		fmt.Printf("Erro ao inicializar dados: %v\n", err)
		return
	}
	fmt.Printf("   ✓ Dados inicializados. MACRODICTFILE contém %d localizações\n", len(components.MACRODICTFILE))
	// Step 2: Export to JSON
	fmt.Println("\n2. Exportando para JSON...")
	WriteMacroDictionaryJSON(true)

	// Also export individual localizations
	localizations := []string{"us", "jp"}
	for _, loc := range localizations {
		WriteMacroDictionaryForLocalizationJSON(loc, false)
	}

	// Step 3: Simulate JSON modification (in real use, user would edit the JSON files)
	fmt.Println("\n3. Simulando modificação do JSON (em uso real, o usuário editaria os arquivos)...")
	fmt.Println("   (Neste exemplo, vamos apenas recarregar os dados sem modificações)")

	// Step 4: Clear current data and reload from JSON
	fmt.Println("\n4. Limpando dados atuais e recarregando do JSON...")

	// Clear existing data to simulate a fresh start
	components.MACRODICTFILE = make(map[string][][]*components.MacroString)
	fmt.Printf("   MACRODICTFILE limpo. Agora contém %d localizações\n", len(components.MACRODICTFILE))

	// Reload from JSON
	macrodicPath := filepath.Join(components.GameFilesRoot, components.ModsFolder, "edits", "macrodic")
	allLocalizationsFile := filepath.Join(macrodicPath, "macro_dictionary_all_localizations.json")

	if err := EditAndSaveMacrodicFromJson(allLocalizationsFile, true); err != nil {
		fmt.Printf("Erro ao recarregar do JSON: %v\n", err)
		return
	}

	// Step 5: Verify the data was loaded correctly
	fmt.Println("\n5. Verificando integridade dos dados recarregados...")
	totalStrings := 0
	for localization, chunks := range components.MACRODICTFILE {
		localizationStrings := 0
		for _, chunk := range chunks {
			for _, macroString := range chunk {
				if macroString != nil && !macroString.IsEmpty() {
					localizationStrings++
				}
			}
		}
		totalStrings += localizationStrings
		fmt.Printf("   - %s: %d chunks, %d strings\n", localization, len(chunks), localizationStrings)
	}
	fmt.Printf("\n✓ Fluxo completo finalizado com sucesso!")
	fmt.Printf("\n✓ Total de strings carregadas: %d", totalStrings)
	fmt.Printf("\n✓ MACRODICTFILE reconstruído e pronto para uso\n")
}

// WriteMacroDictionaryToBinaryFiles writes macro dictionary data back to binary files
// This function demonstrates how to use the new RebuildMacroStrings functionality
func WriteMacroDictionaryToBinaryFiles(print bool) {
	path := filepath.Join(components.GameFilesRoot, components.ModsFolder, "edits", "macrodic", "binary")

	// Ensure the output directory exists
	if err := common.EnsurePathExists(path); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	if print {
		fmt.Println("Writing macro dictionary to binary files...")
	}

	// Process each localization
	for localization, chunks := range components.MACRODICTFILE {
		if print {
			fmt.Printf("Processing localization: %s\n", localization)
		}

		// Process each chunk
		for chunkIndex, chunk := range chunks {
			if len(chunk) == 0 {
				continue // Skip empty chunks
			}

			// Convert MacroString slice to bytes using the new rebuild function
			charset := components.LocalizationToCharset(localization)
			binaryData := components.MacroStringsToBytes(chunk, charset, true) // true = optimize (deduplicate strings)

			// Write to file
			fileName := fmt.Sprintf("macrodic_%s_chunk_%02d.dcp", localization, chunkIndex)
			filePath := filepath.Join(path, fileName)

			err := components.WriteStringToFile(filePath, string(binaryData))
			if err != nil {
				fmt.Printf("Error writing binary file %s: %v\n", filePath, err)
				continue
			}

			if print {
				fmt.Printf("  Written chunk %d: %s (%d bytes, %d strings)\n",
					chunkIndex, fileName, len(binaryData), len(chunk))
			}
		}
	}

	if print {
		fmt.Println("Macro dictionary binary files written successfully!")
	}
}

// TestMacroStringReconstruction tests the round-trip conversion: JSON → MacroString → Binary → MacroString
func TestMacroStringReconstruction(print bool) {
	fmt.Println("=== Testing Macro String Reconstruction ===")

	// Step 1: Load from JSON
	fmt.Println("1. Loading macro data from JSON...")
	macrodicPath := filepath.Join(components.GameFilesRoot, components.ModsFolder, "edits", "macrodic")
	allLocalizationsFile := filepath.Join(macrodicPath, "macro_dictionary_all_localizations.json")

	if err := EditAndSaveMacrodicFromJson(allLocalizationsFile, false); err != nil {
		fmt.Printf("Error loading from JSON: %v\n", err)
		return
	}

	// Step 2: Convert to binary and back for testing
	fmt.Println("2. Testing round-trip conversion...")
	testLocalization := "us"
	if chunks, exists := components.MACRODICTFILE[testLocalization]; exists && len(chunks) > 0 {
		chunk := chunks[6] // Test first chunk
		if len(chunk) > 0 {
			charset := components.LocalizationToCharset(testLocalization)

			// Convert to binary
			binaryData := components.MacroStringsToBytes(chunk, charset, true)
			if print {
				fmt.Printf("   Original chunk had %d strings\n", len(chunk))
				fmt.Printf("   Binary data size: %d bytes\n", len(binaryData))
			}

			// Convert back to MacroString objects
			reconstructed := components.FromStringData(binaryData[2:], charset) // Skip first 2 bytes (count)
			if print {
				fmt.Printf("   Reconstructed chunk has %d strings\n", len(reconstructed))
			}

			// Compare first few strings
			success := true
			compareCount := 5
			if len(chunk) < compareCount {
				compareCount = len(chunk)
			}
			if len(reconstructed) < compareCount {
				compareCount = len(reconstructed)
			}

			for i := 0; i < compareCount; i++ {
				if chunk[i] != nil && reconstructed[i] != nil {
					originalRegular := chunk[i].GetRegularString()
					reconstructedRegular := reconstructed[i].GetRegularString()
					originalSimplified := chunk[i].GetSimplifiedString()
					reconstructedSimplified := reconstructed[i].GetSimplifiedString()

					if originalRegular != reconstructedRegular || originalSimplified != reconstructedSimplified {
						fmt.Printf("   ❌ Mismatch at string %d:\n", i)
						fmt.Printf("      Original: '%s' / '%s'\n", originalRegular, originalSimplified)
						fmt.Printf("      Reconstructed: '%s' / '%s'\n", reconstructedRegular, reconstructedSimplified)
						success = false
					} else if print {
						fmt.Printf("   ✓ String %d matches: '%s'\n", i, originalRegular)
					}
				}
			}

			if success {
				fmt.Println("   ✓ Round-trip conversion successful!")
			} else {
				fmt.Println("   ❌ Round-trip conversion had mismatches!")
			}
		}
	}

	// Step 3: Write binary files
	fmt.Println("3. Writing binary files...")
	WriteMacroDictionaryToBinaryFiles(print)

	fmt.Println("=== Macro String Reconstruction Test Complete ===")
}
