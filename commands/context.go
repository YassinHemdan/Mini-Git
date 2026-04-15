package commands

import (
	"io"
	"path/filepath"
)

type CommandContext struct {
	Dir    string
	Env    map[string]string
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Status int
}

func NewCommandContext(
	dir string,
	env map[string]string,
	args []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer) *CommandContext {
	return &CommandContext{
		Dir:    dir,
		Env:    env,
		Args:   args,
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
}

func (ctx *CommandContext) ExpandedPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(ctx.Dir, path)
}
