package internals

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

/*
We need a class that maanage:

	1- Get the files and dirs of a given path
	2- read these files and dirs
*/
type Workspace struct {
	root string
}

func (w *Workspace) New(root string) error {
	w.root = root
	return nil
}

func (w *Workspace) GetPath() string {
	return w.root
}

func (w *Workspace) listFilesRec(pathname string, filesPaths *[]string) error {
	entries, err := os.ReadDir(pathname)
	if err != nil {
		return fmt.Errorf("Can't read current path - %v", err)
	}

	for _, entry := range entries {
		if entry.Name() == "." || entry.Name() == ".." || entry.Name() == ".git" || entry.Name() == ".jit" || entry.Name() == "bin" {
			continue
		}
		fmt.Println()
		fullpath := filepath.Join(pathname, entry.Name())
		if entry.IsDir() {
			w.listFilesRec(fullpath, filesPaths)
		} else {
			relPath, err := filepath.Rel(w.root, fullpath)
			if err != nil {
				return fmt.Errorf("Can't get relative path - %v", err)
			}

			*filesPaths = append(*filesPaths, relPath)
		}

	}

	return nil
}
func (w *Workspace) ListFiles() ([]string, error) {
	filesPaths := make([]string, 0)

	if err := w.listFilesRec(w.root, &filesPaths); err != nil {
		return nil, fmt.Errorf("Can't list dir files - %v", err)
	}

	return filesPaths, nil
}
func (w *Workspace) GetDirEntriesWithName(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}
func (w *Workspace) GetFileState(fileName string) (os.FileInfo, error) {
	return os.Stat(fileName)
}

func (w *Workspace) GetDirState() os.FileMode {
	return os.FileMode(040000)
}
func (w *Workspace) ReadFile(fileName string) ([]byte, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var file_content bytes.Buffer
	if _, err := io.Copy(&file_content, file); err != nil {
		return nil, err
	}

	return file_content.Bytes(), nil
}
