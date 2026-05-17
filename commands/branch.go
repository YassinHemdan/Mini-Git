package commands

import (
	"JIT/commands/utils"
	"JIT/internals"
	"errors"
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
			fmt.Fprintf(h.ctx.Stderr, "%s\n", err.Error())
			h.ctx.Status = 128
			return
		}
	} else {
		if err := h.createBranch(); err != nil {
			fmt.Fprintf(h.ctx.Stderr, "%s\n", err.Error())
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
	var invalidObject *internals.InvalidObject
	var invalidBranch *internals.InvalidBranch

	branchName := h.ctx.Args[0]
	startOid, err := h.repo.Refs().ReadHead()
	if err != nil {
		return err
	}

	if err := h.repo.Refs().IsValidBranchName(branchName); err != nil {
		if errors.As(err, &invalidBranch) {
			return fmt.Errorf("fatal: %s", err.Error())
		}
		return err
	}

	if len(h.ctx.Args) >= 2 {
		rev := internals.NewRevision(h.repo, h.ctx.Args[1])
		oid, err := rev.Resolve(internals.REVISION_COMMIT)
		if err != nil {
			if errors.As(err, &invalidObject) {
				hintedErrors := rev.HintedErrors()
				errorMessage := ""
				for _, hintedError := range hintedErrors {
					errorMessage += fmt.Sprintf("%s\n", hintedError.Message)
					for _, hint := range hintedError.Hints {
						errorMessage += fmt.Sprintf("hint: %s\n", hint)
					}
				}

				return fmt.Errorf("%sfatal: %s", errorMessage, err.Error())
			}
			return err
		}

		startOid = oid
	}

	if err := h.repo.Refs().CreateBranch(branchName, startOid); err != nil {
		if errors.As(err, &invalidBranch) {
			return fmt.Errorf("fatal: %s", err.Error())
		}
		return err
	}

	return nil
	/*
		Notice something: let say we ran this command, git branch ^alice bob^^
		- the branch name we are trying to create is not valid (^alice)
		- also the bob branch does not exist

		so we have two errors here, one for the InvalidBranch and another one from InvalidObject

		but git prioritize the InvalidBranch over InvalidObject, but here, we are running our revision
			first and if it passes, we go and create our branch, so the InvalidObject error will appear
		so in order to solve this , we can go and create the branch first, with no oid provided, and
		if the branch name is valid and does not already exist, an empty file will be created,
		then we go and run our revision, if it passes, we then take the oid from it and put it in our created file
		and if the revision did not pass, we rollback and delete the file created

		this also good for performance because we might run our revision which is a recursive procedure
		and then at the end we find out that we cannot create the branch because its name was invalid from the beginning


		another way to do it by validating the branchname alone before creating anything and if it fails we
		return the InvalidBranch error
	*/
}
