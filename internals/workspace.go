package internals

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Workspace struct {
	root   string
	ignore map[string]bool
}

func (w *Workspace) New(root string) error {
	w.root = root
	w.ignore = map[string]bool{
		".":          true,
		"..":         true,
		".git":       true,
		".jit":       true,
		"bin":        true,
		".env":       true,
		".gitignore": true,
	}
	return nil
}

func (w *Workspace) GetPath() string {
	return w.root
}

func (w *Workspace) ListFiles(pathname string) ([]string, error) {
	// if the pathname is empty, we will use the root path
	if len(pathname) == 0 {
		pathname = w.root
	}

	info, err := os.Stat(pathname)
	if err != nil {
		return nil, fmt.Errorf("Can't get current path's stat - %v", err)
	}

	if info.IsDir() {
		entries, err := os.ReadDir(pathname)
		if err != nil {
			return nil, fmt.Errorf("Can't get dir's entries - %v", err)
		}

		var files []string
		for _, entry := range entries {
			entryName := entry.Name()
			if w.ignore[entryName] {
				continue
			}

			childName := filepath.Join(pathname, entryName)
			childFiles, err := w.ListFiles(childName)

			if err != nil {
				return nil, fmt.Errorf("Can't get dir's entries - %v", err)
			}

			files = append(files, childFiles...)
		}

		return files, nil
	}
	relPath, err := filepath.Rel(w.root, pathname)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}
	return []string{relPath}, nil
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
