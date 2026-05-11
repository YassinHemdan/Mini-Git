package commands

import (
	"JIT/commands/utils"
	database "JIT/internals/database"
	"JIT/internals/index"
	"fmt"
	"time"
)

type ConfigVariables struct {
	JIT_AUTHOR_NAME  string
	JIT_AUTHOR_EMAIL string
}

func CommitCommand(ctx *CommandContext) {
	// if len(ctx.Args)
	if err := validate(ctx.Args); err != nil {
		fmt.Fprintf(ctx.Stderr, "Invalid arguments provided: %v\n", err)
		ctx.Status = 1
		return
	}
	config := ConfigVariables{}
	config.JIT_AUTHOR_NAME = ctx.Env["JIT_AUTHOR_NAME"]
	config.JIT_AUTHOR_EMAIL = ctx.Env["JIT_AUTHOR_EMAIL"]

	author := database.Author{}
	if err := author.New(config.JIT_AUTHOR_NAME, config.JIT_AUTHOR_EMAIL, time.Now()); err != nil {
		fmt.Fprintf(ctx.Stderr, "failed to create author: %v\n", err)
		ctx.Status = 1
		return
	}
	root_dir := ctx.Dir
	repo, err := utils.Repo(root_dir)

	if err != nil {
		fmt.Fprintf(ctx.Stderr, "Can't initialize repository: %v\n", err)
		ctx.Status = 128
		return
	}

	// notice here that we only need to Load the index not to LoadForUpdate
	if _, err = repo.Index().Load(); err != nil {
		fmt.Fprintf(ctx.Stderr, "Failed to load index: %v\n", err)
		ctx.Status = 1
		return
	}

	entries := toTreeEntry(repo.Index().GetEntries())
	merkleTree := database.BuildTree(entries)

	merkleTree.Traverse(func(t *database.Tree) {
		if err := repo.Database().Store(t); err != nil {
			fmt.Fprintf(ctx.Stderr, "Can't save tree in db: %v\n", err)
			ctx.Status = 1
			return
		}
	})

	parent_id, err := repo.Refs().ReadHead()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "Failed to read HEAD file: %v\n", err)
		ctx.Status = 1
		return
	}

	commit := database.Commit{}
	message := ctx.Args[1]
	if err := commit.New(parent_id, merkleTree.GetOid(), message, author, author); err != nil {
		fmt.Fprintf(ctx.Stderr, "Failed to create a commit object: %v\n", err)
		ctx.Status = 1
		return
	}

	if err := repo.Database().Store(&commit); err != nil {
		fmt.Fprintf(ctx.Stderr, "Failed to store the commit object: %v\n", err)
		ctx.Status = 1
		return
	}

	if err := repo.Refs().UpdateHead(commit.GetOid()); err != nil {
		fmt.Fprintf(ctx.Stderr, "Failed to update HEAD file: %v\n", err)
		ctx.Status = 1
		return
	}


	short_id := repo.Database().ShortId(commit.GetOid())
	is_root := (func() string {
		if len(parent_id) == 0 {
			return "(root_commit) "
		}
		return ""
	})()

	fmt.Printf("[master %s%s] %s\n", is_root, short_id, commit.GetMessage())
	ctx.Status = 0
}

func validate(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("Please provide a commit message")
	}
	if args[0] != "-m" {
		return fmt.Errorf("Invalid flag provided")
	}

	if len(args[1]) == 0 {
		return fmt.Errorf("No commit message provided")
	}
	return nil
}

func toTreeEntry(entries []*index.IndexEntry) []database.BuildEntry {
	treeEntries := make([]database.BuildEntry, 0)

	for _, entry := range entries {
		treeEntries = append(treeEntries, entry)
	}

	return treeEntries
}
