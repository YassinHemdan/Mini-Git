package internals

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"syscall"
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

func (w *Workspace) applyMigration(migration *Migration) error {
	if err := w.applyChangeList(migration, MIGRATION_DELETE); err != nil {
		return err
	}

	dirsToDelete := slices.Collect(maps.Keys(migration.rmdirs))
	slices.SortFunc(dirsToDelete, func(a, b string) int {
		if a < b {
			return 1
		}
		if a > b {
			return -1
		}
		return 0
	})

	for _, dir := range dirsToDelete {
		if err := w.removeDirectory(dir); err != nil {
			return err
		}
	}

	dirsToCreate := slices.Collect(maps.Keys(migration.mkdirs))
	slices.Sort(dirsToCreate)

	for _, dir := range dirsToCreate {
		if err := w.makeDirectory(dir); err != nil {
			return err
		}
	}

	if err := w.applyChangeList(migration, MIGRATION_UPDATE); err != nil {
		return err
	}
	if err := w.applyChangeList(migration, MIGRATION_CREATE); err != nil {
		return err
	}
	return nil
}

/*
Notice that we are removing the file with os.RemoveAll(fullpath)

	Remember that we are changing our workspace to match the tree of the target commit
	so if by any chance we have a directory (in our current tree) with the same name of the file we
	want to (create, delete, update), we wanna make sure that we delete this directory first to avoid
*/
func (w *Workspace) applyChangeList(migration *Migration, action string) error {
	listPairs := migration.changes[action]
	for _, pair := range listPairs {
		pathname := pair.First
		entry := pair.Second

		fullpath := filepath.Join(w.root, pathname)
		os.RemoveAll(fullpath)

		if action == MIGRATION_DELETE {
			continue
		}

		flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
		data, err := migration.blobData(entry.GetOid())

		if err != nil {
			return err
		}

		file, err := os.OpenFile(fullpath, flags, 0666)
		if err != nil {
			return err
		}

		_, err = file.Write(data)
		if err != nil {
			return err
		}

		file.Close()

		mode, err := strconv.ParseUint(entry.GetMode(), 8, 32)
		if err != nil {
			return err
		}
		err = os.Chmod(fullpath, os.FileMode(mode))

	}
	return nil
}
func (w *Workspace) makeDirectory(dir string) error {
	path := filepath.Join(w.root, dir)
	info, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		// we expect it not to be exist, but if there is an error and it exists, we return it
		return err
	}

	if info != nil && info.Mode().IsRegular() {
		// if the dir path exists and it is a regular file, delete it
		if err := os.Remove(path); err != nil {
			return err
		}
	}

	if info == nil || !info.IsDir() {
		// if the path does not exist or it exists but not a dir (a file that we just deleted)
		// we create it
		if err := os.Mkdir(path, 0755); err != nil {
			return err
		}
	}
	return nil
}
func (w *Workspace) removeDirectory(dir string) error {
	// if the directory does not exist, not empty or not a dir, we just ignore
	path := filepath.Join(w.root, dir)
	err := os.Remove(path)

	if err != nil {
		if os.IsNotExist(err) { // doesn't exist
			return nil
		}
		if pathErr, ok := err.(*os.PathError); ok {
			switch pathErr.Err {
			case syscall.ENOTDIR, syscall.ENOTEMPTY: // not a dir or not empty
				return nil
			}
		}
		return err
	}
	return nil
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
func (w *Workspace) IsExecutable(pathname string) (bool, error) {
	fileInfo, err := w.GetFileState(pathname)
	if err != nil {
		return false, err
	}

	if fileInfo.Mode()&0111 != 0 {
		return true, nil
	}
	return false, nil
}
