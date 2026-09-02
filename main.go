package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/if1eight0sty/gitclean/internal/config"
	"github.com/if1eight0sty/gitclean/internal/git"
	"github.com/if1eight0sty/gitclean/internal/protected"
	"github.com/if1eight0sty/gitclean/internal/ui"
	"github.com/spf13/cobra"
)

var deleteFlag bool
var targetBranch string
var protectedBranchesFlag string
var saveConfig bool

func initDeleteFLag() {
	branchesCmd.Flags().BoolVarP(&deleteFlag, "delete", "d", false, "Interactively delete merged branches")
	branchesCmd.Flags().StringVarP(&targetBranch, "target", "t", "", "Target branch to check against (default: main/master)")
	branchesCmd.Flags().StringVarP(&protectedBranchesFlag, "protected", "p", "", "Comma-separated protected branches (default: main,master)")
	branchesCmd.Flags().BoolVar(&saveConfig, "save-config", false, "Save --protected branches to .gitclean config file")
}

var rootCmd = &cobra.Command{
	Use:   "gitclean",
	Short: "gitclean",
	Long:  `gitclean is a simple cli tool to help you manage your git branches.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var branchesCmd = &cobra.Command{
	Use:   "branches",
	Short: "List or delete merged branches",
	Long: `Lists branches that have been merged into the target branch.

Protected branches (won't be deleted):
  - Always: main, master
  - Add more via --protected flag: --protected develop,prod
  - Or create .gitclean file in repo: protected=develop,prod,int

Examples:
  gitclean branches                      # List merged branches
  gitclean branches --protected develop  # Add develop to protected list
  gitclean branches --target .           # Check against current branch
  gitclean branches --delete             # Interactive deletion`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if saveConfig {

			if err := config.SaveConfigFile(protectedBranchesFlag); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Println("Config saved to .gitclean")
			return nil
		}

		current, err := git.GetCurrentBranch()
		if err != nil {
			return fmt.Errorf("could not determine current branch: %w", err)
		}

		target := targetBranch
		if target == "." {
			target = current
		} else if target == "" {
			if git.BranchExists("main") {
				target = "main"
			} else if git.BranchExists("master") {
				target = "master"
			} else {
				target = current
			}
		}

		fmt.Printf("Working on branch: %s\n", current)
		fmt.Printf("Checking merged into: %s\n", target)
		fmt.Println()

		branches, err := git.GetMergedBranches(target)

		if err != nil {
			return err
		}

		protectedBranches := protected.ProtectedBranches(protectedBranchesFlag)

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

func interactiveDelete(branches []string) error {
	fmt.Println("Merged branches:")
	for index, branch := range branches {
		fmt.Printf("[%d] %s\n", index+1, branch)
	}
	fmt.Println()

	target := ui.InputPrompt("Delete from (local/remote/both): ")

	if target != "local" && target != "remote" && target != "both" {
		return fmt.Errorf("invalid choice: must be 'local' or 'remote' or 'both'")
	}

	selection := ui.InputPrompt("Select branches to delete (e.g., 1,3 or 'all'): ")

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

			toDelete = append(toDelete, branches[index-1])
		}
	}

	if len(toDelete) == 0 {
		fmt.Println("No branches selected.")
		return nil
	}

	confirm := ui.InputPrompt(fmt.Sprintf("\nDelete %d branches from %s (y/n): ", len(toDelete), target))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("Canceled.")
		return nil
	}

	for _, branch := range toDelete {
		var err error

		if target == "remote" || target == "both" {
			err = git.DeleteRemote(branch)
			if err != nil {
				fmt.Printf("Failed to delete %s from remote: %v\n", branch, err)
			} else {
				fmt.Printf("Deleted from remote: %s\n", branch)
			}
		}

		if target == "local" || target == "both" {
			err = git.DeleteLocal(branch)
			if err != nil {
				fmt.Printf("Failed to delete %s locally: %v\n", branch, err)
			} else {
				fmt.Printf("Deleted locally: %s\n", branch)
			}
		}

	}
	return nil
}

func main() {

	initDeleteFLag()

	rootCmd.AddCommand(branchesCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
