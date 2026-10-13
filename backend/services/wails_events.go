package services

import "github.com/wailsapp/wails/v3/pkg/application"

// EmitEvent publica um evento no bus do Wails v3.
//
// No v2 era runtime.EventsEmit(ctx, ...), que exigia um context.Context
// válido e virava no-op com ctx vazio (guard isEmptyContext). No v3 os
// eventos não usam ctx: a referência vem de application.Get(), que devolve
// nil antes de application.New() e depois do shutdown — nesses casos a
// publicação é descartada, mantendo o mesmo contrato de "sem app, sem emit".
func EmitEvent(name string, data ...any) {
	if app := application.Get(); app != nil {
		app.Event.Emit(name, data...)
	}
}

// HasEventBus responde se o app v3 já existe — é a condição que substitui o
// antigo isEmptyContext para decidir se vale a pena montar o payload.
func HasEventBus() bool {
	return application.Get() != nil
}
