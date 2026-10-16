//go:build !windows

// Fallback de classificação fora do Windows: só as sentinelas io/fs (as
// mensagens amigáveis comuns vivem em dialogs.go).
package appmenu

// classifySystemError não tem códigos nativos fora do Windows — devolve
// vazio para o chamador seguir com as sentinelas io/fs.
func classifySystemError(err error) string {
	return ""
}
