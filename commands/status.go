package commands

import (
	"JIT/commands/utils"
	"JIT/internals"
	"fmt"
	"slices"
)

type statusHelper struct {
	ctx  *CommandContext
	repo *internals.Repository
	// untracked []string
}

func StatusCommand(ctx *CommandContext) {
	// untracked: make([]string, 0)
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

	// listedFiles, err := h.repo.Workspace().ListFiles(h.ctx.Dir)

	// if err != nil {
	// 	fmt.Fprintf(h.ctx.Stderr, "Couldn't list files: %v\n", err)
	// 	h.ctx.Status = 128
	// 	return
	// }

	_, err = h.repo.Index().Load()
	if err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Couldn't load index: %v\n", err)
		h.ctx.Status = 128
		return
	}

	// untracked := make([]string, 0)
	// for _, pathname := range listedFiles {
	// 	if !repo.Index().IsTracked(pathname) {
	// 		untracked = append(untracked, pathname)
	// 	}
	// }

	// slices.Sort(untracked)
	// for _, pathname := range untracked {
	// 	fmt.Fprintf(h.ctx.Stdout, "?? %s\n", pathname)
	// }

	untracked, err := h.getUntracked("")
	if err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Couldn't scan workspace: %v\n", err)
		h.ctx.Status = 128
		return
	}

	slices.Sort(untracked)
	for _, pathname := range untracked {
		fmt.Fprintf(h.ctx.Stdout, "?? %s\n", pathname)
	}
}

func (h *statusHelper) getUntracked(dirName string) ([]string, error) {
	dirEntriesMap, err := h.repo.Workspace().ListDir(dirName)
	if err != nil {
		return nil, err
	}

	isUntrackedDir := true
	possibleResult := make([]string, 0)
	for entryName, entryInfo := range dirEntriesMap {
		if entryInfo.IsDir() {
			subUntracked, err := h.getUntracked(entryName)
			if err != nil {
				return nil, err
			}

			if (len(subUntracked) == 1 && subUntracked[0] != entryName+"/") || len(subUntracked) > 1 {
				isUntrackedDir = false
			}

			possibleResult = append(possibleResult, subUntracked...)
		} else {
			if h.repo.Index().IsTracked(entryName) == false {
				possibleResult = append(possibleResult, entryName)
			} else {
				isUntrackedDir = false
			}
		}
	}

	if isUntrackedDir == false || (isUntrackedDir == true && len(dirName) == 0) {
		return possibleResult, nil
	} else {
		return []string{dirName + "/"}, err
	}
}
