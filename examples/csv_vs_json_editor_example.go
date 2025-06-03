package main

import (
	"ffxresources/backend/core/reader"
	"ffxresources/backend/core/writer"
	"fmt"
)

/*
EXEMPLO DE USO: CSV vs JSON Event Editor
========================================

Este arquivo demonstra como usar os editores de eventos CSV e JSON,
mostrando as vantagens e desvantagens de cada formato.
*/

func demoWorkflowCSV() {
	fmt.Println("\n=== WORKFLOW CSV ===")

	// 1. Exportar para CSV
	fmt.Println("1. Exportando eventos para CSV...")
	writer.WriteEventFileForAllLocalizations(true)
	fmt.Println("✓ Arquivos CSV criados em: edits/events/")

	// 2. Demonstrar estrutura CSV
	fmt.Println("\n2. Estrutura do CSV:")
	fmt.Println("   Colunas: id | string index | jp | us | de | fr | it | sp | kr")
	fmt.Println("   Cada linha = uma string com suas traduções")
	fmt.Println("   Exemplo:")
	fmt.Println("   event_001,0,こんにちは,Hello,Hallo,Bonjour,Ciao,Hola,안녕하세요")

	// 3. Simular edição (usuário editaria manualmente)
	fmt.Println("\n3. [MANUAL] Editor editaria os arquivos CSV...")
	fmt.Println("   - Abrir arquivo em Excel/LibreOffice Calc")
	fmt.Println("   - Editar textos nas colunas de idioma")
	fmt.Println("   - Salvar arquivo")

	// 4. Importar mudanças
	fmt.Println("\n4. Importando mudanças do CSV...")
	err := reader.EditAndSaveEventCSVFiles(true)
	if err != nil {
		fmt.Printf("Erro: %v\n", err)
	} else {
		fmt.Println("✓ Mudanças aplicadas aos arquivos de evento")
	}
}

func demoWorkflowJSON() {
	fmt.Println("\n=== WORKFLOW JSON ===")

	// 1. Exportar para JSON
	fmt.Println("1. Exportando eventos para JSON...")
	writer.WriteEventFileForAllLocalizationsJSON(true)
	fmt.Println("✓ Arquivo JSON criado: edits/events/events_all_localizations.json")

	// 4. Importar mudanças
	fmt.Println("\n4. Importando mudanças do JSON...")
	err := reader.EditAndSaveEventJSONFiles(true)
	if err != nil {
		fmt.Printf("Erro: %v\n", err)
	} else {
		fmt.Println("✓ Mudanças aplicadas aos arquivos de evento")
	}
}
