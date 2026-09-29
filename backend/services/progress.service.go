package services

import (
	"context"
	"sync"
	"time"

	"ffxresources/backend/core/progress"
	"ffxresources/backend/loggingService"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ProgressService emite o progresso de operações longas por eventos do
// wails (Progress/ShowProgress) — o pipeline do legado, reativado com
// rótulo e contagens para o frontend montar a barra com título.
//
// O fluxo é serial por ciclo (Begin → Steps → End): o mutex protege o
// estado contra usos concorrentes de leitura. Emissão é throttled
// (diThrottle) para não re-renderizar a UI em rajada; o End emite o estado
// final incondicionalmente e fecha o diálogo.
type ProgressService struct {
	mu          sync.Mutex
	ctx         context.Context
	notifier    INotificationService
	label       string
	total       int
	processed   int
	percentage  int
	currentItem string
	issueCount  int
	enabled     bool
	lastEmit    time.Time
}

const (
	// progressEmitThrottle é a cadência máxima de eventos Progress (10/s).
	progressEmitThrottle = 100 * time.Millisecond

	// maxIssueToasts é o cap de toasts individuais por ciclo — além dele os
	// erros continuam indo para o arquivo de diagnóstico (lista completa),
	// mas a UI não é inundada.
	maxIssueToasts = 15
)

// IProgressService é o contrato da ponte de progresso exposto à camada do
// app (o contrato de reporter vem de core/progress).
type IProgressService interface {
	progress.Reporter
}

func NewProgressService(ctx context.Context, notifier INotificationService) *ProgressService {
	return &ProgressService{ctx: ctx, notifier: notifier}
}

// Begin inicia um ciclo de progresso.
func (p *ProgressService) Begin(label string, total int) {
	p.mu.Lock()
	p.label, p.total = label, total
	p.processed, p.percentage, p.currentItem, p.issueCount = 0, 0, "", 0
	p.enabled = true
	p.lastEmit = time.Time{}
	p.mu.Unlock()
	p.publish(true)
}

// Step avança um item; item vazio apenas incrementa o contador.
func (p *ProgressService) Step(item string) {
	p.mu.Lock()
	if !p.enabled {
		p.mu.Unlock()
		return
	}
	p.processed++
	if item != "" {
		p.currentItem = item
	}
	if p.total > 0 {
		pct := p.processed * 100 / p.total
		if pct > 100 {
			pct = 100
		}
		p.percentage = pct
	}
	now := time.Now()
	if now.Sub(p.lastEmit) < progressEmitThrottle {
		p.mu.Unlock()
		return
	}
	p.lastEmit = now
	p.mu.Unlock()
	p.publish(false)
}

// Issue registra um item pulado por erro não-catastrófico: sempre no
// arquivo de diagnóstico (lista completa); toast individual com cap
// anti-flood por ciclo.
func (p *ProgressService) Issue(item, message string) {
	p.mu.Lock()
	p.issueCount++
	count := p.issueCount
	p.mu.Unlock()

	loggingService.DiagWarn("carga/"+item, message, map[string]any{"item": item})
	if count <= maxIssueToasts && p.notifier != nil {
		p.notifier.NotifyWarn(message)
	}
}

// End encerra o ciclo: emite o estado final e fecha o indicador do
// frontend. Ciclos fora de Begin são ignorados (End idempotente).
func (p *ProgressService) End() {
	p.mu.Lock()
	enabled := p.enabled
	p.enabled = false
	p.mu.Unlock()
	if !enabled || isEmptyContext(p.ctx) {
		return
	}
	runtime.EventsEmit(p.ctx, "Progress", p)
	runtime.EventsEmit(p.ctx, "ShowProgress", false)
}

func (p *ProgressService) publish(show bool) {
	if isEmptyContext(p.ctx) {
		return
	}
	if show {
		runtime.EventsEmit(p.ctx, "ShowProgress", true)
	}
	runtime.EventsEmit(p.ctx, "Progress", p)
}

// stateSnapshot expõe o estado corrente para testes (sem tocar no wails).
func (p *ProgressService) stateSnapshot() (label string, total, processed, percentage int, item string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.label, p.total, p.processed, p.percentage, p.currentItem
}

// ensure implementado via métodos públicos: ProgressService É o Reporter
// da ponte (core/progress) — Begin/Step/Issue/End batem por assinatura.
var _ progress.Reporter = (*ProgressService)(nil)
