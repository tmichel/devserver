package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestInjectingReader_InjectBeforeBodyTag(t *testing.T) {
	body := `
<html lang="en">
	<head><title>Test</title></head>
	<body>
		<h1>Test</h1>
	</body>
</html>`

	inj := "<footer>injected</footer>"
	r := newInjectingReader(strings.NewReader(body), inj)

	assertInject(t, r, inj)
}

func TestInjectingReader_InjectLongDocument(t *testing.T) {
	b, err := os.ReadFile("testdata/long.html")
	if err != nil {
		t.Fatalf("unexpected error reading fixture: %v", err)
	}

	s := "<i>injected</i>"
	r := newInjectingReader(bytes.NewReader(b), s)

	assertInject(t, r, s)
}

func TestInjectingReader_InjectLongTagContent(t *testing.T) {
	f, err := os.Open("testdata/long-tag.html")
	if err != nil {
		t.Fatalf("unexpected error reading fixture: %v", err)
	}
	defer f.Close()

	s := "<i>injected</i>"
	r := newInjectingReader(f, s)

	res := assertInject(t, r, s)

	stat, _ := f.Stat()
	if int64(len(res)) < stat.Size() {
		t.Errorf("expected result to be longer than original html")
		t.Logf("result:\n%s", res)
	}
}

func TestInjectingReader_BufferTooSmall(t *testing.T) {
	r := strings.NewReader("<html><body>Content</body></html>")
	ir := newInjectingReader(r, "test")

	n, err := ir.Read(make([]byte, 5))
	if err == nil {
		t.Error("expected an error but gone none")
	}
	if want := 0; n != want {
		t.Errorf("unexpected value of n\nwant: %d\ngot:  %d", want, n)
	}
}

func TestInjectingReader(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		injectString string
		chunkSize    int
		want         string
	}{
		{
			name:         "Basic injection",
			input:        "<html><body>Content</body></html>",
			injectString: "<script>console.log('Injected!');</script>",
			chunkSize:    1024,
			want:         "<html><body>Content<script>console.log('Injected!');</script></body></html>",
		},
		{
			name:         "No body tag",
			input:        "<html><div>Content</div></html>",
			injectString: "<script>console.log('Injected!');</script>",
			chunkSize:    1024,
			want:         "<html><div>Content</div></html>",
		},
		{
			name:         "Multiple body tags",
			input:        "<html><body>Content1</body><body>Content2</body></html>",
			injectString: "<script>console.log('Injected!');</script>",
			chunkSize:    1024,
			want:         "<html><body>Content1<script>console.log('Injected!');</script></body><body>Content2</body></html>",
		},
		{
			name:         "Body tag split across chunks",
			input:        "<html><body>Content</bo" + "dy></html>",
			injectString: "<script>console.log('Injected!');</script>",
			chunkSize:    23,
			want:         "<html><body>Content<script>console.log('Injected!');</script></body></html>",
		},
		{
			name:         "Body tag at the end of a chunk",
			input:        "<html><body>Content</body" + "></html>",
			injectString: "<script>console.log('Injected!');</script>",
			chunkSize:    24,
			want:         "<html><body>Content<script>console.log('Injected!');</script></body></html>",
		},
		{
			name:         "Body tag split with only '>' in next chunk",
			input:        "<html><body>Content</body" + ">",
			injectString: "<script>console.log('Injected!');</script>",
			chunkSize:    24,
			want:         "<html><body>Content<script>console.log('Injected!');</script></body>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := newInjectingReader(strings.NewReader(tt.input), tt.injectString)

			var result bytes.Buffer
			buf := make([]byte, tt.chunkSize)
			for {
				n, err := ir.Read(buf)
				if err != nil && err != io.EOF {
					t.Fatalf("unexpected error: %v", err)
				}
				result.Write(buf[:n])
				if err == io.EOF {
					break
				}
			}

			got := result.String()
			if got != tt.want {
				t.Errorf("injectingReader produced incorrect output\nwant: %q\ngot:  %q", tt.want, got)
			}
		})
	}
}

func TestInjectingReader_LargeInput(t *testing.T) {
	// Create a large input that doesn't fit into a single Read call
	largeInput := strings.Repeat("a", 10000) + "<body>" + strings.Repeat("b", 10000) + "</body>" + strings.Repeat("c", 10000)
	injectString := "<script>console.log('Injected!');</script>"
	want := strings.Repeat("a", 10000) + "<body>" + strings.Repeat("b", 10000) + injectString + "</body>" + strings.Repeat("c", 10000)

	reader := strings.NewReader(largeInput)
	injector := newInjectingReader(reader, injectString)

	var result bytes.Buffer
	buf := make([]byte, 1024)
	for {
		n, err := injector.Read(buf)
		if err != nil && err != io.EOF {
			t.Fatalf("unexpected error: %v", err)
		}
		result.Write(buf[:n])
		if err == io.EOF {
			break
		}
	}

	got := result.String()
	if got != want {
		t.Errorf("large input test failed")
		t.Logf("want (first 100 chars): %q", want[:100])
		t.Logf("got (first 100 chars): %q", got[:100])
		t.Logf("want length: %d", len(want))
		t.Logf("got length: %d", len(got))
	}
}

func assertInject(t *testing.T, r io.Reader, injected string) []byte {
	t.Helper()

	res, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Contains(res, []byte(injected+"</body>")) {
		t.Errorf("expected to find injected content but got none\nbuf: %s", res)
	}

	return res
}
