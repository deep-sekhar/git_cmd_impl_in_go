package main

import (
	"fmt"
	"git_cmd_impl_in_go/internal/core"
	"git_cmd_impl_in_go/internal/network"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		// sample : <build file> init
		core.InitRepo(".")
		fmt.Println("Initialized git directory")

	case "cat-file":
		// sample : <build file> cat-file -p <object> where <object> is the hash of the object we want to read
		core.CatFileCmd(os.Args)

	case "hash-object":
		// sample : <build file> hash-object -w <file> where <file> is the path to the file we want to hash and write as a blob object in the .git/objects directory.
		core.HashObjectCmd(os.Args)

	case "ls-tree":
		// sample : <build file> ls-tree <tree-ish> where <tree-ish> is the hash of the tree object we want to list.
		core.LsTreeCmd(os.Args)

	case "write-tree":
		// sample : <build file> write-tree which will create a tree object from the current directory and print its hash. 
		// It will recursively read the files and directories in the current directory, create blob objects for the files, and tree objects for the directories, and then create a tree object for the current directory that references all the blob and tree objects it contains. Finally, it will print the hash of the created root tree object.
		hash := core.WriteTree(".")
		fmt.Println(hash)

	case "commit-tree":
		// sample : <build file> commit-tree <tree> -p <parent> -m <message> where <tree> is the hash of the tree object we want to commit, <parent> is the hash of the parent commit (optional), and <message> is the commit message. This command will create a new commit object that references the specified tree and parent commit, and then print the hash of the created commit object.
		core.CommitTreeCmd(os.Args)

	case "clone":
		// sample: <build file> clone <repository-url> where <repository-url> is the URL of the remote repository we want to clone. This command will create a new directory with the name of the repository, initialize a new git repository in it, and then fetch all the objects and refs from the remote repository and store them in the .git directory of the newly created repository.
		network.CloneCmd(os.Args)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}