// Package progress é o canal de progresso das operações longas (export de
// textos, extração de arquivos/imagens do .vbf): o serviço que dirige a
// operação cria um Batch no emitter INJETADO (nunca um global), declara os
// Steps — uma ação cada — e repassa cada step como Reporter para quem faz
// o trabalho (builders, loops de extração).
//
// Sem globals: o emitter vive no MetadataService (injetado pelo app.go) e
// emitter nil = batches no-op (testes/CLI/appcore seguem sem progresso).
// Todas as implementações do pacote são nil-receiver-safe, então caller
// sem interesse em progresso passa nil e os métodos viram no-op.
//
// O emitter publica SNAPSHOTS COMPLETOS nos eventos do Wails:
//
//	ProgressStart  → abre o modal (título + steps declarados)
//	ProgressUpdate → tick (throttle 10/s; transições publicam sempre)
//	ProgressEnd    → estado final (done, resumo de arquivos/destino)
//
// Snapshot completo (em vez de delta "1" acumulado no frontend): um evento
// perdido nunca corrompe a exibição — o frontend renderiza o último.
package progress

// Reporter é o canal de UMA ação do batch. O código dirigido (export de um
// kind, extração de uma seleção) não sabe que está num batch: recebe um
// Reporter e chama Begin/Step/Issue/End.
type Reporter interface {
	// Begin inicia a ação. O label de EXIBIÇÃO é o declarado no batch
	// (Step do Batch) — o label daqui é informativo; total 0 =
	// indeterminado até o primeiro SetTotal.
	Begin(label string, total int)
	// SetTotal define o total quando só é conhecido depois de começar
	// (a extração descobre o tamanho do lote ao montá-lo).
	SetTotal(total int)
	// Step avança um item concluído (identificador para exibição).
	Step(item string)
	// Issue registra um item pulado por erro não-catastrófico: a
	// operação continua com o que deu certo.
	Issue(item, message string)
	// End encerra a ação (concluída; o estado de erro é do Fail).
	End()
}

// Nop devolve o Reporter no-op — o "sem progresso" explícito.
func Nop() Reporter { return nopReporter{} }

// OrNop normaliza o reporter recebido: nil vira o no-op, evitando
// nil-check espalhado pelos intermediários (GetCollection, builders).
// Reporters concretos do pacote são nil-receiver-safe, então um step de
// batch sem emitter também passa direto.
func OrNop(r Reporter) Reporter {
	if r == nil {
		return nopReporter{}
	}
	return r
}

type nopReporter struct{}

func (nopReporter) Begin(string, int)    {}
func (nopReporter) SetTotal(int)         {}
func (nopReporter) Step(string)          {}
func (nopReporter) Issue(string, string) {}
func (nopReporter) End()                 {}

// Estados de um step (o frontend pinta cada um).
const (
	StepPending = "pending"
	StepRunning = "running"
	StepDone    = "done"
	StepError   = "error"
)

// StepSnapshot é o estado serializável de UMA ação do batch.
type StepSnapshot struct {
	Label      string `json:"label"`
	Total      int    `json:"total"`
	Processed  int    `json:"processed"`
	IssueCount int    `json:"issueCount"`
	State      string `json:"state"`
	Error      string `json:"error,omitempty"`
}

// BatchSnapshot é o payload completo dos três eventos — o frontend apenas
// renderiza o último recebido.
type BatchSnapshot struct {
	Title       string         `json:"title"`
	Steps       []StepSnapshot `json:"steps"`
	Current     int            `json:"current"`
	Done        bool           `json:"done"`
	Failed      bool           `json:"failed,omitempty"`
	IssueCount  int            `json:"issueCount"`
	Written     int            `json:"written,omitempty"`
	Destination string         `json:"destination,omitempty"`
}
