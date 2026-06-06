package main

import (
	"fmt"
	"github.com/codecrafters-io/git-starter-go/internal/core"
	"github.com/codecrafters-io/git-starter-go/internal/network"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		core.InitRepo(".")
		fmt.Println("Initialized git directory")

	case "cat-file":
		core.CatFileCmd(os.Args)

	case "hash-object":
		core.HashObjectCmd(os.Args)

	case "ls-tree":
		core.LsTreeCmd(os.Args)

	case "write-tree":
		hash := core.WriteTree(".")
		fmt.Println(hash)

	case "commit-tree":
		core.CommitTreeCmd(os.Args)

	case "clone":
		network.CloneCmd(os.Args)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}