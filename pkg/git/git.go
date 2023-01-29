package git

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

const Null = "\x00"

const (
	ObjTypeBlob   = "blob"
	ObjTypeTree   = "tree"
	ObjTypeCommit = "commit"
)

func WriteBlobObject(file string, mode fs.FileMode) (sha [20]byte, _ error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return sha, err
	}
	// see https://git-scm.com/book/en/v2/Git-Internals-Git-Objects for details.
	log.Printf("Write blob: %s", file)
	return writeObject(ObjTypeBlob, content)
}

func WriteTreeObject(dir string) (sha [20]byte, err error) {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return sha, err
	}

	var entry bytes.Buffer
	for _, f := range fs {
		var mode string
		var sha [20]byte
		if f.IsDir() {
			if f.Name() == ".git" { // Skip .git directory
				log.Println("skip .git directory")
				continue
			}
			mode = "40000"
			sha, err = WriteTreeObject(filepath.Join(dir, f.Name()))
		} else {
			mode = fmt.Sprintf("100%o", f.Mode())
			sha, err = WriteBlobObject(filepath.Join(dir, f.Name()), f.Mode())
		}
		if err != nil {
			return sha, err
		}
		// entry format = "#{mode} ${file.name}\0${hash}"
		line := fmt.Sprintf("%s %s\x00", mode, f.Name())
		entry.WriteString(line)
		entry.Write(sha[:])
		log.Printf("entry: %q", line)
	}

	log.Printf("Write tree: %s", dir)
	return writeObject(ObjTypeTree, entry.Bytes())
}

func WriteCommitObject(treeSHA, parentSHA, message string) (sha [20]byte, _ error) {
	now := time.Now().Local()
	timestamp := fmt.Sprintf("%d %s", now.Unix(), now.Format("-0700"))

	var content bytes.Buffer
	content.WriteString(fmt.Sprintf("tree %s\n", treeSHA))
	content.WriteString(fmt.Sprintf("parent %s\n", parentSHA))
	content.WriteString(fmt.Sprintf("author Ladicle <dummy@example.com> %s\n", timestamp))
	content.WriteString(fmt.Sprintf("committer Ladicle <dummy@example.com> %s\n", timestamp))
	content.WriteString("\n")
	content.WriteString(message)

	return writeObject(ObjTypeCommit, content.Bytes())
}

func writeObject(objType string, content []byte) (sha [20]byte, _ error) {
	header := fmt.Sprintf("%s %d\x00", objType, len(content))

	var data bytes.Buffer
	if _, err := data.WriteString(header); err != nil {
		return sha, err
	}
	if _, err := data.Write(content); err != nil {
		return sha, err
	}

	// calculate SHA1 from header and content
	sha = sha1.Sum(data.Bytes())
	shaStr := fmt.Sprintf("%x", sha)
	log.Printf("SHA: %x", sha)

	path := objectPath(shaStr)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		// already exists or unexpected errors
		return sha, err
	}

	if err := os.Mkdir(filepath.Dir(path), 0750); err != nil && !os.IsExist(err) {
		return sha, err
	}
	object, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return sha, err
	}

	writer := zlib.NewWriter(object)
	if _, err := writer.Write(data.Bytes()); err != nil {
		return sha, err
	}
	writer.Close()
	object.Close()
	return sha, nil
}

func DecodeObject(sha string) ([]byte, error) {
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
