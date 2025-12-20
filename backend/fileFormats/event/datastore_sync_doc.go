// Documentação: Datastore como fonte única da verdade para eventos
//
// NOVO FLUXO SIMPLIFICADO (Sem map global EVENTS):
//
// 1. CARREGAMENTO INICIAL (Bulk Loading):
//    ReadAllEventFiles()
//    └── discoverEventFiles() - descobre todos os IDs de eventos
//    └── loadDiscoveredEvents() - carrega cada evento individual
//        ├── ReadCompleteEventFile() - lê arquivo .ebp + localizações
//        └── SetEvent(eventID, eventFile) - registra diretamente no datastore
//
// 2. MODIFICAÇÃO VIA JSON (Individual ou Bulk):
//    ProcessEventsFromJson()
//    ├── processSingleEventData() - modifica um evento específico
//    │   ├── updateEventFromJsonData()
//    │   │   ├── GetEvent(eventID) - obtém do datastore
//    │   │   ├── modifica strings
//    │   │   └── SetEvent(eventID, eventFile) - atualiza datastore
//    │   └── ExportEventStringsToLocalizations() - salva arquivos
//    │
//    └── EditAndSaveEventFromJSON() - modifica múltiplos eventos
//        ├── GetEvent(eventID) - obtém do datastore
//        ├── modifica strings
//        ├── SetEvent(eventID, eventFile) - atualiza datastore
//        └── ExportEventStringsToLocalizations() - salva arquivos
//
// 3. EXPORTAÇÃO:
//    ExportEventStringsToLocalizations()
//    ├── GetEvent(eventID) - obtém do datastore
//    └── writeEventStringsToAllLocalizations() - escreve arquivos
//
// ACESSO GLOBAL:
// - GetEvent(eventID) - lê do datastore
// - SetEvent(eventID, eventFile) - escreve no datastore
// - SetEvents(map[string]*EventFile) - bulk write no datastore
// - ClearEvents() - limpa todos os eventos do datastore
//
// VANTAGENS:
// - Uma única fonte da verdade (datastore)
// - Sem riscos de desalinhamento de dados
// - Sem dependências circulares
// - Acesso thread-safe
//
// EXEMPLO DE USO NO CONVERTER:
//
// // Agora (sem dependência circular):
// if event := datastore.GetEvent(eventID); event != nil {
//     name := event.GetName()
// }

package event

// Esta documentação explica como os eventos são sincronizados entre
// o map EVENTS e o datastore global, resolvendo dependências circulares.
