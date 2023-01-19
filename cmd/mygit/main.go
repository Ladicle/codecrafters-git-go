package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const separator = "\x00"

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
		// ./your_git.sh init
		return runInitCmd()
	case "cat-file":
		// your_git.sh cat-file -p <hash>
		return runCatFileCmd(os.Args[3])
	case "hash-object":
		// ./your_git.sh hash-object -w <file>
		return runHashObjCmd(os.Args[3])
	case "ls-tree":
		// ./your_git.sh ls-tree --name-only <tree_sha>
		return runLsTreeCmd(os.Args[3])
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

func runCatFileCmd(sha string) error {
	dirName := sha[:2]
	fileName := sha[2:]

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
	parts := bytes.SplitN(data, []byte(separator), 2)
	fmt.Print(string(parts[1]))
	return nil
}

func runHashObjCmd(file string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	hasher := sha1.New()
	header := []byte(fmt.Sprintf("blob %d%s", len(content), separator))
	if _, err := hasher.Write(header); err != nil {
		return err
	}
	if _, err := hasher.Write(content); err != nil {
		return err
	}
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	path := filepath.Join(".git", "objects", hash[:2], hash[2:])
	if err := os.Mkdir(filepath.Dir(path), 0750); err != nil && !os.IsExist(err) {
		return err
	}
	object, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	writer := zlib.NewWriter(object)
	if _, err := writer.Write(header); err != nil {
		return err
	}
	if _, err := writer.Write(content); err != nil {
		return err
	}
	writer.Close()
	object.Close()

	fmt.Println(hash)
	return nil
}

func runLsTreeCmd(sha string) error {
	dirName := sha[:2]
	fileName := sha[2:]

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

	lines := bytes.Split(data, []byte(separator))
	for _, line := range lines[1 : len(lines)-1] {
		sep := bytes.Split(line, []byte(" "))
		fmt.Println(string(sep[len(sep)-1]))
	}
	return nil
}
