package lib

import (
	"fmt"
	"os"
	"path/filepath"
)

type GameLocation struct {
	path string
}

const defaultDirName = "data"

func NewGameLocation() *GameLocation {
	rootDirectory = GetExecDir()

	return &GameLocation{
		path: PathJoin(rootDirectory, defaultDirName),
	}
}

func (g *GameLocation) GetPath() string {
	return g.path
}

func (g *GameLocation) SetPath(path string) {
	g.path = filepath.Clean(path)
}

func (g GameLocation) IsSpira() error {
	return containsNewUSPCPath(g.path)
}

func (g GameLocation) IsSpiraPath(path string) bool {
	return hasSpira(path)
}

func containsNewUSPCPath(userPath string) error {
	cleanedPath := filepath.Clean(userPath)

	requiredSequence := filepath.Join("ffx_ps2", "ffx2", "master", "new_uspc")
	requiredPath := filepath.Join(cleanedPath, requiredSequence)

	if _, err := os.Stat(requiredPath); os.IsNotExist(err) {
		return fmt.Errorf("is not a valid spira us path: %s", userPath)
	}
	return nil
}

func hasSpira(path string) bool {
	return IsValidPath.MatchString(path)
}
