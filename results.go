package main

import (
	"bytes"
	"os"
	"path/filepath"
)

type Result interface {
	Execute() error
}

type ErrorResult struct {
	error
}

func (e ErrorResult) Execute() error {
	return e.error
}

type WriteFileResult struct {
	fileName string
	content  *bytes.Buffer
}

func (w WriteFileResult) Execute() error {
	cleanFileName := filepath.Clean(w.fileName)

	file, fileErr := os.Create(cleanFileName)

	if fileErr != nil {
		return fileErr
	}

	defer func() {
		file.Close()
	}()

	if _, writeErr := file.Write(w.content.Bytes()); writeErr != nil {
		return writeErr
	}

	return nil
}
