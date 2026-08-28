package main

import (
	"fmt"
	"os"
	"os/exec"

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

// branch command
var branchsCmd = &cobra.Command{
	Use:   "branches",
	Short: "list branches",
	Long:  `Lists branches that have already been merged into the current branch. These branches can be safely reviewed and optionally deleted.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		branchCmd := exec.Command("git", "branch", "--merged")

		branchOut, branchErr := branchCmd.Output()

		if branchErr != nil {
			// fmt.Fprintln(os.Stderr, "error:", branchErr)
			return branchErr
		}

		fmt.Println(string(branchOut))

		return nil
	},
}

func main() {
	rootCmd.AddCommand(branchsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
