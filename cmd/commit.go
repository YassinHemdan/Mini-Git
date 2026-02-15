/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"Jit/internals"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	Run: func(cmd *cobra.Command, args []string) {

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
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
}
