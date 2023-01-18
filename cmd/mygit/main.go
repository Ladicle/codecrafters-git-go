package main

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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

	cmd := os.Args[1]
	switch cmd {
	case "init":
		return runInitCmd()
	case "cat-file":
		return runCatFileCmd()
	}
	return fmt.Errorf("unknown command: %s", cmd)
}

func runInitCmd() error {
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
		}
	}

	headFileContents := []byte("ref: refs/heads/master\n")
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
	}

	fmt.Println("Initialized git directory")
	return nil
}

func runCatFileCmd() error {
	// your_git.sh cat-file -p <hash>
	hash := os.Args[3]

	dirName := hash[:2]
	fileName := hash[2:]

	f, err := os.Open(filepath.Join(".git", "objects", dirName, fileName))
	if err != nil {
		return err
	}

	r, err := zlib.NewReader(f)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	parts := bytes.SplitN(data, []byte("\x00"), 2)
	fmt.Print(string(parts[1]))
	return nil
}
