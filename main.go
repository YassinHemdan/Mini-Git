

package main

import (
	"JIT/commands"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not load env - %v", err)
		os.Exit(1)
	}

	env := make(map[string]string)
	for _, e := range os.Environ() { // a []strings of envs ["GREETING=hello", "NAME=yassin", ....]
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	dir, _ := os.Getwd()
	ctx := commands.Execute(dir, env, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	os.Exit(ctx.Status)
}
