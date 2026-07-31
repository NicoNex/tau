package doc

import (
	"strings"
	"testing"
)

// TestHighlightSpans is what a colouring has to get right before it is worth
// having: every character of the source comes out once, in the right order,
// and a token is coloured whole.
func TestHighlightSpans(t *testing.T) {
	const src = `println("total:", len(keys(m)), 0644, true)
once = sync.Once()
s = ` + "`raw`" + `
`

	got := string(highlight(src))

	// The text is the source again once the markup is taken off.
	bare := strings.NewReplacer("&#34;", `"`, "&lt;", "<", "&gt;", ">", "&amp;", "&").Replace(got)
	for {
		i := strings.Index(bare, "<span")
		if i < 0 {
			break
		}
		j := strings.Index(bare[i:], ">")
		bare = bare[:i] + bare[i+j+1:]
	}
	bare = strings.ReplaceAll(bare, "</span>", "")

	if bare != src {
		t.Errorf("the source came back as\n%q\nwant\n%q", bare, src)
	}

	for _, want := range []string{
		`<span class="c-bi">println</span>`,           // a builtin
		`<span class="c-fn">Once</span>`,              // a call
		`<span class="c-str">&#34;total:&#34;</span>`, // both quotes inside the colour
		"<span class=\"c-str\">`raw`</span>",
		`<span class="c-num">0644</span>`,
		`<span class="c-lit">true</span>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("%s is not in\n%s", want, got)
		}
	}

	// A name that is not called stays plain.
	if strings.Contains(got, `<span class="c-fn">once</span>`) {
		t.Error("a name with no bracket after it was painted as a call")
	}
}
