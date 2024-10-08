package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path"
	"slices"
	"strings"
	"time"
)

func watchFiles(exts []string, f func(fsEventBatch)) {
	args := []string{
		"--batch-marker=+",
		"--no-defer",
		"--event-flags",
	}

	// Watch the current directory
	args = append(args, ".")

	cmd := exec.Command("fswatch", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("watch error: %v\n", err)
		return
	}

	infof("Watching files: %v", exts)
	go func() {
		if err := cmd.Run(); err != nil {
			fmt.Printf("watch error: %v\n", err)
		}
	}()

	s := bufio.NewScanner(stdout)
	s.Split(SplitOnPlus)
	for s.Scan() {
		var batch fsEventBatch
		for line := range Lines(strings.TrimSpace(s.Text())) {
			event := parseEvent(line)
			if slices.Contains(exts, path.Ext(event.File)) {
				batch = append(batch, event)
			}
		}

		if len(batch) > 0 {
			f(batch)
		}
	}
}

type fsEventBatch []fsEvent

type fsEvent struct {
	File   string
	Events []string
	t      time.Time
}

func (e fsEvent) is(name string) bool {
	return slices.Contains(e.Events, name)
}

func parseEvent(s string) fsEvent {
	parts := strings.Split(s, " ")
	return fsEvent{
		File:   parts[0],
		Events: parts[1:],
		t:      time.Now(),
	}
}

type watchHandler struct {
	bc *Broadcaster[fsEventBatch]
}

func (h *watchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	c := http.NewResponseController(w)

	w.Header().Set("content-type", "text/event-stream")
	w.Header().Set("cache-control", "no-cache")
	w.WriteHeader(http.StatusOK)

	c.Flush()

	ctx := r.Context()
	ch, remove := h.bc.AddListener()
	for {
		select {
		case <-ctx.Done():
			remove()
			return
		case events := <-ch:
			data, err := json.Marshal(map[string]any{
				"events": events,
				"time":   time.Now(),
			})
			if err != nil {
				log.Printf("watchHandler: json encode error: %v", err)
				continue
			}

			fmt.Fprintln(w, "event: change")
			fmt.Fprintf(w, "data: %s\n\n", data)
			c.Flush()
		}
	}
}
