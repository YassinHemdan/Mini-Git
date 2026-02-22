package cmd

import (
	"Jit/internals"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
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

		// read the the content of the cur directory and store them
		entries, err := os.ReadDir(root_dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Can't read the current directory - %v\n", err)
			os.Exit(1)
		}

		var tree_entries []internals.Entry

		for _, entry := range entries {
			// for now, we will only target the fiels "blobs"
			if entry.Name() == "." || entry.Name() == ".." || entry.IsDir() {
				continue
			}

			file, err := os.Open(entry.Name())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Can't open the cur file - %v\n", err)
				os.Exit(1)
			}

			defer file.Close()

			var file_content bytes.Buffer

			if _, err := io.Copy(&file_content, file); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Can't copy/read the cur file - %v\n", err)
				os.Exit(1)
			}

			blob := internals.Blob{}
			if err := blob.New(file_content.Bytes()); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Can't create a blob for the file - %v\n", err)
				os.Exit(1)
			}

			if err := db.Store(&blob); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Can't store the current object/blob - %v\n", err)
				os.Exit(1)
			}

			tree_entry := internals.Entry{}
			if err := tree_entry.New(blob.GetOid(), file.Name()); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Can't create an entry - %v\n", err)
				os.Exit(1)
			}

			tree_entries = append(tree_entries, tree_entry)
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
		fmt.Printf("[(root_commit)%s] %s\n", short_id, commit.GetMessage())
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)

	commitCmd.Flags().StringVarP(&message, "message", "m", "", "Use this given <msg> as the commit message.")
}
