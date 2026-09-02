package git

import (
	"os/exec"
	"strings"
)

func BranchExists(branch string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", branch)
	return cmd.Run() == nil
}

func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")

	currentBranchOutput, err := cmd.Output()
	if err != nil {
		return "", err
	}

	currentBranch := strings.TrimSpace(string(currentBranchOutput))

	return currentBranch, nil
}

func GetMergedBranches(target string) ([]string, error) {
	listCmd := exec.Command(
		"git",
		"--no-pager",
		"branch",
		"--format=%(refname:short)",
		"--merged",
		target,
	)

	branch, err := listCmd.Output()
	if err != nil {
		return nil, err
	}

	branches := strings.Split(strings.TrimSpace(string(branch)), "\n")

	var results []string
	for _, branch := range branches {
		if branch != "" {
			results = append(results, branch)
		}
	}

	return results, nil
}
