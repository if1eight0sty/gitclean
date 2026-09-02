package protected

import (
	"strings"

	"github.com/if1eight0sty/gitclean/internal/config"
)

func parseAndAddBranches(protected map[string]bool, branchString string) {
	branches := strings.Split(branchString, ",")
	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if branch != "" {
			protected[branch] = true
		}
	}
}

func ProtectedBranches(custom string) map[string]bool {
	protected := map[string]bool{
		"main":   true,
		"master": true,
	}

	if custom != "" {
		parseAndAddBranches(protected, custom)
		return protected
	}

	protectedConfig := config.ReadConfig()
	if protectedConfig != "" {
		parseAndAddBranches(protected, protectedConfig)
	}

	return protected
}
