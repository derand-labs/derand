package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// Confirm requires the answer must be explicitly "y".
func Confirm(prompt string, args ...any) bool {
	fmt.Printf(prompt, args...)
	reader := bufio.NewReader(os.Stdin)

	answer, err := reader.ReadString('\n')
	if err != nil {
		// This should not happen
		panic(err)
	}

	return strings.TrimSpace(answer) == "y"
}

func AskPassword(prompt string, args ...any) string {
	fmt.Printf(prompt, args...)
	passwordB, err := term.ReadPassword(syscall.Stdin)
	fmt.Println()
	if err != nil {
		// This should not happen
		panic(err)
	}
	return string(passwordB)
}
