package progress

import (
	"sync"
	"time"

	"ffxresources/backend/loggingService"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	// emitThrottle é a cadência máxima de ProgressUpdate (10/s) —
	// transições (Begin/End/Fail do step, fim do batch) publicam
	// incondicionalmente, então o estado final sempre chega.
	emitThrottle = 100 * time.Millisecond

	// maxIssueToasts é o cap de toasts individuais por batch: além dele
	// os erros seguem no arquivo de diagnóstico (lista completa), mas a
	// UI não é inundada.
	maxIssueToasts = 15
)

// Emitter é a ponte wails dos batches: construído UMA vez no startup
// (app.go) e injetado no MetadataService. nil = sem progresso (testes,
// CLI, appcore). warn recebe as mensagens de Issue (toasts com cap); sem
// app wails a publicação é descartada — mesmo contrato do v2/v3.
type Emitter struct {
	warn func(string)
	// emit é a seam de publicação (tests capturam sem wails).
	emit func(event string, snap BatchSnapshot)
}

// NewEmitter cria a ponte. warn pode ser nil (Issues sem toast).
func NewEmitter(warn func(string)) *Emitter {
	return &Emitter{warn: warn, emit: emitWails}
}

func emitWails(event string, snap BatchSnapshot) {
	if app := application.Get(); app != nil {
		app.Event.Emit(event, snap)
	}
}

// NewBatch cria a operação agrupada (nil-safe: emitter nil → batch nil).
func (e *Emitter) NewBatch(title string) *Batch {
	if e == nil {
		return nil
	}
	return &Batch{em: e, title: title}
}

// stepState é o estado interno de uma ação (protegido pelo mutex do
// batch — os steps de um batch nunca correm em paralelo).
type stepState struct {
	label     string
	total     int
	processed int
	issues    int
	state     string
	errMsg    string
}

// Batch é UMA operação agrupada: título (o modal) + um step por ação, na
// ordem de execução. Quem dirige declara os steps e os conduz pelos
// Reporters; o End publica o snapshot final. nil-safe em todos os
// métodos (sem emitter tudo é no-op).
type Batch struct {
	em      *Emitter
	mu      sync.Mutex
	title   string
	steps   []*stepState
	ended   bool
	failed  bool
	toasts  int // Issues já tostadas (cap)
	lastHit time.Time
	started bool
	// Resumo do fim (SetSummary): quantos arquivos saíram e pra onde.
	written     int
	destination string
}

// Steps declara as ações do batch NA ORDEM de execução e devolve os
// canais. A declaração publica o snapshot com a lista — na PRIMEIRA vai
// como ProgressStart (o modal abre com tudo) e nas seguintes como update.
// Após End() é no-op (devolve canais soltos que não publicam nada).
func (b *Batch) Steps(labels ...string) []*Step {
	if b == nil {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]*Step, 0, len(labels))
	for _, label := range labels {
		if b.ended {
			out = append(out, &Step{})
			continue
		}
		b.steps = append(b.steps, &stepState{label: label, state: StepPending})
		out = append(out, &Step{b: b, idx: len(b.steps) - 1})
	}
	if !b.ended && len(labels) > 0 {
		event := "ProgressUpdate"
		if !b.started {
			event = "ProgressStart"
			b.started = true
		}
		b.em.emit(event, b.snapshotLocked())
	}
	return out
}

// Step declara UMA ação (atalho de Steps para os batches de passo único).
func (b *Batch) Step(label string) *Step {
	steps := b.Steps(label)
	if len(steps) == 0 {
		return nil
	}
	return steps[0]
}

// SetSummary define o resumo do estado final: quantos arquivos a operação
// escreveu e o destino (vai no ProgressEnd — o modal mostra no resultado).
func (b *Batch) SetSummary(written int, destination string) {
	if b == nil {
		return
	}
	b.mu.Lock()
	b.written, b.destination = written, destination
	b.mu.Unlock()
}

// Fail marca o batch como falho (uma ação abortou a operação): o snapshot
// final carrega failed=true para o modal pintar o resultado de erro.
func (b *Batch) Fail() {
	if b == nil {
		return
	}
	b.mu.Lock()
	b.failed = true
	b.mu.Unlock()
}

// End publica o snapshot final (done) e encerra o ciclo. Idempotente.
func (b *Batch) End() {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.ended {
		return
	}
	b.ended = true
	b.em.emit("ProgressEnd", b.snapshotLocked())
}

// currentLocked devolve o índice da ação corrente: o último step que saiu
// de pending (rodando ou já fechado).
func (b *Batch) currentLocked() int {
	cur := 0
	for i, st := range b.steps {
		if st.state != StepPending {
			cur = i
		}
	}
	return cur
}

func (b *Batch) snapshotLocked() BatchSnapshot {
	steps := make([]StepSnapshot, len(b.steps))
	issues := 0
	for i, st := range b.steps {
		steps[i] = StepSnapshot{
			Label:      st.label,
			Total:      st.total,
			Processed:  st.processed,
			IssueCount: st.issues,
			State:      st.state,
			Error:      st.errMsg,
		}
		issues += st.issues
	}
	return BatchSnapshot{
		Title:       b.title,
		Steps:       steps,
		Current:     b.currentLocked(),
		Done:        b.ended,
		Failed:      b.failed,
		IssueCount:  issues,
		Written:     b.written,
		Destination: b.destination,
	}
}

// publishLocked publica o snapshot: force=true publica incondicionalmente
// (transições); senão respeita o throttle de 10/s.
func (b *Batch) publishLocked(force bool) {
	now := time.Now()
	if !force && now.Sub(b.lastHit) < emitThrottle {
		return
	}
	b.lastHit = now
	b.em.emit("ProgressUpdate", b.snapshotLocked())
}

// Step é o canal (Reporter) de UMA ação do batch. nil-receiver-safe: sem
// emitter o batch é nil e os steps também — todos os métodos são no-op.
type Step struct {
	b   *Batch
	idx int
}

// Begin inicia a ação (pending → running). O label daqui é informativo: o
// de exibição é o declarado no batch. Publica incondicionalmente.
func (s *Step) Begin(label string, total int) {
	if s == nil || s.b == nil {
		return
	}
	s.b.mu.Lock()
	defer s.b.mu.Unlock()
	st := s.b.steps[s.idx]
	st.total = total
	if st.state == StepPending {
		st.state = StepRunning
	}
	s.b.publishLocked(true)
}

// SetTotal define o total da ação (indeterminada até aqui) e publica.
func (s *Step) SetTotal(total int) {
	if s == nil || s.b == nil {
		return
	}
	s.b.mu.Lock()
	defer s.b.mu.Unlock()
	s.b.steps[s.idx].total = total
	s.b.publishLocked(false)
}

// Step avança um item concluído (throttle: transições publicam sempre, os
// ticks de meio de lote respeitam 10/s — o End leva o estado cheio).
func (s *Step) Step(item string) {
	if s == nil || s.b == nil {
		return
	}
	s.b.mu.Lock()
	defer s.b.mu.Unlock()
	s.b.steps[s.idx].processed++
	s.b.publishLocked(false)
}

// Issue registra um item pulado: conta na ação (issues + processed) e
// tosta com cap — além do cap o erro segue só no arquivo de diagnóstico.
func (s *Step) Issue(item, message string) {
	if s == nil || s.b == nil {
		return
	}
	s.b.mu.Lock()
	st := s.b.steps[s.idx]
	st.issues++
	if item != "" {
		st.processed++
	}
	s.b.toasts++
	toast := s.b.toasts <= maxIssueToasts
	s.b.publishLocked(false)
	warn := s.b.em.warn
	s.b.mu.Unlock()

	loggingService.DiagWarn("carga/"+item, message, map[string]any{"item": item})
	if toast && warn != nil {
		warn(message)
	}
}

// End encerra a ação (running → done). Idempotente; publica sempre.
func (s *Step) End() {
	if s == nil || s.b == nil {
		return
	}
	s.b.mu.Lock()
	defer s.b.mu.Unlock()
	st := s.b.steps[s.idx]
	if st.state == StepRunning {
		st.state = StepDone
	}
	s.b.publishLocked(true)
}

// Fail marca a ação como falha (mensagem no resultado). Fora do Reporter:
// é do autor do batch (quem vê o erro da operação), não do código
// dirigido. Vale mesmo depois de um End — o autor decide o veredito.
func (s *Step) Fail(message string) {
	if s == nil || s.b == nil {
		return
	}
	s.b.mu.Lock()
	defer s.b.mu.Unlock()
	st := s.b.steps[s.idx]
	st.state = StepError
	st.errMsg = message
	s.b.failed = true
	s.b.publishLocked(true)
}
