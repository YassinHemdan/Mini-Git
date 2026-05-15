package commands

import (
	"JIT/commands/utils"
	"JIT/internals"
	"fmt"
	"path/filepath"
)

type branchCommandHandler struct {
	ctx  *CommandContext
	repo *internals.Repository
}

func BranchCommand(ctx *CommandContext) {
	handler := &branchCommandHandler{ctx: ctx}
	handler.run()
}

func (h *branchCommandHandler) run() {
	repo, err := utils.Repo(h.ctx.Dir)
	if err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Can't initialize repository: %v\n", err)
		h.ctx.Status = 128
		return
	}

	h.repo = repo

	if len(h.ctx.Args) == 0 {
		if err := h.listBranches(); err != nil {
			fmt.Fprintf(h.ctx.Stderr, "Error: Could not list current branches: %v\n", err)
			h.ctx.Status = 128
			return
		}
	} else {
		if err := h.createBranch(); err != nil {
			fmt.Fprintf(h.ctx.Stderr, "%v\n", err)
			h.ctx.Status = 128
			return
		}
	}
}

func (h *branchCommandHandler) listBranches() error {
	headsPath := filepath.Join(h.repo.RepoPath(), "refs", "heads")
	branchs, err := h.repo.Workspace().ListFiles(headsPath)
	if err != nil {
		return err
	}

	for _, branch := range branchs {
		fmt.Fprintf(h.ctx.Stdout, "%s\n", filepath.Base(branch))
	}
	return nil
}

func (h *branchCommandHandler) createBranch() error {
	branchName := h.ctx.Args[0]
	startOid, err := h.repo.Refs().ReadHead()
	if err != nil {
		return err
	}

	if len(h.ctx.Args) >= 2 {
		rev := internals.NewRevision(h.repo, h.ctx.Args[1])
		oid, err := rev.Resolve()
		if err != nil {
			return err
		}

		startOid = oid
	}

	fmt.Printf("%x\n", startOid)
	return h.repo.Refs().CreateBranch(branchName, startOid)
}
