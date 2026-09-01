package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var deleteFlag bool
var targetBranch string

func initDeleteFLag() {
	branchesCmd.Flags().BoolVarP(&deleteFlag, "delete", "d", false, "Interactively delete merged branches")
	branchesCmd.Flags().StringVarP(&targetBranch, "target", "t", "", "Target branch to check against (default: main/master)")
}

var rootCmd = &cobra.Command{
	Use:   "branches [branch1] [branch2]...",
	Short: "List or delete merged branches",
	Long: `Lists branches that have been merged into the target branch (default: mai
n/master).
These branches can be safely reviewed and optionally deleted.

Examples:
  gitclean branches                    # List merged branches (default: main)
  gitclean branches --target develop   # Check against develop branch
  gitclean branches --delete           # Interactive deletion
  gitclean branches --delete feature/login feature/signup  # Delete specific branc
hes`,
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

		target := targetBranch
		if target == "" {
			if branchExists("main") {
				target = "main"
			} else if branchExists("master") {
				target = "master"
			} else {
				target = current
			}
		}

		fmt.Printf("Working on branch: %s\n", current)
		fmt.Printf("Checking merged into: %s\n\n", target)
		fmt.Println()

		branches, err := getMergedBranches(target)

		if err != nil {
			return err
		}

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
			deletable = append(deletable, branch)
		}

		if len(deletable) == 0 {
			fmt.Println("No merged branches for delete.")
			return nil
		}

		if !deleteFlag {
			fmt.Println("Merged branches (Safe for delete):")

			for _, branch := range deletable {
				fmt.Printf(" - %s\n", branch)
			}

			fmt.Println("\n Run with --delete to remove branches")
			return nil
		}

		return interactiveDelete(deletable)
	},
}

func branchExists(branch string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", branch)
	return cmd.Run() == nil

}

func getMergedBranches(target string) ([]string, error) {
	listBranchCmd := exec.Command(
		"git",
		"--no-pager",
		"branch",
		"--format=%(refname:short)",
		"--merged",
		target,
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

func interactiveDelete(branches []string) error {
	fmt.Println("Merged branches:")
	for index, branch := range branches {
		fmt.Printf("[%d] %s\n", index+1, branch)
	}
	fmt.Println()

	fmt.Print("Delete from (local/remote): ")
	reader := bufio.NewReader(os.Stdin)
	target, _ := reader.ReadString('\n')
	target = strings.TrimSpace(strings.ToLower(target))

	if target != "local" && target != "remote" {
		return fmt.Errorf("invalid choice: must be 'local' or 'remote'")
	}

	fmt.Print("Select branches to delete (e.g., 1,3 or 'all'): ")
	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(strings.ToLower(selection))

	var toDelete []string

	if selection == "all" {
		toDelete = branches
	} else {
		numbers := strings.Split(selection, ",")

		for _, number := range numbers {
			number = strings.TrimSpace(number)
			index := 0

			_, err := fmt.Sscanf(number, "%d", &index)

			if err != nil || index < 1 || index > len(branches) {
				return fmt.Errorf("invalid selection: %s", number)
			}

			toDelete = append(toDelete, number)
		}
	}

	if len(toDelete) == 0 {
		fmt.Println("No branches selected.")
		return nil
	}

	fmt.Printf("\nDelete %d branches from %s (y/n): ", len(toDelete), target)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("Canceled.")
		return nil
	}

	for _, branch := range toDelete {
		var error error

		if target == "local" {
			deleteLocalBranch(branch)
		} else {

		}
		if error != nil {
			fmt.Printf("Fail to delete %s: %v\n", branch, error)
		} else {
			fmt.Printf("Deleted: %s\n", branch)
		}
	}
	return nil
}

func deleteLocalBranch(branch string) error {
	cmd := exec.Command("git", "branch", "-d", branch)
	return cmd.Run()
}

func deleteRemoteBranch(branch string) error {
	cmd := exec.Command("git", "push", "origin", "--delete", branch)
	return cmd.Run()
}

func main() {

	initDeleteFLag()

	rootCmd.AddCommand(branchesCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
