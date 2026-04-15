package commands

import (
	"JIT/internals"
	database "JIT/internals/database"
	"JIT/internals/utils"
	"fmt"
	"path/filepath"
	"syscall"
)

func AddCommand(ctx *CommandContext) {
	if len(ctx.Args) <= 0 {
		fmt.Fprintf(ctx.Stderr, "No files provided to add\n")
		ctx.Status = 1
		return
	}
	root_dir := ctx.Dir

	jit_dir := filepath.Join(root_dir, utils.JitMetadataDir)
	repo, err := internals.NewRepository(jit_dir)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "Can't initialize repository: %v\n", err)
		ctx.Status = 1
		return
	}

	filesToAdd := make([]string, 0)

	for _, path := range ctx.Args {

		fullpath := filepath.Join(root_dir, path)
		files, err := repo.Workspace().ListFiles(fullpath)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "Could not list files: %v\n", err)
			ctx.Status = 1
			return
		}

		filesToAdd = append(filesToAdd, files...)
	}

	verified, err := repo.Index().LoadForUpdate()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "Could not load index: %v\n", err)
		ctx.Status = 1
		return
	}

	fmt.Println(verified)

	for _, fileName := range filesToAdd {
		defer repo.Index().ReleaseLock()

		file_content, err := repo.Workspace().ReadFile(fileName)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "Could not read file %s: %v\n", fileName, err)
			ctx.Status = 1
			return
		}

		isExecutable, err := (func() (bool, error) {
			fileInfo, err := repo.Workspace().GetFileState(fileName)
			if err != nil {
				return false, err
			}

			if fileInfo.Mode()&0111 != 0 {
				return true, nil
			}
			return false, nil
		})()

		if err != nil {
			fmt.Fprintf(ctx.Stderr, "Could not get file %s state: %v\n", fileName, err)
			ctx.Status = 1
			return
		}

		blob := database.Blob{}
		if err := blob.New(file_content, fileName, isExecutable); err != nil {
			fmt.Fprintf(ctx.Stderr, "Could not create a new blob: %v\n", err)
			ctx.Status = 1
			return
		}

		if err := repo.Database().Store(&blob); err != nil {
			fmt.Fprintf(ctx.Stderr, "Could not store the blob: %v\n", err)
			ctx.Status = 1
			return
		}

		fileInfo, err := repo.Workspace().GetFileState(fileName)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "Can't get file state: %v\n", err)
			ctx.Status = 1
			return
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)

		if !ok {
			fmt.Fprintf(ctx.Stderr, "Error: Could not get file's stat\n")
			ctx.Status = 1
			return
		}

		if err := repo.Index().Add(fileName, blob.GetOid(), stat); err != nil {
			fmt.Fprintf(ctx.Stderr, "Could not add file: %v\n", err)
			ctx.Status = 1
			return
		}

		fmt.Printf("Adding file: %s\n", fileName)
	}

	if err := repo.Index().WriteUpdates(); err != nil {
		fmt.Fprintf(ctx.Stderr, "Could not update index file: %v\n", err)
		ctx.Status = 1
		return
	}

	ctx.Status = 0
}
