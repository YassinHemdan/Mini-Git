package commands

import (
	"JIT/commands/utils"
	"JIT/internals"
	database "JIT/internals/database"
	"JIT/internals/index"
	colorUtil "JIT/utils"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
)

const (
	DELETED  = "deleted"
	MODIFIED = "modified"
	ADDED    = "new file"
	GREEN    = "green"
	RED      = "red"
)

var shortStatusMap map[string]string
var longStatusMap map[string]string

type statusHelper struct {
	ctx               *CommandContext
	repo              *internals.Repository
	untracked         []string
	changed           []string // printing order
	workspace_changes map[string]string
	index_changes     map[string]string
	states            map[string]os.FileInfo
	headTree          map[string]database.Entry
	statusSize        int
}

func StatusCommand(ctx *CommandContext) {
	helper := &statusHelper{
		ctx:               ctx,
		states:            make(map[string]os.FileInfo),
		index_changes:     make(map[string]string),
		workspace_changes: make(map[string]string),
		headTree:          make(map[string]database.Entry),
	}
	shortStatusMap = make(map[string]string)
	shortStatusMap[DELETED] = "D"
	shortStatusMap[MODIFIED] = "M"
	shortStatusMap[ADDED] = "A"

	longStatusMap = make(map[string]string)
	longStatusMap[DELETED] = "deleted"
	longStatusMap[MODIFIED] = "modified"
	longStatusMap[ADDED] = "new file"
	helper.run()
}

func (h *statusHelper) run() {
	repo, err := utils.Repo(h.ctx.Dir)
	if err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Can't initialize repository: %v\n", err)
		h.ctx.Status = 128
		return
	}
	h.repo = repo

	if _, err := h.repo.Index().LoadForUpdate(); err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Couldn't load index: %v\n", err)
		h.ctx.Status = 128
		return
	}

	if err := h.scan(""); err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Couldn't scan workspace: %v\n", err)
		h.ctx.Status = 128
		return
	}

	if err := h.loadHeadTree(); err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Couldn't Load HeadTree: %v\n", err)
		h.ctx.Status = 128
		return
	}

	if err := h.checkIndexEntries(); err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Couldn't check index entries: %v\n", err)
		h.ctx.Status = 128
		return
	}

	h.collectDeletedHeadFiles()
	h.printStatus()

	if err := h.repo.Index().WriteUpdates(); err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Couldn't update Index: %v\n", err)
		h.ctx.Status = 128
		return
	}

}

func (h *statusHelper) printStatus() {
	if len(h.ctx.Args) == 0 || h.ctx.Args[0] != "--porcelain" {
		h.printLongFormat()
	} else {
		h.printPorcelainFormat()
	}
}

func (h *statusHelper) printPorcelainFormat() {
	slices.Sort(h.changed)

	// we might have a file that got added and modified at the same time
	// in that case it might be in our slice more than once
	visited := make(map[string]struct{})

	for _, pathname := range h.changed {
		if _, ok := visited[pathname]; !ok {
			fmt.Fprintf(h.ctx.Stdout, "%s %s\n", h.getFileStatus(pathname), pathname)
		}
		visited[pathname] = struct{}{}
	}
	slices.Sort(h.untracked)
	for _, pathname := range h.untracked {
		fmt.Fprintf(h.ctx.Stdout, "?? %s\n", pathname)
	}
}
func (h *statusHelper) printLongFormat() {
	message := "On branch master"
	message +=
		h.indexChangesMessage() +
			h.workspaceChangesMessage() +
			h.untrackedMessage() +
			h.commitMessage() +
			"\n"
	fmt.Fprintf(h.ctx.Stdout, "%s", message)
}
func (h *statusHelper) scan(dirName string) error {
	dirEntriesMap, err := h.repo.Workspace().ListDir(dirName)
	if err != nil {
		return err
	}
	for entryName, entryInfo := range dirEntriesMap {
		if !h.repo.Index().IsTracked(entryName) { // not tracked, just add it
			// we want to make sure that it is not an empty dir
			found, err := h.isTrackableFile(entryName, entryInfo)
			if err != nil {
				return err
			}

			if found { // we can add it .. either a file or non-empty dir
				if entryInfo.IsDir() {
					entryName += "/"
				}
				h.untracked = append(h.untracked, entryName)
			}
			continue
		}

		//if a directory marked as tracked, there might be inner files that still untracked we need to find them (if exists)
		if entryInfo.IsDir() {
			if err := h.scan(entryName); err != nil {
				return err
			}
		} else {
			h.states[entryName] = entryInfo
		}
	}
	return nil
}

func (h *statusHelper) isTrackableFile(entryName string, entryInfo os.FileInfo) (bool, error) {
	/*
		a BFS algorithm to check of the current entry has any nested file inside it
		- if the entry itself is a file, return true
		- if the the entry is a directory, we will run bfs to expand more

		we could use DFS here, but I wanted to get a file ASAP
	*/

	if !entryInfo.IsDir() {
		return true, nil
	}
	queue := []string{entryName}
	for len(queue) != 0 {
		dirName := queue[0]
		queue = queue[1:]

		childEntries, err := h.repo.Workspace().ListDir(dirName)
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

func (h *statusHelper) checkIndexEntries() error {
	indexEntries := h.repo.Index().GetEntries()
	for _, entry := range indexEntries {
		if err := h.checkIndexAgainstWorkspace(entry); err != nil {
			return err
		}
		if err := h.checkIndexAgainstHeadTree(entry); err != nil {
			return err
		}
	}
	return nil
}

func (h *statusHelper) checkIndexAgainstWorkspace(entry *index.IndexEntry) error {
	info, ok := h.states[entry.GetPathname()] // exists in workspace ? check if it is modified
	if !ok {
		// in index but not in workspace ? it means that it got deleted
		h.recordChange(entry.GetPathname(), h.workspace_changes, DELETED)
		return nil
	}

	// check the stat
	stat := info.Sys().(*syscall.Stat_t)
	if !entry.IsMatchedStat(stat) {
		h.recordChange(entry.GetPathname(), h.workspace_changes, MODIFIED)
		return nil
	}
	if !entry.IsMatchedTime(stat) {
		/*
			if the timestamps got changed, that does not mean the content got changed
			there is a case where we can change the content and then revert back again,
				that means the timestamps changed but the content remains the same
				so we need to check with the oid (the content itself)
		*/
		blob, err := h.createBlob(entry.GetPathname())
		if err != nil {
			return err
		}

		oid, err := h.repo.Database().HashObject(blob)
		if err != nil {
			return err
		}
		if string(oid) != string(entry.GetOid()) {
			h.recordChange(entry.GetPathname(), h.workspace_changes, MODIFIED)
		} else {
			// content not changed but timestamps got changed.
			// Update them so that we don't need to visit them again
			h.repo.Index().UpdateEntryStat(entry, stat)
		}
	}
	return nil
}

func (h *statusHelper) checkIndexAgainstHeadTree(indexEntry *index.IndexEntry) error {
	val, ok := h.headTree[indexEntry.GetPathname()]
	if !ok {
		// not committed before
		h.recordChange(indexEntry.GetPathname(), h.index_changes, ADDED)
	} else {
		// committed before, lets check if its content or mode got changed
		if val.GetMode() != indexEntry.GetMode() || string(val.GetOid()) != string(indexEntry.GetOid()) {
			h.recordChange(indexEntry.GetPathname(), h.index_changes, MODIFIED)
		}
	}
	return nil
}
func (h *statusHelper) collectDeletedHeadFiles() {
	for pathname := range h.headTree {
		if !h.repo.Index().IsTrackedFile(pathname) {
			h.recordChange(pathname, h.index_changes, DELETED)
		}
	}
}
func (h *statusHelper) recordChange(pathname string, changesMap map[string]string, changeType string) {
	h.changed = append(h.changed, pathname)
	changesMap[pathname] = changeType
	h.statusSize = max(h.statusSize, len(changeType))
}

func (h *statusHelper) loadHeadTree() error {
	commitOid, err := h.repo.Refs().ReadHead()
	if err != nil {
		return fmt.Errorf("Could not read Refs - %v\n", err)
	}

	// what if there aren't any commits yet ?
	if len(commitOid) == 0 {
		return nil
	}
	obj, err := h.repo.Database().Load(commitOid)
	if err != nil || obj.Type() != "commit" {
		return fmt.Errorf("Could not load commit object from DB - %v\n", err)
	}

	var loadTree func(oid []byte) (map[string]database.Entry, error)
	loadTree = func(treeOid []byte) (map[string]database.Entry, error) {
		obj, err := h.repo.Database().Load(treeOid)
		if err != nil || obj.Type() != "tree" {
			return nil, fmt.Errorf("Could not load tree object from DB - %v\n", err)
		}

		treeObj := obj.(*database.Tree)
		return treeObj.GetEntries(), nil
	}

	var showTree func(entries map[string]database.Entry, pathname string) error
	showTree = func(entries map[string]database.Entry, pathname string) error {
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
				h.headTree[fullpathname] = entry
			}
		}

		return nil
	}

	commitObj := obj.(*database.Commit)
	loadedTreeEntries, err := loadTree(commitObj.GetTreeOid())
	if err != nil {
		return err
	}
	if err := showTree(loadedTreeEntries, ""); err != nil {
		return err
	}

	return nil
}

func (h *statusHelper) getFileStatus(pathname string) string {
	left, right := " ", " "
	val, ok := h.index_changes[pathname]
	if ok {
		left = shortStatusMap[val]
	}
	val, ok = h.workspace_changes[pathname]
	if ok {
		right = shortStatusMap[val]
	}

	return left + right
}

func (h *statusHelper) createBlob(pathname string) (*database.Blob, error) {
	data, err := h.repo.Workspace().ReadFile(pathname)
	if err != nil {
		return nil, err
	}
	blob := database.NewBlob(data)
	return blob, nil
}

func (h *statusHelper) indexChangesMessage() string {
	if len(h.index_changes) == 0 {
		return ""
	}
	message := "\nChanges to be committed:\n"
	message += "  (use \"git restore --staged <file>...\" to unstage)\n"
	message += h.changedFilesMessage(h.index_changes, GREEN)

	return message

}
func (h *statusHelper) workspaceChangesMessage() string {
	if len(h.workspace_changes) == 0 {
		return ""
	}
	message := "\nChanges not staged for commit:\n"
	message += "  (use \"git add/rm <file>...\" to update what will be committed)\n"
	message += "  (use \"git restore <file>...\" to discard changes in working directory)\n"
	message += h.changedFilesMessage(h.workspace_changes, RED)

	return message
}
func (h *statusHelper) untrackedMessage() string {
	if len(h.untracked) == 0 {
		return ""
	}
	message := "\nUntracked files:\n"
	message += "  (use \"git add <file>...\" to include in what will be committed)\n"

	slices.Sort(h.untracked)
	for _, path := range h.untracked {
		message += fmt.Sprintf("%8s%s\n", "", path)
	}

	return message
}
func (h *statusHelper) changedFilesMessage(changesSet map[string]string, color string) string {
	message := ""
	indexPaths := slices.Collect(maps.Keys(changesSet))
	slices.Sort(indexPaths)

	for _, path := range indexPaths {
		extraSpaces := h.statusSize - len(changesSet[path])
		prefixSpacesOne := strings.Repeat(" ", 8)
		prefixSpacesTwo := strings.Repeat(" ", extraSpaces+3)
		statusMessage := longStatusMap[changesSet[path]]
		message += fmt.Sprintf("%s%s:%s%s\n", prefixSpacesOne, statusMessage, prefixSpacesTwo, path)
	}
	return colorUtil.Format(color, message)
	// return message
}
func (h *statusHelper) commitMessage() string {
	message := ""
	if len(h.index_changes) != 0 {
		return message
	} else if len(h.workspace_changes) != 0 {
		message += "\nno changes added to commit (use \"git add\" and/or \"git commit -a\")"
	} else if len(h.untracked) != 0 {
		message += "\nnothing added to commit but untracked files present (use \"git add\" to track)"
	} else {
		message += "\nnothing to commit, working tree clean"
	}

	return message
}
