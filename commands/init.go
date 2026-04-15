package commands

import (
	"JIT/internals/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func InitCommand(ctx *CommandContext) {
	root := ctx.Dir
	if len(ctx.Args) > 0 {
		root = ctx.ExpandedPath(ctx.Args[0])
	}

	jit_dir := filepath.Join(root, utils.JitMetadataDir)

	if err := os.MkdirAll(jit_dir, utils.JitDefaultPermission); err != nil {
		fmt.Fprintf(ctx.Stderr, "Error: Can't create jit directory - %v\n", err)
		ctx.Status = 1
		return
	}

	for _, content := range strings.Split(utils.JitMetadataContent, "|") {
		filepath := filepath.Join(jit_dir, content)
		if err := os.Mkdir(filepath, utils.JitDefaultPermission); err != nil {
			fmt.Fprintf(ctx.Stderr, "Error: Can't create %s in .jit - %v\n", filepath, err)
			ctx.Status = 1
			return
		}
	}

	fmt.Fprintf(ctx.Stdout, "Initialized empty Jit repository in %s\n", jit_dir)
	ctx.Status = 0
}
