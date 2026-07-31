package mod

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The discovery talks to a real server, over the real client, so that what is
// tested is the request that goes out and not only the parsing of a string.
func TestDiscoverOverHTTP(t *testing.T) {
	var asked string

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asked = r.URL.Path + "?" + r.URL.RawQuery
		fmt.Fprintf(w, `<html><head><meta name="tau-import"
			content="%s git https://github.com/x/text"></head></html>`,
			strings.TrimPrefix(r.Host+r.URL.Path, ""))
	}))
	defer srv.Close()

	old := vanityClient
	vanityClient = srv.Client()
	defer func() { vanityClient = old }()

	host := strings.TrimPrefix(srv.URL, "https://")
	prefix, url, err := discover(host + "/text")
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://github.com/x/text" {
		t.Errorf("url = %q", url)
	}
	if prefix != host+"/text" {
		t.Errorf("prefix = %q", prefix)
	}
	// Both queries go out, so a host that only knows one of them still answers.
	if !strings.Contains(asked, "tau-get=1") || !strings.Contains(asked, "go-get=1") {
		t.Errorf("asked for %q", asked)
	}
}
