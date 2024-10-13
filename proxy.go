package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"
	"strings"
	"time"
)

func runProxy(addr string, target *url.URL, fileWatch *Broadcaster[fsEventBatch]) {
	rp := httputil.NewSingleHostReverseProxy(target)
	rp.ModifyResponse = injectScript

	mux := http.NewServeMux()
	mux.Handle("/", rp)
	mux.Handle("/_dev", &watchHandler{fileWatch})

	srv := http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 1 * time.Minute,
		IdleTimeout:       1 * time.Minute,
		MaxHeaderBytes:    8 * (1 << 10), // 8K
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("proxy: error running server: %v", err)
	}
}

const reloadJs = `
<script type="module" defer>
	const es = new EventSource("/_dev");
	es.addEventListener("change", (e) => {
		const data = JSON.parse(e.data)
		console.info("change event", data);

		for (const {File: file, Ext: ext, Events: events} of data.events) {
			const isCss = ext === ".css";
			const isUpdated = events.includes("Updated");

			if (isCss && isUpdated) {
				for (const link of document.getElementsByTagName("link")) {
					const url = new URL(link.href)

					if (url.host === location.host && url.pathname === file) {
						const next = link.cloneNode();
						next.href = file + '?' + Math.random().toString(36).slice(2);
						next.onload = () => link.remove();
						link.parentNode.insertBefore(next, link.nextSibling);
						console.info("replaced css", { old: link, new: next });
						return
					}
				}
			}
		}

		console.info("relading due to file change")
		window.location.reload();
	});
</script>`

func injectScript(resp *http.Response) error {
	if !strings.HasPrefix(resp.Header.Get("content-type"), "text/html") {
		return nil
	}

	// Let the reverse proxy figure out the Content-Length
	resp.Header.Del("content-length")
	resp.Body = newInjectingReader(resp.Body, reloadJs)
	return nil
}

const bodyEndMarker = "</body>"

var _ io.ReadCloser = (*injectingReader)(nil)

type injectingReader struct {
	buf     []byte
	wrapped io.Reader
	content []byte

	seenMarker bool
	atEOF      bool
}

func newInjectingReader(r io.Reader, content string) *injectingReader {
	return &injectingReader{
		wrapped: r,
		content: []byte(content),
	}
}

func (ir *injectingReader) Read(p []byte) (n int, err error) {
	if len(p) < len(bodyEndMarker) {
		return 0, fmt.Errorf("inject: buffer is too small, len(p) >= %d must be true", len(bodyEndMarker))
	}

	// flush buffer
	if ir.atEOF {
		n = copy(p, ir.buf)
		if n < len(ir.buf) {
			ir.buf = ir.buf[n:]
			return n, nil
		}
		return n, io.EOF
	}

	tmp := make([]byte, len(ir.buf)+len(p))
	copy(tmp, ir.buf)
	n, err = ir.wrapped.Read(tmp[len(ir.buf):])

	// Return non-EOF error immediately
	ir.atEOF = errors.Is(err, io.EOF)
	if err != nil && !ir.atEOF {
		return 0, err
	}

	// Set boundary based on read bytes
	// tmp should contain whatever was in ir.buf and what we read just now
	tmp = tmp[:n+len(ir.buf)]

	var searchBuf []byte
	if !ir.seenMarker {
		if i := bytes.Index(tmp, []byte(bodyEndMarker)); i != -1 {
			// Marker found
			tmp = slices.Concat(tmp[:i], ir.content, tmp[i:])
			ir.seenMarker = true
		} else if len(tmp) >= len(bodyEndMarker) && bytes.IndexByte(tmp[len(tmp)-len(bodyEndMarker)-1:], '<') != -1 {
			// Possibly we have tag in the last few bytes
			searchBuf = tmp[len(tmp)-len(bodyEndMarker)-1:]
			tmp = tmp[:len(tmp)-len(bodyEndMarker)-1]
		}
	}

	n = copy(p, tmp)
	ir.buf = append(searchBuf, tmp[n:]...)
	return n, nil
}

func (ir *injectingReader) Close() error {
	if rc, ok := ir.wrapped.(io.Closer); ok {
		return rc.Close()
	}
	return nil
}
