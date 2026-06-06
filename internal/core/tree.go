package core

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type TreeEntry struct {
	mode string
	name string
	hash []byte
}

func WriteTree(dirPath string) string {
	candidates, _ := os.ReadDir(dirPath)
	var treeEntries []TreeEntry

	for _, candidate := range candidates {
		if candidate.Name() == ".git" { continue }

		var mode, hash string
		candidatePath := filepath.Join(dirPath, candidate.Name())

		if candidate.IsDir() {
			mode = "40000"
			hash = WriteTree(candidatePath)
		} else {
			info, _ := candidate.Info()
			if info.Mode()&0111 != 0 {
				mode = "100755"
			} else {
				mode = "100644"
			}
			fileContent, _ := os.ReadFile(candidatePath)
			hash, _ = WriteObject("blob", fileContent)
		}

		hashBytes, _ := hex.DecodeString(hash)
		treeEntries = append(treeEntries, TreeEntry{
			mode: mode,
			name: candidate.Name(),
			hash: hashBytes,
		})
	}

	sort.Slice(treeEntries, func(i, j int) bool {
		return treeEntries[i].name < treeEntries[j].name
	})

	var treeData []byte
	for _, entry := range treeEntries {
		entryData := fmt.Sprintf("%s %s\x00", entry.mode, entry.name)
		treeData = append(treeData, []byte(entryData)...)
		treeData = append(treeData, entry.hash...)
	}

	treeHash, _ := WriteObject("tree", treeData)
	return treeHash
}

func LsTreeCmd(args []string) {
	if len(args) < 4 || args[2] != "--name-only" {
		fmt.Fprintf(os.Stderr, "usage: mygit ls-tree --name-only <tree>\n")
		os.Exit(1)
	}
	content, _ := ReadObject(args[3])
	for len(content) > 0 {
		spaceIndex := bytes.IndexByte(content, ' ')
		content = content[spaceIndex+1:]
		nullIndex := bytes.IndexByte(content, 0)
		name := content[:nullIndex]
		fmt.Println(string(name))
		content = content[nullIndex+21:]
	}
}