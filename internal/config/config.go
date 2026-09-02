package config

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func getConfigPath() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	repoRoot := strings.TrimSpace(string(output))
	configPath := filepath.Join(repoRoot, ".gitclean")

	return configPath, nil
}

func SaveConfigFile(protected string) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	if configPath == "" {
		return nil
	}
	return os.WriteFile(configPath, []byte("protected="+protected), 0644)
}

func ReadConfig() string {

	configPath, configErr := getConfigPath()
	if configErr != nil {
		return ""
	}

	file, fileErr := os.Open(configPath)
	if fileErr != nil {
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
