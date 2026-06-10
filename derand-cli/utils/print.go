package utils

import "fmt"

func PrintTitle(args ...any) {
	a := append([]any{"•"}, args...)
	fmt.Println(a...)
}

func PrintSubtitle(args ...any) {
	a := append([]any{"   →"}, args...)
	fmt.Println(a...)
}

func Bold(s string) string {
	return "\033[1m" + s + "\033[0m"
}

func Red(s string) string {
	return "\033[31m" + s + "\033[0m"
}

func Green(s string) string {
	return "\033[32m" + s + "\033[0m"
}

func Yellow(s string) string {
	return "\033[33m" + s + "\033[0m"
}
