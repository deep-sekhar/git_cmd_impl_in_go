package network

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"git_cmd_impl_in_go/internal/core"
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

	// create and change to the target directory for the new repository.
	if err := os.MkdirAll(dir, 0755); err != nil { os.Exit(1) }
	if err := os.Chdir(dir); err != nil { os.Exit(1) }
	
	// initialise the git repo
	core.InitRepo(".")
	// get the target commit hash from the remote repository 
	headHash := discoverRefs(url)
	if headHash != "" {
		// fetch and store the tree, commit, and blob objects from the remote repository in the .git/objects directory of the newly created repository.
		fetchAndUnpack(url, headHash)
		// decompress the commit object to read its content.Then, we will recursively read the tree objects and blob objects to reconstruct the file structure of the repository in the target directory.
		checkoutCommit(headHash, ".")
	}
}

func checkoutCommit(commitHash string, targetDir string) {
	commitData, _ := core.ReadObject(commitHash)
	lines := strings.Split(string(commitData), "\n")
	treeHash := ""
	for _, line := range lines {
		// the line that starts with "tree " will contain the hash of the tree object that represents the file structure of the repository at the commit. We need to extract this tree hash to be able to checkout the files and directories in the repository.
		// format of this line : tree <hash>
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
	// as defined in the writeTree function, the format of content in tree object is:
	// <mode> <name>\0<20-byte hash> for each entry in the tree.
	for len(content) > 0 {
		spaceIndex := bytes.IndexByte(content, ' ')
		mode := string(content[:spaceIndex])
		content = content[spaceIndex+1:]
		
		nullIndex := bytes.IndexByte(content, 0)
		// file name
		name := string(content[:nullIndex])
		content = content[nullIndex+1:]
		
		// convert 20 byte hash to hex string
		// we will use this hash to read the corresponding blob or tree object from the .git/objects directory and then write the file content to the target directory
		entryHash := hex.EncodeToString(content[:20])
		content = content[20:]
		// full path of the file or directory to be checked out in the target directory
		targetPath := filepath.Join(targetDir, name)
		
		// if mode is "40000" or "040000", it represents a directory, so we need to create the directory and recursively checkout its contents.
		if mode == "40000" || mode == "040000" {
			os.MkdirAll(targetPath, 0755)
			checkoutTree(entryHash, targetPath)
		} else {
			// else it represents a file, so we need to read the corresponding blob object from the .git/objects directory using the entry hash, and then write the file content to the target path in the target directory. We also need to set the file permissions based on the mode specified in the tree object (e.g., 100644 for regular files and 100755 for executable files).
			blobData, _ := core.ReadObject(entryHash)
			os.WriteFile(targetPath, blobData, 0644)
			if mode == "100755" { os.Chmod(targetPath, 0755) }
		}
	}
}