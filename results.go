package main

import (
	"bytes"
	"io"
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
	getFile func() (io.WriteCloser, error)
	content *bytes.Buffer
}

func (w WriteFileResult) Execute() error {
	file, fileError := w.getFile()

	if fileError != nil {
		return fileError
	}

	defer func() {
		file.Close()
	}()

	if _, writeErr := file.Write(w.content.Bytes()); writeErr != nil {
		return writeErr
	}

	return nil
}
