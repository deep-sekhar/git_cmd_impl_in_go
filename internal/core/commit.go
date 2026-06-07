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
	// commit object format is:
	// tree <tree hash>
	// parent <parent commit hash> (optional)
	// author <author name> <author email> <timestamp> <timezone>
	// committer <committer name> <committer email> <timestamp> <timezone>
	//
	// <commit message>
	
	timestamp := time.Now().Unix()
	// we need time in the format of "Author Name <author_mail> 1609459200 -0700" where 1609459200 is the Unix timestamp and -0700 is the timezone offset. We can use time.Now().Format("-0700") to get the timezone offset in the required format.
	timezone := time.Now().Format("-0700")
	authorInfo := fmt.Sprintf("Author Name <author_mail> %d %s", timestamp, timezone)
	// The author and committer information is the same in this implementation
	committerInfo := fmt.Sprintf("Committer Name <committer_mail> %d %s", timestamp, timezone)
	commitContent := fmt.Sprintf("tree %s\n", treeHash)
	if parentSha != "" {
		commitContent += fmt.Sprintf("parent %s\n", parentSha)
	}
	commitContent += fmt.Sprintf("author %s\ncommitter %s\n\n%s\n", authorInfo, committerInfo, message)

	commitHash, _ := WriteObject("commit", []byte(commitContent))
	fmt.Println(commitHash)
}