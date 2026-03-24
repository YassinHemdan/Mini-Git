package internals

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

/*
We need a class that maanage:

	1- Get the files and dirs of a given path
	2- read these files and dirs
*/
type Workspace struct {
	path    string
	entries []os.DirEntry
}

func (w *Workspace) New(path string) error {
	w.path = path
	entries, err := os.ReadDir(path)

	if err != nil {
		return fmt.Errorf("Cant't read the current directry - %v", err)
	}

	for _, entry := range entries { // only get the files
		if entry.Name() == "." || entry.Name() == ".." || entry.Name() == ".git" || entry.Name() == ".jit" {
			continue
		}

		// if entry.Name() == "temp" {
		// 	w.entries = append(w.entries, entry)
		// }

		w.entries = append(w.entries, entry)
	}

	return nil
}

func (w *Workspace) GetPath() string {
	return w.path
}
func (w *Workspace) GetDirEntries() []os.DirEntry {
	return w.entries
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
