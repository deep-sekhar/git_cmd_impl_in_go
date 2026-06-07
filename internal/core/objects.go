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
	// Git stores objects in .git/objects/xx/yyyy... where xx are the first 2 chars of the hash, and yyyy... are the remaining chars
	dir, fileName := hash[:2], hash[2:]
	objectPath := filepath.Join(".git", "objects", dir, fileName)
	
	file, err := os.Open(objectPath)
	if err != nil { return "", nil, err }
	defer file.Close()
	
	r, err := zlib.NewReader(file)
	if err != nil { return "", nil, err }
	defer r.Close()
	
	// Openfile works on files in the filesystem, but we need to read the decompressed content of the object which is not directly accessible as a file. So we read the decompressed content into memory and then parse it.
	data, err := io.ReadAll(r)
	if err != nil { return "", nil, err }
	
	// split the data with null byte as the separator and take only the first 2 parts (object type + size, and content)
	// format : <object type> <size>\0<content>
	parts := bytes.SplitN(data, []byte{0}, 2)
	if len(parts) < 2 { return "", nil, fmt.Errorf("invalid object format") }
	
	headerParts := strings.Split(string(parts[0]), " ")
	// return object type, content, and error
	return headerParts[0], parts[1], nil
}

func ReadObject(hash string) ([]byte, error) {
	_, content, err := ReadObjectFull(hash)
	return content, err
}

func WriteObject(objectType string, content []byte) (string, error) {
	// \x00 is the null byte
	// Git object format is <object type> <size>\0<content> where content is the binary content of the object
	header := fmt.Sprintf("%s %d\x00", objectType, len(content))
	dataToHash := append([]byte(header), content...)

	h := sha1.New()
	// Write the data to the hash function. The Write method takes a byte slice and updates the hash state with it.
	h.Write(dataToHash)
	// find the hash of the content, Sum takes the current hash state and returns the hash as a byte slice. We convert it to a hex string using fmt.Sprintf with %x verb which formats the byte slice as a hexadecimal string.
	hash := fmt.Sprintf("%x", h.Sum(nil))

	// first 2 chars of the hash are the directory name, and the remaining chars are the file name
	dir, fileName := hash[:2], hash[2:]
	objectDir := filepath.Join(".git", "objects", dir)
	objectPath := filepath.Join(objectDir, fileName)

	// create the directory if it doesn't exist
	if err := os.MkdirAll(objectDir, 0755); err != nil { return "", err }

	// If the object already exists, we don't need to write it again.
	// os.Stat returns an error if the file does not exist, and we check for that using os.IsNotExist. 
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		file, err := os.Create(objectPath)
		if err != nil { return "", err }
		defer file.Close()

		// write the compressed data to the file using zlib. We create a new zlib writer that wraps the file, and then write the data to it. Finally, we close the writer to flush the data to the file.
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