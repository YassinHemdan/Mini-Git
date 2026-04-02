package cmd

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	internals "JIT/internals"
	database "JIT/internals/database"

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

		author := database.Author{}

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

		filesPathNames, err := workspace.ListFiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Can't list files of the current directory - %v\n", err)
			os.Exit(1)
		}

		entries := make([]database.Entry, 0)
		for _, filePath := range filesPathNames {
			fileContent, err := workspace.ReadFile(filePath)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Can't read file - %v\n", err)
				os.Exit(1)
			}

			blob := database.Blob{}
			isExecutable := func() bool {
				fileInfo, err := workspace.GetFileState(filePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: Can't get file state - %v\n", err)
					os.Exit(1)
				}
				if fileInfo.Mode()&0111 != 0 {
					return true
				}
				return false
			}
			if err := blob.New(fileContent, filePath, isExecutable()); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Can't create new blob - %v\n", err)
				os.Exit(1)
			}

			entries = append(entries, &blob)
		}

		tree := database.Tree{}
		tree = tree.Build(entries)
		tree.Traverse(func(e database.Entry) {
			if err := db.Store(e); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Can't save entry in db - %v\n", err)
				os.Exit(1)
			}
		})

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

		commit := database.Commit{}
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

func init() {
	rootCmd.AddCommand(commitCmd)

	commitCmd.Flags().StringVarP(&message, "message", "m", "", "Use this given <msg> as the commit message.")
}
