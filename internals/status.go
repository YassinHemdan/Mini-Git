package internals

import (
	internals "JIT/internals/database"
	index "JIT/internals/index"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

const (
	UNTRACKED = "untracked"
	DELETED   = "deleted"
	MODIFIED  = "modified"
	ADDED     = "new file"
	GREEN     = "green"
	RED       = "red"
)

var shortStatusMap map[string]string
var longStatusMap map[string]string

type Status struct {
	repo              *Repository
	untracked         []string
	changed           []string
	workspace_changes map[string]string
	index_changes     map[string]string
	states            map[string]os.FileInfo
	headTree          map[string]internals.Entry
	statusSize        int
	inspector         *Inspector
}

func newStatus(repo *Repository) (*Status, error) {
	st := &Status{
		repo:              repo,
		states:            make(map[string]os.FileInfo),
		index_changes:     make(map[string]string),
		workspace_changes: make(map[string]string),
		headTree:          make(map[string]internals.Entry),
		inspector:         newInspector(repo),
	}

	if err := st.scan(""); err != nil {
		return nil, fmt.Errorf("Couldn't scan workspace: %v\n", err)
	}

	if err := st.loadHeadTree(); err != nil {
		return nil, fmt.Errorf("Couldn't Load HeadTree: %v\n", err)
	}

	if err := st.checkIndexEntries(); err != nil {
		return nil, fmt.Errorf("Couldn't check index entries: %v\n", err)
	}

	st.collectDeletedHeadFiles()

	return st, nil
}

// to detect the untracked files or directories
func (s *Status) scan(dirName string) error {
	dirEntriesMap, err := s.repo.Workspace().ListDir(dirName)
	if err != nil {
		return err
	}
	for entryName, entryInfo := range dirEntriesMap {
		if !s.repo.Index().IsTracked(entryName) { // not tracked, just add it
			// we want to make sure that it is not an empty dir
			found, err := s.isTrackableFile(entryName, entryInfo)
			if err != nil {
				return err
			}

			if found { // we can add it .. either a file or non-empty dir
				if entryInfo.IsDir() {
					entryName += "/"
				}
				s.untracked = append(s.untracked, entryName)
			}
			continue
		}

		//if a directory marked as tracked, there might be inner files that still untracked we need to find them (if exists)
		if entryInfo.IsDir() {
			if err := s.scan(entryName); err != nil {
				return err
			}
		} else {
			s.states[entryName] = entryInfo
		}
	}
	return nil
}

// to check if the a directory is empty or not
// if there are only nested directories with no files, we consider the higher directory is empty
// if there is at least one file -> not empty
func (s *Status) isTrackableFile(entryName string, entryInfo os.FileInfo) (bool, error) {
	/*
		a BFS algorithm to check if the current entry has any nested file inside it
		- if the entry itself is a file, return true
		- if the the entry is a directory, we will run bfs to expand more
	*/

	if !entryInfo.IsDir() {
		return true, nil
	}
	queue := []string{entryName}
	for len(queue) != 0 {
		dirName := queue[0]
		queue = queue[1:]

		childEntries, err := s.repo.Workspace().ListDir(dirName)
		if err != nil {
			return false, err
		}

		for childName, childInfo := range childEntries {
			if !childInfo.IsDir() {
				return true, nil
			}
			queue = append(queue, childName)
		}

	}

	return false, nil

	// false + nil => the directory is empty
	// true + nil => directory is not empty
	// ... + err => we have a problem :)
}

func (s *Status) checkIndexEntries() error {
	indexEntries := s.repo.Index().GetEntries()
	for _, entry := range indexEntries {
		if err := s.checkIndexAgainstWorkspace(entry); err != nil {
			return err
		}
		s.checkIndexAgainstHeadTree(entry)
	}
	return nil
}

func (s *Status) checkIndexAgainstWorkspace(entry *index.IndexEntry) error {
	info, _ := s.states[entry.GetPathname()]
	status, err := s.inspector.checkIndexAgainstWorkspace(entry, info)
	if err != nil {
		return err
	}

	if status == "" {
		stat := info.Sys().(*syscall.Stat_t)
		s.repo.Index().UpdateEntryStat(entry, stat)
	} else {
		s.recordChange(entry.GetPathname(), s.workspace_changes, status)
	}

	return nil

}
func (s *Status) checkIndexAgainstHeadTree(indexEntry *index.IndexEntry) {
	val, _ := s.headTree[indexEntry.GetPathname()]
	status := s.inspector.checkIndexAgainstHeadTree(indexEntry, val)
	if status != "" {
		s.recordChange(indexEntry.GetPathname(), s.index_changes, status)
	}
}

func (s *Status) collectDeletedHeadFiles() {
	for pathname := range s.headTree {
		if !s.repo.Index().IsTrackedFile(pathname) {
			s.recordChange(pathname, s.index_changes, DELETED)
		}
	}
}
func (s *Status) recordChange(pathname string, changesMap map[string]string, changeType string) {
	s.changed = append(s.changed, pathname)
	changesMap[pathname] = changeType
	s.statusSize = max(s.statusSize, len(changeType))
}
func (s *Status) loadHeadTree() error {
	commitOid, err := s.repo.Refs().ReadHead()
	if err != nil {
		return fmt.Errorf("Could not read Refs - %v\n", err)
	}

	// what if there aren't any commits yet ?
	if len(commitOid) == 0 {
		return nil
	}
	obj, err := s.repo.Database().Load(commitOid)
	if err != nil || obj.Type() != "commit" {
		return fmt.Errorf("Could not load commit object from DB - %v\n", err)
	}

	var loadTree func(oid []byte) (map[string]internals.Entry, error)
	loadTree = func(treeOid []byte) (map[string]internals.Entry, error) {
		obj, err := s.repo.Database().Load(treeOid)
		if err != nil || obj.Type() != "tree" {
			return nil, fmt.Errorf("Could not load tree object from DB - %v\n", err)
		}

		treeObj := obj.(*internals.Tree)
		return treeObj.GetEntries(), nil
	}

	var showTree func(entries map[string]internals.Entry, pathname string) error
	showTree = func(entries map[string]internals.Entry, pathname string) error {
		for _, entry := range entries {
			fullpathname := filepath.Join(pathname, entry.GetName())
			if entry.Type() == "tree" {
				subTreeEntires, err := loadTree(entry.GetOid())
				if err != nil {
					return err
				}
				if err := showTree(subTreeEntires, fullpathname); err != nil {
					return err
				}
			} else if entry.Type() == "blob" {
				s.headTree[fullpathname] = entry
			}
		}

		return nil
	}

	commitObj := obj.(*internals.Commit)
	loadedTreeEntries, err := loadTree(commitObj.GetTreeOid())
	if err != nil {
		return err
	}
	if err := showTree(loadedTreeEntries, ""); err != nil {
		return err
	}

	return nil
}
func (s *Status) GetUntracked() []string {
	return s.untracked
}
func (s *Status) GetChanged() []string {
	return s.changed
}
func (s *Status) GetWorkspaceChanges() map[string]string {
	return s.workspace_changes
}
func (s *Status) GetIndexChanges() map[string]string {
	return s.index_changes
}
func (s *Status) GetStates() map[string]os.FileInfo {
	return s.states
}
func (s *Status) GetHeadTree() map[string]internals.Entry {
	return s.headTree
}
func (s *Status) GetStatusSize() int {
	return s.statusSize
}
