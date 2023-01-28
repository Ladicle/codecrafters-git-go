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

const SepNull = "\x00"

const (
	TypeTree = "40000"
	TypeBlob = "100644"
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
	case "debug":
		// your_git.sh debug <hash>
		return runDebugCmd(os.Args[2])
	}
	return fmt.Errorf("unknown command: %s", cmd)
}

func runDebugCmd(hash string) error {
	data, err := decodeObject(hash)
	if err != nil {
		return err
	}

	fmt.Println(len(data))

	// header
	buf := bytes.NewBuffer(data)
	header, _ := buf.ReadString('\000')
	header = header[:len(header)-1]
	fmt.Printf("%s\n", header)

	// tree entries
	for {
		mode, err := buf.ReadString(' ')
		if err != nil {
			break
		}
		mode = mode[:len(mode)-1]
		name, _ := buf.ReadString('\000')
		name = name[:len(name)-1]
		sha := buf.Next(20)
		fmt.Printf("%v %s %x\n", tpeToStr(mode), name, sha)
	}
	return nil
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
	data, err := decodeObject(sha)
	if err != nil {
		return err
	}
	parts := bytes.SplitN(data, []byte(SepNull), 2)
	fmt.Print(string(parts[1]))
	return nil
}

func runHashObjCmd(file string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	header := fmt.Sprintf("%s %d%s", TypeBlob, len(content), SepNull)

	var buf bytes.Buffer
	if _, err := buf.WriteString(header); err != nil {
		return err
	}
	if _, err := buf.Write(content); err != nil {
		return err
	}

	hash := fmt.Sprintf("%x", sha1.New().Sum(buf.Bytes()))

	path := objectPath(hash)
	if err := os.Mkdir(filepath.Dir(path), 0750); err != nil && !os.IsExist(err) {
		return err
	}
	object, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil && !os.IsExist(err) {
		return err
	}
	writer := zlib.NewWriter(object)
	if _, err := writer.Write(buf.Bytes()); err != nil {
		return err
	}
	writer.Close()
	object.Close()

	fmt.Println(hash)
	return nil
}

func runLsTreeCmd(sha string) error {
	data, err := decodeObject(sha)
	if err != nil {
		return err
	}
	lines := bytes.Split(data, []byte(SepNull))
	for _, line := range lines[1 : len(lines)-1] {
		sep := bytes.Split(line, []byte(" "))
		fmt.Println(string(sep[len(sep)-1]))
	}
	return nil
}

func decodeObject(sha string) ([]byte, error) {
	obj, err := os.Open(objectPath(sha))
	if err != nil {
		return nil, err
	}
	r, err := zlib.NewReader(obj)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}

func objectPath(sha string) string {
	return filepath.Join(".git", "objects", sha[:2], sha[2:])
}

func tpeToStr(tpe string) string {
	switch tpe {
	case TypeBlob:
		return "blob"
	case TypeTree:
		return "tree"
	}
	return "unknown"
}
