package common

import (
	"os"
	"time"
)

// FileStamp é o carimbo físico de um binário: o que define o conteúdo no
// disco (caminho resolvido, tamanho, mtime). O auto-reload dos formatos usa
// a comparação de carimbos para detectar binário modificado externamente
// (tradução copiada/editada manualmente em mods/) e recarregar na próxima
// leitura, sem reiniciar o app.
type FileStamp struct {
	Path    string
	Size    int64
	ModTime time.Time
}

// StampFileFrom carimba o caminho relativo ao GameFilesRoot resolvido na
// árvore indicada. ok=false quando o arquivo não existe (ou erro de stat) —
// o chamador trata ausência como "sem carimbo".
func StampFileFrom(rel string, src FileSource) (FileStamp, bool) {
	acc, err := NewFileAccessorFrom(rel, src)
	if err != nil || !acc.Exists {
		return FileStamp{}, false
	}
	return FileStamp{
		Path:    acc.ResolvedPath,
		Size:    acc.Size,
		ModTime: acc.Info.ModTime(),
	}, true
}

// StampFile é StampFileFrom na árvore preferida (mods-first) — a mesma
// resolução da leitura normal dos formatos.
func StampFile(rel string) (FileStamp, bool) {
	return StampFileFrom(rel, SourcePreferred)
}

// StampOSPath carimba um caminho absoluto/OS direto (sem árvore). ok=false
// quando o caminho não existe.
func StampOSPath(path string) (FileStamp, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return FileStamp{}, false
	}
	return FileStamp{Path: path, Size: info.Size(), ModTime: info.ModTime()}, true
}

// StampsChanged informa se o carimbo atual diverge do registrado. Registros
// vazios divergem: ausência nova (ou removida) de arquivo é mudança.
func StampsChanged(stored, current FileStamp) bool {
	return stored != current
}
