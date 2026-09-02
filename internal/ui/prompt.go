package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func InputPrompt(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	input = strings.TrimSpace(strings.ToLower(input))

	return input
}
