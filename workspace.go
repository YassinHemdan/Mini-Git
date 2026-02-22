package main

import (
	"bytes"
	"io"
	"os"
)

type Workspace struct {
	path_name string
}

func (w *Workspace) New(path_name string) error {
	w.path_name = path_name

	return nil
}

func (w *Workspace) GetFiles() ([]os.DirEntry, error) {
	entries, err := os.ReadDir(w.path_name)
	if err != nil {
		return entries, err
	}

	var files []os.DirEntry

	for _, entry := range entries {
		if entry.Name() == "." || entry.Name() == ".." || entry.IsDir() {
			continue
		}

		files = append(files, entry)
	}

	return files, nil
}

func (w *Workspace) GetFileContent(file *os.File) ([]byte, error) {
	var file_content bytes.Buffer

	if _, err := io.Copy(&file_content, file); err != nil {
		return file_content.Bytes(), err
	}

	return file_content.Bytes(), nil
}
