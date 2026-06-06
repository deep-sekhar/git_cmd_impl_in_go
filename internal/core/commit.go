package core

import (
	"fmt"
	"os"
	"time"
)

func CommitTreeCmd(args []string) {
	if len(args) < 5 {
		fmt.Fprintf(os.Stderr, "usage: mygit commit-tree <tree> -m <message>\n")
		os.Exit(1)
	}
	treeHash := args[2]
	var parentSha, message string
	for i := 3; i < len(args); i++ {
		if args[i] == "-m" && i+1 < len(args) {
			message = args[i+1]
			i++
		} else if args[i] == "-p" && i+1 < len(args) {
			parentSha = args[i+1]
			i++
		}
	}

	timestamp := time.Now().Unix()
	timezone := time.Now().Format("-0700")
	authorInfo := fmt.Sprintf("Author Name <author_mail> %d %s", timestamp, timezone)
	
	commitContent := fmt.Sprintf("tree %s\n", treeHash)
	if parentSha != "" {
		commitContent += fmt.Sprintf("parent %s\n", parentSha)
	}
	commitContent += fmt.Sprintf("author %s\ncommitter %s\n\n%s\n", authorInfo, authorInfo, message)

	commitHash, _ := WriteObject("commit", []byte(commitContent))
	fmt.Println(commitHash)
}