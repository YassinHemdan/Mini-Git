package cmd

import (
	"Jit/internals"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type ConfigVariables struct {
	JIT_AUTHOR_NAME  string
	JIT_AUTHOR_EMAIL string
}

var (
	message string
)
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	Run: func(cmd *cobra.Command, args []string) {
		err := godotenv.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load env variables - %v\n", err)
			os.Exit(1)
		}
		config := ConfigVariables{}
		config.JIT_AUTHOR_NAME = os.Getenv("G_AUTHOR_NAME")
		config.JIT_AUTHOR_EMAIL = os.Getenv("G_AUTHOR_EMAIL")

		author := internals.Author{}

		if err := author.New(config.JIT_AUTHOR_NAME, config.JIT_AUTHOR_EMAIL, time.Now()); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create a new author - %v\n", err)
			os.Exit(1)
		}

		root_dir, err := os.Getwd()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Can't fetch current directory - %v\n", err)
			os.Exit(1)
		}

		jit_dir := strings.Join([]string{root_dir, internals.JitMetadataDir}, string(os.PathSeparator))
		db_dir := strings.Join([]string{jit_dir, "objects"}, string(os.PathSeparator))

		db := internals.Database{}
		if err := db.New(db_dir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Can't create a new database - %v\n", err)
			os.Exit(1)
		}

		workspace := internals.Workspace{}

		if err := workspace.New(root_dir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Can't read the current directory - %v\n", err)
			os.Exit(1)
		}

		tree_entries, err := saveSubTrees(workspace.GetDirEntries(), root_dir, &workspace, &db)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Can't save sub trees - %v\n", err)
			os.Exit(1)
		}

		tree := internals.Tree{}
		if err := tree.New(tree_entries); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Can't create a tree - %v\n", err)
			os.Exit(1)
		}
		if err := db.Store(&tree); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Can't store the current object/tree - %v\n", err)
			os.Exit(1)
		}

		refs := internals.Refs{}

		if err := refs.New(jit_dir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create a Refs - %v\n", err)
			os.Exit(1)
		}

		parent_id, err := refs.ReadHead()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to read HEAD file - %v\n", err)
			os.Exit(1)
		}

		commit := internals.Commit{}
		if err := commit.New(parent_id, tree.GetOid(), message, author, author); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create a commit object - %v\n", err)
			os.Exit(1)
		}

		if err := db.Store(&commit); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to store the commit object - %v\n", err)
			os.Exit(1)
		}

		if err := refs.UpdateHead(commit.GetOid()); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to update HEAD file - %v\n", err)
			os.Exit(1)
		}

		short_id := hex.EncodeToString(commit.GetOid())[:7]
		is_root := (func() string {
			if len(parent_id) == 0 {
				return "(root_commit) "
			}
			return ""
		})()

		fmt.Println("Num Of changes = ", db.GetChanges())
		fmt.Printf("[master %s%s] %s\n", is_root, short_id, commit.GetMessage())
	},
}

func saveSubTrees(dirEntries []os.DirEntry, path string, workspace *internals.Workspace, db *internals.Database) ([]internals.Entry, error) {
	var treeEntries []internals.Entry
	// fmt.Println("Hi")

	for _, entry := range dirEntries {
		if entry.Name() == "." || entry.Name() == ".." {
			continue
		}

		fullpath := filepath.Join(path, entry.Name())
		treeEntry := internals.Entry{}
		// fmt.Println(entry.Name())
		if entry.IsDir() {

			subOSEntries, err := workspace.GetDirEntriesWithName(fullpath)

			if err != nil {
				return nil, fmt.Errorf("Could not read directory - %v", err)
			}

			subTreeEntries, err := saveSubTrees(subOSEntries, fullpath, workspace, db)
			if err != nil {
				return nil, fmt.Errorf("Could not save inner trees - %v", err)
			}
			tree := internals.Tree{}
			if err := tree.New(subTreeEntries); err != nil {
				return nil, fmt.Errorf("Could not create a tree - %v", err)
			}

			// fmt.Printf("Saving Tree %s\n", entry.Name())
			if err := db.Store(&tree); err != nil {
				return nil, fmt.Errorf("Could not save tree - %v", err)
			}

			fileMode := workspace.GetDirState()

			if err := treeEntry.New(tree.GetOid(), entry.Name(), fileMode); err != nil {
				return nil, fmt.Errorf("Could not create an entry - %v", err)
			}
		} else {
			// it is a file, save it as a blob

			fileContent, err := workspace.ReadFile(fullpath)

			if err != nil {
				return nil, fmt.Errorf("Could not read file - %v", err)
			}

			blob := internals.Blob{}

			if err := blob.New(fileContent); err != nil {
				return nil, fmt.Errorf("Could not create a new blob - %v", err)
			}

			// fmt.Printf("Saving blob %s\n", entry.Name())
			if err := db.Store(&blob); err != nil {
				return nil, fmt.Errorf("Could not save blob to db - %v", err)
			}

			fileInfo, err := workspace.GetFileState(fullpath)

			if err := treeEntry.New(blob.GetOid(), entry.Name(), fileInfo.Mode().Perm()); err != nil {
				return nil, fmt.Errorf("Could not create an entry - %v", err)
			}

		}
		treeEntries = append(treeEntries, treeEntry)
	}

	return treeEntries, nil
}
func init() {
	rootCmd.AddCommand(commitCmd)

	commitCmd.Flags().StringVarP(&message, "message", "m", "", "Use this given <msg> as the commit message.")
}
