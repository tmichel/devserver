package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var bold = color.New(color.FgHiWhite, color.Bold)

func infof(format string, args ...any) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}

	// NOTE: use Sprintf to create the color formatting string and use
	// fmt.Print to write it to STDOUT in one go. color.Printf uses separate
	// writes for setting the color, writing the string, and resetting the
	// color. Separate writes to STDOUT allows concurrent writes to interleave
	// with each other which can result in text being formatted which should
	// not be formatted.
	fmt.Print(bold.Sprintf("===> "+format, args...))
}
