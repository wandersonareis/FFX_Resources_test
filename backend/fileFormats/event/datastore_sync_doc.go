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
// 2. MODIFICAÇÃO VIA DTO (Individual ou Bulk, fora deste pacote):
//    backend/builders.ApplyEventsDTO()
//    ├── event.GetEvent(eventID) - obtém do datastore
//    ├── modifica strings
//    ├── event.SetEvent(eventID, eventFile) - atualiza datastore
//    └── event.ExportEventStringsToLocalizations() - salva arquivos
//    O pacote event desconhece JSON/DTO: só expõe binário e acesso ao store.
//    O JSON vive em backend/formatters/json e o DTO em backend/dto,
//    montado por backend/builders a partir dos dados brutos.
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
//     name := event.GetID()
// }

package event

// Esta documentação explica como os eventos são sincronizados entre
// o map EVENTS e o datastore global, resolvendo dependências circulares.
