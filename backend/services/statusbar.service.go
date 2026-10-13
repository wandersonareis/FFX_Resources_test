package services

import (
	"sync"
)

// StatusBarService emite avisos do backend para a barra de status do
// frontend (evento do wails "StatusWarning"). Fonte principal: erros de
// desalinhamento entre original e tradução — hoje só console + arquivo de
// diagnóstico.
//
// O serviço é SEM ESTADO de exibição: quem rotaciona (3s), abre o alert e
// descarta é o frontend. Reemissão do mesmo ID é dedupeado no frontend,
// que mantém os descartados da sessão.
//
// O ctx do v2 (só usado por runtime.EventsEmit) saiu: no v3 o evento sai
// pelo bus de application.Get().
type StatusBarService struct{}

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

func NewStatusBarService() *StatusBarService {
	return &StatusBarService{}
}

// PushWarning emite o aviso para o frontend. Sem app wails (testes/dev sem
// janela): no-op silencioso — o aviso continua no log de arquivo.
func (s *StatusBarService) PushWarning(w StatusBarWarning) {
	if !HasEventBus() {
		return
	}
	EmitEvent("StatusWarning", w)
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
