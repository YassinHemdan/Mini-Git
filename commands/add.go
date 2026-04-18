package commands

import (
	"JIT/commands/utils"
	"JIT/internals"
	database "JIT/internals/database"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

type addCommandHandler struct {
	ctx  *CommandContext
	repo *internals.Repository
}

func AddCommand(ctx *CommandContext) {
	handler := &addCommandHandler{ctx: ctx}
	handler.run()
}

func (h *addCommandHandler) run() {
	fmt.Println("Add command called")
	if len(h.ctx.Args) <= 0 {
		fmt.Fprintf(h.ctx.Stderr, "No files provided to add\n")
		h.ctx.Status = 128
		return
	}
	root_dir := h.ctx.Dir
	repo, err := utils.Repo(root_dir)

	if err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Can't initialize repository: %v\n", err)
		h.ctx.Status = 128
		return
	}

	h.repo = repo

	verified, err := h.repo.Index().LoadForUpdate()
	if err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Could not load index: %v\n", err)
		h.ctx.Status = 128
		return
	}

	fmt.Println(verified)

	filesToAdd, err := h.expandedPaths()
	if err != nil {
		h.handleError(err)
		return
	}

	for _, fileName := range filesToAdd {
		if err := h.addToIdx(fileName); err != nil {
			h.handleError(err)
			return
		}
		fmt.Printf("Adding file: %s\n", fileName)
	}

	if err := h.repo.Index().WriteUpdates(); err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Could not update index file: %v\n", err)
		h.ctx.Status = 128
		return
	}

	h.ctx.Status = 0
}

func (h *addCommandHandler) addToIdx(path string) error {
	file_content, err := h.repo.Workspace().ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return &utils.NoPermissionError{Path: path, Err: err}
		}
		return err
	}

	isExecutable, err := (func() (bool, error) {
		fileInfo, err := h.repo.Workspace().GetFileState(path)
		if err != nil {
			return false, err
		}

		if fileInfo.Mode()&0111 != 0 {
			return true, nil
		}
		return false, nil
	})()

	if err != nil {
		return err
	}

	blob := database.Blob{}
	if err := blob.New(file_content, path, isExecutable); err != nil {
		return err
	}

	if err := h.repo.Database().Store(&blob); err != nil {
		return err
	}

	fileInfo, err := h.repo.Workspace().GetFileState(path)
	if err != nil {
		return err
	}

	stat, ok := fileInfo.Sys().(*syscall.Stat_t)

	if !ok {
		return fmt.Errorf("could not get system stat for %s", path)
	}

	if err := h.repo.Index().Add(path, blob.GetOid(), stat); err != nil {
		return err
	}

	return nil
	// fmt.Printf("Adding file: %s\n", fileName)
}

func (h *addCommandHandler) expandedPaths() ([]string, error) {
	filesToAdd := make([]string, 0)
	for _, path := range h.ctx.Args {
		fullpath := filepath.Join(h.ctx.Dir, path)
		listedFiles, err := h.repo.Workspace().ListFiles(fullpath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, &utils.MissingFileError{Path: path, Err: err}
			}
			return nil, err
		}
		filesToAdd = append(filesToAdd, listedFiles...)
	}
	return filesToAdd, nil
}

func (h *addCommandHandler) handleError(err error) {
	var missingFile *utils.MissingFileError
	var noPermission *utils.NoPermissionError

	switch {
	case errors.As(err, &missingFile):
		h.handleMissingFile(missingFile)
	case errors.As(err, &noPermission):
		h.handleUnreadableFile(noPermission)
	default:
		fmt.Fprintf(h.ctx.Stderr, "fatal: %v\n", err)
		h.repo.Index().ReleaseLock()
		h.ctx.Status = 128
	}
}

func (h *addCommandHandler) handleMissingFile(err *utils.MissingFileError) {
	fmt.Fprintf(h.ctx.Stderr, "fatal: %s\n", err.Error())
	h.repo.Index().ReleaseLock()
	h.ctx.Status = 128
}

func (h *addCommandHandler) handleUnreadableFile(err *utils.NoPermissionError) {
	fmt.Fprintf(h.ctx.Stderr, "error: %s\n", err.Error())
	fmt.Fprintf(h.ctx.Stderr, "fatal: adding files failed\n")
	h.repo.Index().ReleaseLock()
	h.ctx.Status = 128
}
