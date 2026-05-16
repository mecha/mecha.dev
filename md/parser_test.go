package md

import (
	"strings"
	"testing"
)

func TestToHTMLHighlightsFencedCodeBlocks(t *testing.T) {
	html := string(ToHTML("```go main.go,readonly\npackage main\n\nfunc main() {}\n```"))

	if !strings.Contains(html, `<pre class="tok-chroma"><code>`) {
		t.Fatalf("expected highlighted code block wrapper, got %q", html)
	}

	if !strings.Contains(html, `class="tok-`) {
		t.Fatalf("expected syntax token classes in highlighted output, got %q", html)
	}

	if strings.Contains(html, `style=`) {
		t.Fatalf("expected class-based highlighting without inline styles, got %q", html)
	}
}
