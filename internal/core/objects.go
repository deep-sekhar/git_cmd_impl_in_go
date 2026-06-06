package core

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ReadObjectFull(hash string) (string, []byte, error) {
	dir, fileName := hash[:2], hash[2:]
	objectPath := filepath.Join(".git", "objects", dir, fileName)
	
	file, err := os.Open(objectPath)
	if err != nil { return "", nil, err }
	defer file.Close()
	
	r, err := zlib.NewReader(file)
	if err != nil { return "", nil, err }
	defer r.Close()
	
	data, err := io.ReadAll(r)
	if err != nil { return "", nil, err }
	
	parts := bytes.SplitN(data, []byte{0}, 2)
	if len(parts) < 2 { return "", nil, fmt.Errorf("invalid object format") }
	
	headerParts := strings.Split(string(parts[0]), " ")
	return headerParts[0], parts[1], nil
}

func ReadObject(hash string) ([]byte, error) {
	_, content, err := ReadObjectFull(hash)
	return content, err
}

func WriteObject(objectType string, content []byte) (string, error) {
	header := fmt.Sprintf("%s %d\x00", objectType, len(content))
	dataToHash := append([]byte(header), content...)

	h := sha1.New()
	h.Write(dataToHash)
	hash := fmt.Sprintf("%x", h.Sum(nil))

	dir, fileName := hash[:2], hash[2:]
	objectDir := filepath.Join(".git", "objects", dir)
	objectPath := filepath.Join(objectDir, fileName)

	if err := os.MkdirAll(objectDir, 0755); err != nil { return "", err }

	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		file, err := os.Create(objectPath)
		if err != nil { return "", err }
		defer file.Close()

		w := zlib.NewWriter(file)
		w.Write(dataToHash)
		w.Close()
	}
	return hash, nil
}

func CatFileCmd(args []string) {
	if len(args) < 4 || args[2] != "-p" {
		fmt.Fprintf(os.Stderr, "usage: mygit cat-file -p <object>\n")
		os.Exit(1)
	}
	content, err := ReadObject(args[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s", string(content))
}

func HashObjectCmd(args []string) {
	if len(args) < 4 || args[2] != "-w" {
		fmt.Fprintf(os.Stderr, "usage: mygit hash-object -w <file>\n")
		os.Exit(1)
	}
	content, err := os.ReadFile(args[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	hash, _ := WriteObject("blob", content)
	fmt.Println(hash)
}