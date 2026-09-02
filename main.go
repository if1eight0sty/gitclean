package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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

		if saveConfig {
			if err := saveConfigFile(protectedBranchesFlag); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Println("Config saved to .gitclean")
			return nil
		}

		current, err := getCurrentBranch()
		if err != nil {
			return fmt.Errorf("could not determine current branch: %w", err)
		}

		target := targetBranch
		if target == "." {
			target = current
		} else if target == "" {
			if branchExists("main") {
				target = "main"
			} else if branchExists("master") {
				target = "master"
			} else {
				target = current
			}
		}

		fmt.Printf("Working on branch: %s\n", current)
		fmt.Printf("Checking merged into: %s\n", target)
		fmt.Println()

		branches, err := getMergedBranches(target)

		if err != nil {
			return err
		}

		protectedBranches := buildProtectedBranches(protectedBranchesFlag)

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

	fmt.Print("Delete from (local/remote/both): ")
	reader := bufio.NewReader(os.Stdin)
	target, _ := reader.ReadString('\n')
	target = strings.TrimSpace(strings.ToLower(target))

	if target != "local" && target != "remote" {
		return fmt.Errorf("invalid choice: must be 'local' or 'remote' or 'both'")
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

			toDelete = append(toDelete, branches[index-1])
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
		var err error

		if target == "local" || target == "both" {
			err = deleteLocalBranch(branch)
			if err != nil {
				fmt.Printf("Failed to delete %s locally: %v\n", branch, err)
			} else {
				fmt.Printf("Deleted locally: %s\n", branch)
			}
		}

		if target == "remote" || target == "both" {
			err = deleteRemoteBranch(branch)
			if err != nil {
				fmt.Printf("Failed to delete %s from remote: %v\n", branch, err)
			} else {
				fmt.Printf("Deleted from remote: %s\n", branch)
			}
		}
	}
	return nil
}

func deleteLocalBranch(branch string) error {
	// checkCmd := exec.Command("git", "ls-remote", "--heads", "origin", branch)
	// output, err := checkCmd.CombinedOutput()

	// if err != nil || len(output) == 0 {
	// 	return fmt.Errorf("branch %s does not exists on remote", branch)
	// }

	if !branchExists(branch) {
		return fmt.Errorf("branch %s does not exists locally", branch)
	}

	cmd := exec.Command("git", "branch", "-d", branch)
	return cmd.Run()
}

func deleteRemoteBranch(branch string) error {
	done := showSpinner(fmt.Sprintf("Deleting %s from remote...", branch))
	defer done()

	cmd := exec.Command("git", "push", "origin", "--delete", branch)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	return nil
}

func showSpinner(message string) func() {
	stop := make(chan bool)
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	go func() {
		i := 0
		for {
			select {
			case <-stop:
				fmt.Printf("\r%s Done!          \n", message)
				return
			default:
				fmt.Printf("\r%s %s", spinner[i%len(spinner)], message)
				time.Sleep(100 * time.Millisecond)
				i++
			}
		}
	}()
	return func() {
		stop <- true
	}
}

func readConfig() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	repoRoot := strings.TrimSpace(string(output))
	configPath := filepath.Join(repoRoot, ".gitclean")
	file, err := os.Open(configPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "protected=") {
			return strings.TrimPrefix(line, "protected=")
		}
	}
	return ""
}
func saveConfigFile(protected string) error {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	repoRoot := strings.TrimSpace(string(output))
	configPath := filepath.Join(repoRoot, ".gitclean")

	return os.WriteFile(configPath, []byte("protected="+protected), 0644)
}

func buildProtectedBranches(custom string) map[string]bool {
	protected := map[string]bool{
		"main":   true,
		"master": true,
	}
	if custom != "" {
		branches := strings.Split(custom, ",")
		for _, branch := range branches {
			branch = strings.TrimSpace(branch)
			if branch != "" {
				protected[branch] = true
			}
		}
		return protected
	}

	configProtected := readConfig()
	if configProtected != "" {
		branches := strings.Split(configProtected, ",")
		for _, branch := range branches {
			branch = strings.TrimSpace(branch)
			if branch != "" {
				protected[branch] = true
			}
		}
	}

	return protected
}

func main() {

	initDeleteFLag()

	rootCmd.AddCommand(branchesCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
