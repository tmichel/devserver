package main

import (
	"bufio"
	"iter"
	"strings"
)

// Lines takes a string and returns an iterator, where each string is a line
// from the input.
func Lines(s string) iter.Seq[string] {
	sc := bufio.NewScanner(strings.NewReader(s))
	return func(yield func(string) bool) {
		for sc.Scan() {
			if !yield(sc.Text()) {
				return
			}
		}
	}
}
