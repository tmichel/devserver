package main

import (
	"strings"

	"github.com/fatih/color"
)

var bold = color.New(color.FgHiWhite, color.Bold)

func infof(format string, args ...any) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	bold.Printf("===> "+format, args...)
}
