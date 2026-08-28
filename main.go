package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitclean",
	Short: "A Git maintenance tool",
	Long:  `gitclean is a Git maintenance tool that helps developers clean up merged branches and generate changelogs from Git history.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("this is the root command")
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "there is an error", err)
		os.Exit(0)
	}
}
