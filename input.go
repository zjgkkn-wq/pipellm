package main

import (
	"bufio"
	"os"
	"strings"
)

func ReadStdin() string {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// Terminal mode - no piped input
		return ""
	}

	var input strings.Builder
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input.WriteString(scanner.Text() + "\n")
	}
	return strings.TrimSpace(input.String())
}
