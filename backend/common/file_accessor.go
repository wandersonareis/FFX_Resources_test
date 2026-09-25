package common

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	ModsFolder        = "mods/"
	ModsTranslatedDir = "translated"
	DirData           = "data"
	DirExtracted      = "extracted"
	DirTranslated     = "translated"
	DirReimported     = "reimported"
)

// DefaultTranslatedDir deriva o diretório de tradução de um gamefiles:
// <gamefiles>/mods/translated. Usado quando não há valor no config.json.
func DefaultTranslatedDir(gameFilesDir string) string {
	return filepath.Join(gameFilesDir, ModsFolder, ModsTranslatedDir)
}

var (
	ResourcesRoot = GetExecDir()
	GameFilesRoot = ResourcesRoot
	// DisableMods desliga a preferência por arquivos em mods/. O estado
	// é controlado em runtime via SetModsEnabled (config EnableMods).
	DisableMods = false
)

// SetModsEnabled liga/desliga a preferência mods-first da resolução de
// caminhos (FileAccessor + loaders de binário). O valor vem do config
// (EnableMods) e é aplicado no bootstrap da aplicação.
func SetModsEnabled(enabled bool) {
	DisableMods = !enabled
}

func SetGameFilesRoot(path string) {
	if path == "" {
		fmt.Println("Invalid path provided for GameFilesRoot")
		return
	}
	GameFilesRoot = filepath.Clean(path)
	fmt.Printf("GameFilesRoot set to: %s\n", GameFilesRoot)
}

func getRealFile(path string) string {
	return filepath.Join(GameFilesRoot, path)
}

func getModdedFile(path string) string {
	return filepath.Join(GameFilesRoot, ModsFolder, path)
}

// WriteBytesToFile writes a slice of integers as bytes to a file
// Creates necessary directories before writing
func WriteBytesToFile(path string, bytes []byte) error {
	CreateDirectories(path)

	err := os.WriteFile(path, bytes, 0644)
	if err != nil {
		fmt.Println("Failed to write file")
		return err
	}

	return nil
}

// CreateDirectories creates the necessary directories for a given file path
// If the path is a directory, it ensures the directory exists
// If the path is a file, it creates all necessary parent directories
func CreateDirectories(path string) {
	info, err := os.Stat(path)
	var dirPath string

	if err == nil && info.IsDir() {
		dirPath = path
	} else {
		dirPath = filepath.Dir(path)
	}

	if dirPath == "" || dirPath == "." {
		return
	}

	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		fmt.Println("Failed to create directories: ", err)
	}
}

type FileAccessor struct {
	RootPath     string
	ResolvedPath string
	Info         os.FileInfo
	Size         int64
	Exists       bool
}

// NewFileAccessor creates a new FileAccessor instance for the given path
// This function handles both existing and non-existing paths gracefully
//
// Parameters:
//   - path: File path to access (can be relative or absolute)
//
// Returns:
//   - *FileAccessor: FileAccessor instance with file information
//   - error: Error if path resolution fails (not if file doesn't exist)
func NewFileAccessor(path string) (FileAccessor, error) {
	resolvedPath, err := resolvePath(path)
	if err != nil {
		return FileAccessor{}, err
	}

	fileInfo, exists := getFileInfo(resolvedPath)

	return FileAccessor{
		RootPath:     path,
		ResolvedPath: resolvedPath,
		Info:         fileInfo,
		Size:         getFileSize(fileInfo),
		Exists:       exists,
	}, nil
}

func (f *FileAccessor) ReadBytes() ([]byte, error) {
	data, err := os.ReadFile(f.ResolvedPath)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// NewRealFileAccessor cria um FileAccessor apontando SEMPRE para o arquivo
// original (<gamefiles>/<path>), ignorando mods — para enumeração de
// diretórios (descoberta de eventos) e casos onde o caminho deve ser o
// original mesmo com mods habilitado. A leitura por arquivo continua podendo
// preferir mods via NewFileAccessor.
func NewRealFileAccessor(path string) (FileAccessor, error) {
	resolvedPath, err := getRealPath(path)
	if err != nil {
		return FileAccessor{}, err
	}

	fileInfo, exists := getFileInfo(resolvedPath)

	return FileAccessor{
		RootPath:     path,
		ResolvedPath: resolvedPath,
		Info:         fileInfo,
		Size:         getFileSize(fileInfo),
		Exists:       exists,
	}, nil
}

// getRealPath resolve o caminho absoluto na árvore original, sem mods.
func getRealPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	return getRealFile(path), nil
}

func resolvePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	if DisableMods {
		return getRealFile(path), nil
	}

	return resolveModdedPath(path), nil
}

func resolveModdedPath(path string) string {
	moddedPath := getModdedFile(path)
	if fileExists(moddedPath) {
		return moddedPath
	}
	return getRealFile(path)
}

func getFileInfo(path string) (os.FileInfo, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	return info, true
}

func getFileSize(info os.FileInfo) int64 {
	if info == nil {
		return 0
	}
	return info.Size()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (f *FileAccessor) String() string {
	if f.Exists {
		return fmt.Sprintf("FileAccessor{RootPath: %s, ResolvedPath: %s, Size: %d bytes}", f.RootPath, f.ResolvedPath, f.Size)
	}
	return fmt.Sprintf("FileAccessor{RootPath: %s, ResolvedPath: %s, Exists: false}", f.RootPath, f.ResolvedPath)
}
