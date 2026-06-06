package network

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/codecrafters-io/git-starter-go/internal/core"
	"os"
	"path/filepath"
	"strings"
)

func CloneCmd(args []string) {
	if len(args) < 4 {
		fmt.Fprintf(os.Stderr, "usage: mygit clone <url> <dir>\n")
		os.Exit(1)
	}
	url := args[2]
	dir := args[3]

	if err := os.MkdirAll(dir, 0755); err != nil { os.Exit(1) }
	if err := os.Chdir(dir); err != nil { os.Exit(1) }
	
	core.InitRepo(".")
	headHash := discoverRefs(url)
	if headHash != "" {
		fetchAndUnpack(url, headHash)
		checkoutCommit(headHash, ".")
	}
}

func checkoutCommit(commitHash string, targetDir string) {
	commitData, _ := core.ReadObject(commitHash)
	lines := strings.Split(string(commitData), "\n")
	treeHash := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "tree ") {
			treeHash = strings.TrimPrefix(line, "tree ")
			break
		}
	}
	if treeHash != "" {
		checkoutTree(treeHash, targetDir)
	}
}

func checkoutTree(treeHash string, targetDir string) {
	treeData, _ := core.ReadObject(treeHash)
	content := treeData
	for len(content) > 0 {
		spaceIndex := bytes.IndexByte(content, ' ')
		mode := string(content[:spaceIndex])
		content = content[spaceIndex+1:]
		
		nullIndex := bytes.IndexByte(content, 0)
		name := string(content[:nullIndex])
		content = content[nullIndex+1:]
		
		entryHash := hex.EncodeToString(content[:20])
		content = content[20:]
		targetPath := filepath.Join(targetDir, name)
		
		if mode == "40000" || mode == "040000" {
			os.MkdirAll(targetPath, 0755)
			checkoutTree(entryHash, targetPath)
		} else {
			blobData, _ := core.ReadObject(entryHash)
			os.WriteFile(targetPath, blobData, 0644)
			if mode == "100755" { os.Chmod(targetPath, 0755) }
		}
	}
}