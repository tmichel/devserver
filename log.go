package main

import "fmt"

type ansi string

const (
	escape ansi = "\x1b"
	reset       = escape + "[0m"
	bold        = escape + "[1m"
)

func ansiColor(str string, mod ansi) string {
	return fmt.Sprintf("%s%s%s", mod, str, reset)
}

func termf(mod ansi, format string, args ...any) string {
	return fmt.Sprintf(string(mod)+format+string(reset), args...)
}

func infof(format string, args ...any) {
	fmt.Println(termf(bold, "===> "+format, args...))
}
