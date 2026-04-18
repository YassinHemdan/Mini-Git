package commands

import (
	"fmt"
	"io"
)

type CommandFunc func(ctx *CommandContext)

var COMMANDS = map[string]CommandFunc{
	"init":   InitCommand,
	"add":    AddCommand,
	"commit": CommitCommand,
	"status": StatusCommand,
}

func Execute(
	dir string,
	env map[string]string,
	argv []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) *CommandContext {
	if len(argv) == 0 {
		fmt.Fprintf(stderr, "No command given\n")
		return &CommandContext{Status: 1}
	}

	commandName := argv[0]
	args := argv[1:]

	cmdFunc, ok := COMMANDS[commandName]
	if !ok {
		fmt.Fprintf(stderr, "%s is not a jit command\n", commandName)
		return &CommandContext{Status: 1}
	}
	ctx := NewCommandContext(dir, env, args, stdin, stdout, stderr)
	cmdFunc(ctx)

	return ctx
}
