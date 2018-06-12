package internal

import (
	"os"
	"fmt"
	"bytes"
	"io"
	"log"
	"path/filepath"
	"strings"
)

func ReadFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf("error opening file: %s", err)
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)

	if err != nil {
		log.Fatal(fmt.Sprintf("error reading from buffer: %s", err))
	}

	data := buf.String()

	return data, nil

}

func walkFile(path string, f os.FileInfo, err error) error {
	if ! f.IsDir() {
		dir, file := filepath.Split(path)
		policyPath := strings.Join(strings.Split(dir, "/")[1:], "/")
		fmt.Println(file)
		fmt.Println(policyPath)
	}

	return nil
}

func PutPoliciesFromDir(path string) error {
	err := filepath.Walk(path, walkFile)
	return err
}
