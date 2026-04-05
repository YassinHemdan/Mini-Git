/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	internals "JIT/internals"
	database "JIT/internals/database"
	"fmt"
	"os"
	"path/filepath"
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
		workspace.New(root_dir)
		for _, path := range args {

			fullpath := filepath.Join(root_dir, path)
			files, err := workspace.ListFiles(fullpath)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Could not list files - %v\n", err)
				os.Exit(1)
			}

			for _, fileName := range files {
				fmt.Printf("Adding file: %s\n", fileName)
				file_content, err := workspace.ReadFile(fileName)
				// fmt.Println(string(file_content))
				// fmt.Println("***************************************************")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: Could not read file - %v\n", err)
					os.Exit(1)
				}

				blob := database.Blob{}
				isExecutable := func() bool {
					fileInfo, err := workspace.GetFileState(fileName)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: Can't get file state - %v\n", err)
						os.Exit(1)
					}
					if fileInfo.Mode()&0111 != 0 {
						return true
					}
					return false
				}
				if err := blob.New(file_content, fileName, isExecutable()); err != nil {
					fmt.Fprintf(os.Stderr, "Error: Could not create a new blob - %v\n", err)
					os.Exit(1)
				}

				if err := db.Store(&blob); err != nil {
					fmt.Fprintf(os.Stderr, "Error: Could not store the blob - %v\n", err)
					os.Exit(1)
				}

				fileInfo, err := workspace.GetFileState(fileName)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: Can't get file state - %v\n", err)
					os.Exit(1)
				}
				stat, ok := fileInfo.Sys().(*syscall.Stat_t)
				if !ok {
					fmt.Fprintf(os.Stderr, "Error: Could not get file's stat - %v\n", err)
					os.Exit(1)
				}

				if err := index.Add(fileName, blob.GetOid(), stat); err != nil {
					fmt.Fprintf(os.Stderr, "Error: Could not add file - %v\n", err)
					os.Exit(1)
				}
			}

		}
		if err := index.WriteUpdates(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not add file - %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
