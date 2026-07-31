package doc

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// callbackOnly are the names that live on an object nobody returns: it is
// built inside a function and handed to something the caller wrote, so no
// chain of returns leads to it and nothing static can find it. They are
// listed rather than skipped so that the list stays small and honest.
var callbackOnly = map[string]bool{
	// The response writer and the request a handler is called with.
	"Write": true, "WriteHeader": true, "WriteString": true, "RemoteAddr": true,
	// The record a flag set keeps for one flag.
	"Kind": true, "Value": true, "Default": true,
}

// TestCoverage counts the exported assignments the source holds and the ones
// that were found, module by module: a name the reader can write is a name
// this has to document.
func TestCoverage(t *testing.T) {
	re := regexp.MustCompile(`(?m)^\t*(?:[a-zA-Z_][a-zA-Z0-9_]*\.)?([A-Z][a-zA-Z0-9_]*) = `)

	filepath.Walk("../../stdlib", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".tau") || strings.HasSuffix(path, "_test.tau") {
			return nil
		}

		src, _ := os.ReadFile(path)
		want := map[string]bool{}
		for _, m := range re.FindAllStringSubmatch(string(src), -1) {
			want[m[1]] = true
		}

		p, err := Load(filepath.Base(path), path)
		if err != nil {
			t.Errorf("%s: %v", path, err)
			return nil
		}

		got := map[string]bool{}
		var walk func([]Entry)
		walk = func(es []Entry) {
			for _, e := range es {
				got[e.Name] = true
				walk(e.Children)
			}
		}
		walk(p.Entries)

		for name := range want {
			if !got[name] && !callbackOnly[name] {
				t.Errorf("%s: %s is in the source but not in the docs", path, name)
			}
		}
		return nil
	})
}
