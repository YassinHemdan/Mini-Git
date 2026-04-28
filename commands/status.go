package commands

import (
	"JIT/commands/utils"
	"JIT/internals"
	database "JIT/internals/database"
	"JIT/internals/index"
	"fmt"
	"os"
	"slices"
	"syscall"
)

const (
	WORKSPACE_MODIFIED = "workspace_modified"
	WORKSPACE_DELETED  = "workspace_deleted"
)

type statusHelper struct {
	ctx       *CommandContext
	repo      *internals.Repository
	untracked []string
	changed   []string // printing order
	changes   map[string]map[string]bool
	states    map[string]os.FileInfo
}

func StatusCommand(ctx *CommandContext) {
	helper := &statusHelper{ctx: ctx, states: make(map[string]os.FileInfo), changes: make(map[string]map[string]bool)}
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

	h.detectChanges()
	h.printStatus()

	if err := h.repo.Index().WriteUpdates(); err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Couldn't update Index: %v\n", err)
		h.ctx.Status = 128
		return
	}

}

func (h *statusHelper) printStatus() {
	slices.Sort(h.changed)
	for _, pathname := range h.changed {
		fmt.Fprintf(h.ctx.Stdout, "%s %s\n", h.getFileStatus(pathname), pathname)
	}
	slices.Sort(h.untracked)
	for _, pathname := range h.untracked {
		fmt.Fprintf(h.ctx.Stdout, "?? %s\n", pathname)
	}
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

func (h *statusHelper) detectChanges() {
	indexEntries := h.repo.Index().GetEntries()
	for _, entry := range indexEntries {
		h.checkIndexEntry(entry)
	}
}

func (h *statusHelper) checkIndexEntry(entry *index.IndexEntry) error {
	info, ok := h.states[entry.GetPathname()] // exists in workspace ? check if it is modified
	if !ok {
		// in index but not in workspace ? it means that it got deleted
		h.recordChange(entry.GetPathname(), WORKSPACE_DELETED)
		return nil
	}

	// check the stat
	stat := info.Sys().(*syscall.Stat_t)
	if !entry.IsMatchedStat(stat) {
		h.recordChange(entry.GetPathname(), WORKSPACE_MODIFIED)
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
			h.recordChange(entry.GetPathname(), WORKSPACE_MODIFIED)
		} else {
			// content not changed but timestamps got changed.
			// Update them so that we don't need to visit them again
			h.repo.Index().UpdateEntryStat(entry, stat)
		}
	}
	return nil
}

func (h *statusHelper) recordChange(pathname, changeType string) {
	if _, ok := h.changes[pathname]; !ok {
		h.changes[pathname] = make(map[string]bool)
	}

	h.changed = append(h.changed, pathname)
	h.changes[pathname][changeType] = true
}

func (h *statusHelper) getFileStatus(pathname string) string {
	status := "  "
	if h.changes[pathname][WORKSPACE_DELETED] {
		status = " D"
	}
	if h.changes[pathname][WORKSPACE_MODIFIED] {
		status = " M"
	}

	return status
}

func (h *statusHelper) createBlob(pathname string) (*database.Blob, error) {
	data, err := h.repo.Workspace().ReadFile(pathname)
	if err != nil {
		return nil, err
	}
	blob := database.NewBlob(data)
	return blob, nil
}
