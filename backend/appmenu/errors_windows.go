//go:build windows

// Classificação de erros de sistema no Windows: rede/UNC (caminho de
// servidor \\servidor\..., compartilhamento removido), arquivo em uso por
// outro programa e acesso negado nativo — os casos que o acesso externo
// (menu) mais pega, já que o usuário aponta caminhos livres.
package appmenu

import (
	"errors"
	"syscall"

	"golang.org/x/sys/windows"
)

// classifySystemError mapeia os códigos nativos do Windows para mensagens
// legíveis. Vazio = não reconheceu (o chamador segue para as sentinelas
// io/fs e o erro cru).
func classifySystemError(err error) string {
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return ""
	}
	switch errno {
	case windows.ERROR_ACCESS_DENIED:
		return "Acesso negado pelo sistema (arquivo/pasta protegidos — verifique as permissões ou execute como administrador)."
	case windows.ERROR_SHARING_VIOLATION:
		return "O arquivo está em uso por outro programa — feche-o e tente de novo."
	case windows.ERROR_LOCK_VIOLATION:
		return "O arquivo está travado por outro processo — feche-o e tente de novo."
	case windows.ERROR_PATH_NOT_FOUND:
		return "Caminho não encontrado (a pasta, a unidade ou o compartilhamento de rede não existe)."
	case windows.ERROR_BAD_NETPATH:
		return "Falha de acesso em rede/UNC: caminho \\\\servidor\\... inalcançável (rede indisponível, compartilhamento removido ou firewall)."
	case windows.ERROR_NETWORK_UNREACHABLE, windows.ERROR_NETWORK_ACCESS_DENIED:
		return "Falha de rede: o destino está inalcançável (verifique a conexão e as permissões do compartilhamento)."
	case windows.ERROR_HANDLE_DISK_FULL, windows.ERROR_DISK_FULL:
		return "Disco cheio: não há espaço para gravar o arquivo."
	}
	return ""
}
