package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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

		for (const {File: file, Events: events} of data.events) {
			const isCss = file.toLowerCase().endsWith(".css");
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
	if strings.HasPrefix(resp.Header.Get("content-type"), "text/html") {
		// TODO: use a streaming reader that modifies the content stream on the fly
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		i := bytes.Index(body, []byte("</body>"))
		if i > -1 {
			tail := append([]byte(reloadJs), body[i:]...)
			body = append(body[:i], tail...)
		}
		resp.Body = io.NopCloser(bytes.NewReader(body))
		// Let the reverse proxy figure this out
		resp.Header.Del("content-length")
	}

	return nil
}
