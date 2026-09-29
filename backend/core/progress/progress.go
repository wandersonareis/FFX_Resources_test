// Package progress é o canal neutro de progresso das operações longas
// (carregamento de árvore, exportação): o domínio (fileFormats, builders)
// publica contadores SEM conhecer wails — quem emite os eventos é a ponte
// registrada pela camada do app. Sem reporter registrado, tudo é no-op:
// logging/progresso nunca podem derrubar a operação.
package progress

import "sync"

// Reporter é a ponte de progresso implementada pela camada do app.
type Reporter interface {
	// Begin inicia um ciclo com rótulo; total 0 = indeterminado.
	Begin(label string, total int)
	// Step avança um item (identificador do arquivo/entrada).
	Step(item string)
	// Issue registra um problema não-catastrófico: o item é pulado e a
	// operação continua.
	Issue(item, message string)
	// End encerra o ciclo (o frontend fecha o indicador).
	End()
}

var (
	mu       sync.RWMutex
	reporter Reporter
)

// Set instala a ponte (chamado uma vez pela camada do app no startup).
func Set(r Reporter) {
	mu.Lock()
	reporter = r
	mu.Unlock()
}

func get() Reporter {
	mu.RLock()
	r := reporter
	mu.RUnlock()
	return r
}

// Begin inicia um ciclo de progresso (no-op sem ponte registrada).
func Begin(label string, total int) {
	if r := get(); r != nil {
		r.Begin(label, total)
	}
}

// Step avança o ciclo corrente (no-op fora de um ciclo).
func Step(item string) {
	if r := get(); r != nil {
		r.Step(item)
	}
}

// Issue registra um item pulado por erro não-catastrófico.
func Issue(item, message string) {
	if r := get(); r != nil {
		r.Issue(item, message)
	}
}

// End encerra o ciclo corrente (no-op se nenhum).
func End() {
	if r := get(); r != nil {
		r.End()
	}
}
