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

func NewWorkspace(root string) (*Workspace, error) {
	ignore := make(map[string]bool)
	ignore["."] = true
	ignore[".."] = true
	ignore[".git"] = true
	ignore[".jit"] = true
	ignore[".env"] = true
	ignore[".gitignore"] = true
	ignore["bin"] = true

	return &Workspace{
		root:   root,
		ignore: ignore,
	}, nil
}

func (w *Workspace) GetPath() string {
	return w.root
}

func (w *Workspace) ListFiles(pathname string) ([]string, error) {
	if len(pathname) == 0 {
		pathname = w.root
	}

	info, err := os.Stat(pathname)
	if err != nil {
		return nil, fmt.Errorf("Can't get current path's stat - %w", err)
	}

	if info.IsDir() {
		entries, err := os.ReadDir(pathname)
		if err != nil {
			return nil, fmt.Errorf("Can't get dir's entries - %w", err)
		}

		files := make([]string, 0)
		for _, entry := range entries {
			entryName := entry.Name()
			if w.ignore[entryName] {
				continue
			}

			childName := filepath.Join(pathname, entryName)
			childFiles, err := w.ListFiles(childName)

			if err != nil {
				return nil, fmt.Errorf("Can't get dir's entries - %w", err)
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

/*
the difference between ListDir and ListFiles is that ListDir only gets

	1st level entries of a agiven directory

ListFiles returnes all the files in depth
*/
func (w *Workspace) ListDir(pathname string) (map[string]os.FileInfo, error) {
	if len(pathname) == 0 {
		pathname = w.root
	} else {
		pathname = filepath.Join(w.root, pathname)
	}

	info, err := os.Stat(pathname)
	if err != nil {
		return nil, fmt.Errorf("Can't get current path's stat - %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("Error: Expected a dir name - %w", err)
	}
	entries, err := os.ReadDir(pathname)
	if err != nil {
		return nil, fmt.Errorf("Can't get dir's entries - %w", err)
	}

	result := make(map[string]os.FileInfo)

	for _, entry := range entries {
		entryName := entry.Name()
		if w.ignore[entryName] {
			continue
		}
		fullEntryPath := filepath.Join(pathname, entryName)

		entryInfo, err := os.Stat(fullEntryPath)
		if err != nil {
			return nil, fmt.Errorf("Can't get current path's stat - %w", err)
		}

		relPath, err := filepath.Rel(w.root, fullEntryPath)

		if err != nil {
			return nil, fmt.Errorf("failed to get relative path: %w", err)
		}

		result[relPath] = entryInfo
	}
	return result, nil
}

func (w *Workspace) GetFileState(fileName string) (os.FileInfo, error) {
	return os.Stat(w.fullpath(fileName))
}

func (w *Workspace) GetDirState() os.FileMode {
	return os.FileMode(040000)
}
func (w *Workspace) ReadFile(fileName string) ([]byte, error) {
	file, err := os.Open(w.fullpath(fileName))
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

func (w *Workspace) fullpath(pathname string) string {
	return filepath.Join(w.root, pathname)
}
