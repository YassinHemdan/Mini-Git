package commands

import (
	"JIT/commands/utils"
	"JIT/internals"
	"fmt"
)

type checkoutCommandHandler struct {
	ctx  *CommandContext
	repo *internals.Repository
}

func CheckoutCommand(ctx *CommandContext) {
	handler := &checkoutCommandHandler{ctx: ctx}
	handler.run()
}

func (h *checkoutCommandHandler) run() {
	root_dir := h.ctx.Dir
	repo, err := utils.Repo(root_dir)

	if err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Can't initialize repository: %v\n", err)
		h.ctx.Status = 128
		return
	}
	h.repo = repo
	target := h.ctx.Args[0]

	revision := internals.NewRevision(repo, target)
	targetOid, err := revision.Resolve(internals.REVISION_COMMIT)
	if err != nil { // TODO: handle the invalidObject error
		fmt.Fprintf(h.ctx.Stderr, "Can't resolve revision: %v\n", err)
		h.ctx.Status = 128
		return
	}

	currentOid, err := repo.Refs().ReadHead()
	if err != nil { // TODO: handle the invalidObject error
		fmt.Fprintf(h.ctx.Stderr, "Can't read HEAD: %v\n", err)
		h.ctx.Status = 128
		return
	}

	diff, err := repo.Database().TreeDiff(currentOid, targetOid)
	if err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Can't get TreeDiff: %v\n", err)
		h.ctx.Status = 128
		return
	}

	migration := repo.Migration(diff)
	if err := migration.ApplyChanges(); err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Migration Failed: %v\n", err)
		h.ctx.Status = 128
		return
	}

	if err := repo.Refs().UpdateHead(targetOid); err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Can't update HEAD %v\n", err)
		h.ctx.Status = 128
		return
	}
}
