package cmd

import (
	"JIT/internals"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)


var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		// fetch cur dir
		dir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		path := strings.Join([]string{dir, internals.JitMetadataDir}, string(os.PathSeparator))

		err = os.Mkdir(path, internals.JitDefaultPermission)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		for _, content := range strings.Split(internals.JitMetadataContent, "|") {
			filePath := strings.Join([]string{path, content}, string(os.PathSeparator))
			err := os.Mkdir(filePath, internals.JitDefaultPermission)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

}
