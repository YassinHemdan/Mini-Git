package commands

import (
	"JIT/commands/utils"
	"JIT/internals"
	colorUtil "JIT/utils"
	"fmt"
	"maps"
	"slices"
	"strings"
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
	ctx    *CommandContext
	repo   *internals.Repository
	status *internals.Status
}

func StatusCommand(ctx *CommandContext) {
	helper := &statusHelper{
		ctx: ctx,
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

	h.status, err = repo.Status() // business logic
	if err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Couldn't load status: %v\n", err)
		h.ctx.Status = 128
		return
	}

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
	slices.Sort(h.status.GetChanged())

	// we might have a file that got added and modified at the same time
	// in that case it might be in our slice more than once
	visited := make(map[string]struct{})

	for _, pathname := range h.status.GetChanged() {
		if _, ok := visited[pathname]; !ok {
			fmt.Fprintf(h.ctx.Stdout, "%s %s\n", h.getFileStatus(pathname), pathname)
		}
		visited[pathname] = struct{}{}
	}
	slices.Sort(h.status.GetUntracked())
	for _, pathname := range h.status.GetUntracked() {
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
func (h *statusHelper) getFileStatus(pathname string) string {
	left, right := " ", " "
	val, ok := h.status.GetIndexChanges()[pathname]
	if ok {
		left = shortStatusMap[val]
	}
	val, ok = h.status.GetWorkspaceChanges()[pathname]
	if ok {
		right = shortStatusMap[val]
	}

	return left + right
}
func (h *statusHelper) indexChangesMessage() string {
	if len(h.status.GetIndexChanges()) == 0 {
		return ""
	}
	message := "\nChanges to be committed:\n"
	message += "  (use \"git restore --staged <file>...\" to unstage)\n"
	message += h.changedFilesMessage(h.status.GetIndexChanges(), GREEN)

	return message

}
func (h *statusHelper) workspaceChangesMessage() string {
	if len(h.status.GetWorkspaceChanges()) == 0 {
		return ""
	}
	message := "\nChanges not staged for commit:\n"
	message += "  (use \"git add/rm <file>...\" to update what will be committed)\n"
	message += "  (use \"git restore <file>...\" to discard changes in working directory)\n"
	message += h.changedFilesMessage(h.status.GetWorkspaceChanges(), RED)

	return message
}
func (h *statusHelper) untrackedMessage() string {
	if len(h.status.GetUntracked()) == 0 {
		return ""
	}
	message := "\nUntracked files:\n"
	message += "  (use \"git add <file>...\" to include in what will be committed)\n"

	slices.Sort(h.status.GetUntracked())
	for _, path := range h.status.GetUntracked() {
		message += fmt.Sprintf("%8s%s\n", "", path)
	}

	return message
}
func (h *statusHelper) changedFilesMessage(changesSet map[string]string, color string) string {
	message := ""
	indexPaths := slices.Collect(maps.Keys(changesSet))
	slices.Sort(indexPaths)

	for _, path := range indexPaths {
		extraSpaces := h.status.GetStatusSize() - len(changesSet[path])
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
	if len(h.status.GetIndexChanges()) != 0 {
		return message
	} else if len(h.status.GetWorkspaceChanges()) != 0 {
		message += "\nno changes added to commit (use \"git add\" and/or \"git commit -a\")"
	} else if len(h.status.GetUntracked()) != 0 {
		message += "\nnothing added to commit but untracked files present (use \"git add\" to track)"
	} else {
		message += "\nnothing to commit, working tree clean"
	}

	return message
}
