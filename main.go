package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/shlex"
)

func main() {
	port := flag.String("port", "18080", "upstream port")
	addr := flag.String("addr", "127.0.0.1:8080", "devserver bind address")
	liveReload := flag.Bool("live-reload", true, "enable/disable automatic reload via server sent events")
	buildCmd := flag.String("build-cmd", "make", "command to run to build the server")
	serverCmd := flag.String("server-cmd", "bin/server -addr {}",
		"command for running the server; use the {} placeholder for the host:port argument")
	webRoot := flag.String("web-root", "", "web root directory, reported file paths are relative to this directory")
	flag.Parse()

	target, err := url.Parse("http://127.0.0.1:" + *port)
	if err != nil {
		log.Fatalf("url parse error: %v", err)
	}

	restart := make(chan struct{})
	broadcaster := NewBroadcaster[fsEventBatch]()

	go rerun(target.Host, restart, *buildCmd, *serverCmd)
	go waitForEnter(restart)

	if *liveReload {
		go watchFiles([]string{".tmpl", ".html", ".css", ".js"}, func(b fsEventBatch) {
			b2 := make(fsEventBatch, 0, len(b))
			for _, e := range b {
				b2 = append(b2, webRootRel(*webRoot, e))
			}
			broadcaster.Broadcast(b2)
		})
	}

	runProxy(*addr, target, broadcaster)
}

func webRootRel(webRoot string, e fsEvent) fsEvent {
	if webRoot == "" {
		return e
	}
	if rel, err := filepath.Rel(webRoot, e.File); err == nil {
		e.File = "/" + rel
	} else {
		log.Printf("webRootRel: %v", err)
	}
	return e
}

// waitForEnter waits for a new line on os.Stdin. When a new line is received
// it sends a message on the ch channel.
func waitForEnter(ch chan<- struct{}) {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		ch <- struct{}{}
	}
}

// rerun builds and runs the server over and over again. A message on the
// restart channel initiates rebuild & restart.
func rerun(addr string, restart <-chan struct{}, buildCmd string, serverCmd string) {
	// build -> stop -> run
	run := func(stop func()) func() {
		ctx, cancel := context.WithCancel(context.Background())
		build(ctx, buildCmd)

		if stop != nil {
			stop()
		}

		done := startServer(ctx, addr, serverCmd)

		// TODO: move to waitForEnter, Enter might not restart the server
		fmt.Println("Hit Enter to rebuild and restart")

		return func() {
			cancel() // Stop the server

			select {
			case <-done: // Wait for server to stop
				infof("Stopped server")
			case <-time.After(10 * time.Second):
				log.Print("server stop timeout after 10 seconds")
			}
		}
	}

	stop := run(nil)
	for range restart {
		infof("Restarting...")
		stop = run(stop)
	}

	// Stop the server before exiting
	stop()
}

// Build the server binary using buildCmd.
func build(ctx context.Context, buildCmd string) {
	args, err := shlex.Split(buildCmd)
	if err != nil {
		log.Fatalf("build command parser error: %v", err)
	}

	start := time.Now()
	infof("Building...")
	fmt.Println(time.Now().Format(time.UnixDate))
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("build error: %s\n", err)
	}
	fmt.Printf("%s", out)

	infof("Build done; took %s", time.Since(start))
}

// Start the server using serverCmd. In serverCmd {} placeholder is replaced
// with the address in the form of host:port. When ctx is cancelled SIGTERM is
// sent to the server process.
func startServer(ctx context.Context, addr string, serverCmd string) <-chan struct{} {
	args, err := shlex.Split(serverCmd)
	if err != nil {
		log.Fatalf("server command parser error: %v", err)
	}

	for i, arg := range args {
		if arg == "{}" {
			args[i] = addr
			break
		}
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	// TODO: allow following logs
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	cmd.Cancel = func() error {
		return cmd.Process.Signal(syscall.SIGTERM)
	}

	done := make(chan struct{})

	go func() {
		if err := cmd.Run(); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Printf("server error: %s\n", err)
		}
		close(done)
	}()

	infof("Started server: %v", cmd)
	return done
}
