package main

import (
	"reflect"
	"testing"
)

func TestPrepareCommand(t *testing.T) {
	testCases := []struct {
		serverCmd string
		addr      string
		args      []string
		wantError bool
	}{
		{"bin/server -addr {}", "localhost:8888", []string{"bin/server", "-addr", "localhost:8888"}, false},
		{"bin/server -port {port}", "localhost:8888", []string{"bin/server", "-port", "8888"}, false},
		{"bin/server -host {host}", "localhost:8888", []string{"bin/server", "-host", "localhost"}, false},
		{"bin/server -host {host} -port {port}", "localhost:8888", []string{"bin/server", "-host", "localhost", "-port", "8888"}, false},
		{"bin/server -host {host} -port {port}", "localhost", nil, true},
	}

	for _, tt := range testCases {
		t.Run(tt.serverCmd, func(t *testing.T) {
			result, err := prepareCommand(tt.serverCmd, tt.addr)
			checkError(t, err, tt.wantError)

			if !reflect.DeepEqual(result, tt.args) {
				t.Errorf("\nargs=%v\ngot: %v", tt.args, result)
			}
		})
	}
}

func checkError(t *testing.T, err error, wantError bool) {
	t.Helper()
	if wantError && err == nil {
		t.Error("expected an error but got none")
	} else if !wantError && err != nil {
		t.Errorf("not expected any errors but got %v", err)
	}
}
