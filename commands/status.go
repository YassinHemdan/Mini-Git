package commands

import (
	"JIT/commands/utils"
	"JIT/internals"
	"fmt"
	"os"
	"slices"
)

type statusHelper struct {
	ctx       *CommandContext
	repo      *internals.Repository
	untracked []string
}

func StatusCommand(ctx *CommandContext) {
	helper := &statusHelper{ctx: ctx}
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

	_, err = h.repo.Index().Load()
	if err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Couldn't load index: %v\n", err)
		h.ctx.Status = 128
		return
	}

	err = h.scan("")
	if err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Couldn't scan workspace: %v\n", err)
		h.ctx.Status = 128
		return
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
