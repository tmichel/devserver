package main

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"
)

func TestSplitOnPlus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple split",
			input:    "hello+world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "Multiple splits",
			input:    "a+b+c+d",
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "No split character",
			input:    "hello world",
			expected: []string{"hello world"},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "Only split character",
			input:    "+",
			expected: []string{""},
		},
		{
			name:     "Split at beginning",
			input:    "+hello",
			expected: []string{"", "hello"},
		},
		{
			name:     "Split at end",
			input:    "hello+",
			expected: []string{"hello"},
		},
		{
			name:     "Multiple empty splits",
			input:    "+++",
			expected: []string{"", "", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := bufio.NewScanner(bytes.NewReader([]byte(tt.input)))
			scanner.Split(SplitOnPlus)

			var result []string
			for scanner.Scan() {
				result = append(result, scanner.Text())
			}

			if err := scanner.Err(); err != nil {
				t.Errorf("Scanner error: %v", err)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
