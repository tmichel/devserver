package main

import (
	"iter"
	"reflect"
	"testing"
)

// Helper function to collect all lines from the iterator

func TestLines(t *testing.T) {
	collectLines := func(seq iter.Seq[string]) []string {
		var lines []string
		for line := range seq {
			lines = append(lines, line)
		}
		return lines
	}
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "Single line",
			input:    "Hello, World!",
			expected: []string{"Hello, World!"},
		},
		{
			name:     "Multiple lines",
			input:    "Line 1\nLine 2\nLine 3",
			expected: []string{"Line 1", "Line 2", "Line 3"},
		},
		{
			name:     "Lines with empty lines",
			input:    "Line 1\n\nLine 3\n",
			expected: []string{"Line 1", "", "Line 3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seq := Lines(tt.input)
			result := collectLines(seq)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Lines() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLinesEarlyTermination(t *testing.T) {
	input := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	seq := Lines(input)

	var result []string
	seq(func(line string) bool {
		result = append(result, line)
		return len(result) < 3 // Stop after collecting 3 lines
	})

	expected := []string{"Line 1", "Line 2", "Line 3"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Early termination test failed. Got %v, want %v", result, expected)
	}
}
