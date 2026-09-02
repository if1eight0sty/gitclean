package git

import (
	"fmt"
	"os/exec"

	"github.com/if1eight0sty/gitclean/internal/ui"
)

func DeleteLocal(branch string) error {
	if !BranchExists(branch) {
		return fmt.Errorf("Branch %s does not exists locally", branch)
	}
	cmd := exec.Command("git", "branch", "-d", branch)
	return cmd.Run()
}

func DeleteRemote(branch string) error {
	done := ui.ShowSpinner(fmt.Sprintf("Deleting %s from remote...", branch))
	defer done()

	cmd := exec.Command("git", "push", "origin", "--delete", branch)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	return nil
}
