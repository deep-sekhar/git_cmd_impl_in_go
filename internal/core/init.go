package core

import (
	"os"
	"path/filepath"
)

// InitRepo creates the .git directory structure
func InitRepo(basePath string) error {
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		if err := os.MkdirAll(filepath.Join(basePath, dir), 0755); err != nil {
			return err
		}
	}
	headPath := filepath.Join(basePath, ".git", "HEAD")
	return os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0644)
}