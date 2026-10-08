//go:build windows

package services

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// RevealFileOnOS destaca o arquivo no Explorer do Windows.
//
// Por que a API de shell e não `explorer.exe /select,"<caminho>"`: ao montar a
// linha de comando, o Go escapa as aspas embutidas como `\"` (syscall.EscapeArg,
// golang/go#21739). O explorer NÃO lê /select, via CommandLineToArgvW — ele
// rasga o trecho cru da linha de comando — e, ao não reconhecer o switch, cai
// na pasta padrão da conta (numa máquina com OneDrive Known-Folder-Move, a
// pasta do OneDrive). Era exatamente o sintoma reportado: "Abrir até o arquivo"
// abria o OneDrive sem destacar arquivo algum.
//
// SHOpenFolderAndSelectItems não passa por linha de comando: recebe o PIDL do
// arquivo e destaca o item dentro da pasta que o contém. Sem subprocesso, sem
// janela de console, sem o exit code 1 do explorer (que era o motivo do código
// antigo não fazer Wait).
//
// Os PIDLs vêm do allocator de tarefa da shell e são liberados com ILFree.
var (
	shell32                        = windows.NewLazySystemDLL("shell32.dll")
	procILCreateFromPathW          = shell32.NewProc("ILCreateFromPathW")
	procSHOpenFolderAndSelectItems = shell32.NewProc("SHOpenFolderAndSelectItems")
	procILFree                     = shell32.NewProc("ILFree")
)

// revealFileOnOS abre o Explorer com abs destacado. O caminho precisa existir
// e ser absoluto (revealInExplorer garante isso antes de chamar).
func revealFileOnOS(abs string) error {
	pathPtr, err := windows.UTF16PtrFromString(abs)
	if err != nil {
		return fmt.Errorf("caminho inválido: %w", err)
	}

	pidl, _, _ := procILCreateFromPathW.Call(uintptr(unsafe.Pointer(pathPtr)))
	if pidl == 0 {
		// A shell não achou o item: o Explorer abriria a pasta padrão da
		// conta, que é o comportamento silencioso que se quer eliminar.
		return fmt.Errorf("arquivo não encontrado pela shell: %s", abs)
	}
	defer procILFree.Call(pidl)

	// pidlFolder = PIDL do próprio item, cidl = 0, apidl = nil: a shell abre
	// a pasta PAI do item e o deixa selecionado. É o uso canônico de
	// SHOpenFolderAndSelectItems para "revelar no explorador".
	hr, _, _ := procSHOpenFolderAndSelectItems.Call(pidl, 0, 0, 0)
	if hr != 0 {
		return fmt.Errorf("não foi possível revelar o arquivo (hr=0x%08X): %s", uint32(hr), abs)
	}
	return nil
}
