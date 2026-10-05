package services

import (
	"context"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// StatusBarService emite avisos do backend para a barra de status do
// frontend (evento do wails "StatusWarning"). Fonte principal: erros de
// desalinhamento entre original e tradução — hoje só console + arquivo de
// diagnóstico.
//
// O serviço é SEM ESTADO de exibição: quem rotaciona (3s), abre o alert e
// descarta é o frontend. Reemissão do mesmo ID é dedupeado no frontend,
// que mantém os descartados da sessão.
type StatusBarService struct {
	mu  sync.Mutex
	ctx context.Context
}

// StatusBarWarning é o aviso que a barra de status exibe.
type StatusBarWarning struct {
	// ID estável do aviso (kind/version/entrada): o frontend usa para
	// dedupe e dismiss.
	ID string `json:"id"`
	// Kind/EntryID/Version localizam a entrada (events/objects/macro/…).
	Kind    string `json:"kind"`
	EntryID string `json:"entryId"`
	Version string `json:"version"`
	// Severity: "warn" | "error".
	Severity string `json:"severity"`
	Message  string `json:"message"`
	// Details: rows só em mods / só em data, contagens etc.
	Details map[string]any `json:"details,omitempty"`
}

// IStatusBarService é o contrato exposto à camada do app (registro no
// startup, igual à ponte de progresso).
type IStatusBarService interface {
	PushWarning(w StatusBarWarning)
}

func NewStatusBarService(ctx context.Context) *StatusBarService {
	return &StatusBarService{ctx: ctx}
}

// PushWarning emite o aviso para o frontend. Contexto vazio (testes/dev sem
// wails): no-op silencioso — o aviso continua no log de arquivo.
func (s *StatusBarService) PushWarning(w StatusBarWarning) {
	s.mu.Lock()
	ctx := s.ctx
	s.mu.Unlock()
	if isEmptyContext(ctx) {
		return
	}
	runtime.EventsEmit(ctx, "StatusWarning", w)
}

// statusbarBridge é a ponte instalada pela camada do app no startup.
// Avisos publicados sem ponte (testes) são no-op.
var (
	statusbarMu       sync.RWMutex
	statusbarInstance IStatusBarService
)

// SetStatusBarService instala a ponte (chamado uma vez no startup).
func SetStatusBarService(s IStatusBarService) {
	statusbarMu.Lock()
	statusbarInstance = s
	statusbarMu.Unlock()
}

// statusBarService devolve a ponte (nil quando não instalada).
func statusBarService() IStatusBarService {
	statusbarMu.RLock()
	defer statusbarMu.RUnlock()
	return statusbarInstance
}
