package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/codecrafters-io/git-starter-go/pkg/cmd"
)

// Usage: your_git.sh <command> <arg1> <arg2> ...
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		return errors.New("invalid argument: command is required")
	}

	name := os.Args[1]
	switch name {
	case "init":
		// ./your_git.sh init
		return cmd.RunInitCmd()
	case "cat-file":
		// your_git.sh cat-file -p <hash>
		return cmd.RunCatFileCmd(os.Args[3])
	case "hash-object":
		// ./your_git.sh hash-object -w <file>
		return cmd.RunHashObjCmd(os.Args[3])
	case "ls-tree":
		// ./your_git.sh ls-tree --name-only <tree_sha>
		return cmd.RunLsTreeCmd(os.Args[3])
	case "write-tree":
		// ./your_git.sh write-tree
		return cmd.RunWriteTreeCmd()
	case "commit-tree":
		// ./your_git.sh commit-tree <tree_sha> -p <commit_sha> -m <message>
		return cmd.RunCommitTreeCmd(os.Args[2], os.Args[4], os.Args[6])
	case "debug":
		// your_git.sh debug <hash>
		return cmd.RunDebugCmd(os.Args[2])
	}
	return fmt.Errorf("unknown command: %s", name)
}
