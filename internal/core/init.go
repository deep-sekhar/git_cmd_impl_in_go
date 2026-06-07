package core

import (
	"os"
	"path/filepath"
)

// InitRepo creates the .git directory structure
func InitRepo(basePath string) error {
	// create .git, .git/objects, and .git/refs directories
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		// xxx bits -> 1st bit for write permission, 2nd bit for read permission, 3rd bit for execute permission
		// 7 = 111 = read + write + execute -> for owner
		// 5 = 101 = read + execute -> for group, and others
		if err := os.MkdirAll(filepath.Join(basePath, dir), 0755); err != nil {
			return err
		}
	}
	headPath := filepath.Join(basePath, ".git", "HEAD")
	return os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0644)
}