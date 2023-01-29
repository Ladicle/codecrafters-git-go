package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/git-starter-go/pkg/git"
)

func RunDebugCmd(hash string) error {
	data, err := git.DecodeObject(hash)
	if err != nil {
		return err
	}

	// header
	buf := bytes.NewBuffer(data)
	header, _ := buf.ReadString('\000')
	header = header[:len(header)-1]
	fmt.Printf("Header ----------------------------------------\n%s\n", header)

	// tree entries
	fmt.Println("Content ---------------------------------------")
	for {
		mode, err := buf.ReadString(' ')
		if err != nil {
			break
		}
		mode = mode[:len(mode)-1]
		name, _ := buf.ReadString('\000')
		name = name[:len(name)-1]
		sha := buf.Next(20)
		fmt.Printf("%s %s %x\n", mode, name, sha)
	}
	fmt.Println("-----------------------------------------------")
	return nil
}

func RunInitCmd() error {
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

func RunCatFileCmd(sha string) error {
	data, err := git.DecodeObject(sha)
	if err != nil {
		return err
	}
	parts := bytes.SplitN(data, []byte(git.Null), 2)
	fmt.Print(string(parts[1]))
	return nil
}

func RunHashObjCmd(file string) error {
	info, err := os.Stat(file)
	if err != nil {
		return err
	}

	sha, err := git.WriteBlobObject(file, info.Mode())
	if err != nil {
		return err
	}
	fmt.Printf("%x\n", sha)
	return nil
}

func RunLsTreeCmd(sha string) error {
	data, err := git.DecodeObject(sha)
	if err != nil {
		return err
	}
	lines := bytes.Split(data, []byte(git.Null))
	for _, line := range lines[1 : len(lines)-1] {
		sep := bytes.Split(line, []byte(" "))
		fmt.Println(string(sep[len(sep)-1]))
	}
	return nil
}

func RunWriteTreeCmd() error {
	workDir, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	sha, err := git.WriteTreeObject(workDir)
	fmt.Printf("%x\n", sha)
	return err
}

func RunCommitTreeCmd(treeSHA, parentSHA, message string) error {
	sha, err := git.WriteCommitObject(treeSHA, parentSHA, message)
	if err != nil {
		return err
	}
	fmt.Printf("%x\n", sha)
	return nil
}
