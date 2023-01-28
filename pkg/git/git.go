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
)

const Null = "\x00"

func WriteBlobObject(file string, mode fs.FileMode) (sha [20]byte, _ error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return sha, err
	}
	// header format = "blob #{content.bytesize}\0"
	// see https://git-scm.com/book/en/v2/Git-Internals-Git-Objects for details.
	header := fmt.Sprintf("blob %d\x00", len(content))

	log.Printf("Write blob: %s", file)
	return writeObject(header, content)
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

	// header format = "tree #{entry.bytesize}\0"
	header := fmt.Sprintf("tree %d\x00", entry.Len())
	log.Printf("header: %q", header)

	log.Printf("Write tree: %s", dir)
	return writeObject(header, entry.Bytes())
}

func writeObject(header string, content []byte) (sha [20]byte, _ error) {
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
