/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	internals "JIT/internals"
	database "JIT/internals/database"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("add called")

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

		// lets join the jit path with "index"
		// .jit/index
		index, err := internals.NewIndex(jit_dir + string(os.PathSeparator) + "index")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Can't create an index file - %v\n", err)
			os.Exit(1)
		}
		workspace := internals.Workspace{}
		for _, path := range args {
			fmt.Printf("Adding file: %s\n", path)

			file_content, err := workspace.ReadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Could not read file - %v\n", err)
				os.Exit(1)
			}

			blob := database.Blob{}

			isExecutable := func() bool {
				fileInfo, err := workspace.GetFileState(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: Can't get file state - %v\n", err)
					os.Exit(1)
				}
				if fileInfo.Mode()&0111 != 0 {
					return true
				}
				return false
			}
			if err := blob.New(file_content, path, isExecutable()); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Could not create a new blob - %v\n", err)
				os.Exit(1)
			}

			if err := db.Store(&blob); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Could not store the blob - %v\n", err)
				os.Exit(1)
			}

			fileInfo, err := workspace.GetFileState(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Can't get file state - %v\n", err)
				os.Exit(1)
			}
			stat, ok := fileInfo.Sys().(*syscall.Stat_t)
			if !ok {
				fmt.Fprintf(os.Stderr, "Error: Could not get file's stat - %v\n", err)
				os.Exit(1)
			}

			if err := index.Add(path, blob.GetOid(), stat); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Could not add file - %v\n", err)
				os.Exit(1)
			}

			if err := index.WriteUpdates(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Could not add file - %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
