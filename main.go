package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var deleteFlag bool

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
var branchesCmd = &cobra.Command{
	Use:   "branches",
	Short: "list branches",
	Long:  `Lists branches that have already been merged into the current branch. These branches can be safely reviewed and optionally deleted.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		current, err := getCurrentBranch()
		if err != nil {
			return fmt.Errorf("could not determine current branch: %w", err)
		}

		fmt.Printf("Working on branch: %s\n", current)

		branches, err := getMergedBranches()

		if err != nil {
			return err
		}
		// Define which branches are off-limits
		protectedBranches := map[string]bool{
			"main":    true,
			"master":  true,
			"develop": true,
		}

		var deletable []string
		for _, branch := range branches {
			if branch == current || protectedBranches[branch] {
				continue
			}
			fmt.Println(branch)
			deletable = append(deletable, branch)
		}

		return nil
	},
}

func getMergedBranches() ([]string, error) {
	listBranchCmd := exec.Command(
		"git",
		"--no-pager",
		"branch",
		"--format=%(refname:short)",
		"--merged",
	)

	branchOut, err := listBranchCmd.Output()
	if err != nil {
		return nil, err
	}
	var results []string

	branches := strings.Split(strings.TrimSpace(string(branchOut)), "\n")

	for _, branch := range branches {
		if branch != "" {
			results = append(results, branch)
		}
	}

	return results, nil
}

func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")

	currentBranchOutput, err := cmd.Output()

	if err != nil {
		return "", err
	}

	currentBranch := strings.TrimSpace((string(currentBranchOutput)))

	return currentBranch, nil
}

func main() {

	branchesCmd.Flags().BoolVarP(&deleteFlag, "delete", "d", false, "Delete the merged branches")

	rootCmd.AddCommand(branchesCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
