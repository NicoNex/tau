package mod

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Where a module is fetched from, when that is not simply its path.
//
// An import path is an address, which is what makes the scheme need nothing
// central. It does not have to be the address of a *repository* though: a name
// like tau.dev/text can be a name its owner keeps, pointing at whatever forge
// holds the code this year. Without that, a project that moves host has to
// change every import that ever mentioned it, and the name of a library ends
// up belonging to a company rather than to whoever wrote it.
//
// The redirection is the one Go settled on, a meta tag served under the path:
//
//	<meta name="tau-import" content="tau.dev/text git https://github.com/x/text">
//
// A go-import tag is read as well, and the same way. A host already serving
// Go modules is serving the answer to the same question, and there is no
// reason to make somebody publish it twice.

// The forges whose paths are their clone URLs. They are asked first, so the
// usual import costs no request at all.
//
// ponytail: a prefix list. It is short because it only has to cover what is
// common; anything else falls through to the meta tag, which is the general
// answer.
var knownForges = []string{
	"github.com/",
	"gitlab.com/",
	"codeberg.org/",
	"bitbucket.org/",
	"git.sr.ht/",
	"gitea.com/",
}

var metaRe = regexp.MustCompile(
	`(?is)<meta[^>]+name=["'](tau|go)-import["'][^>]+content=["']([^"']+)["']`)

// altMetaRe is the same tag with the attributes the other way round, which is
// as valid HTML and as common in the wild.
var altMetaRe = regexp.MustCompile(
	`(?is)<meta[^>]+content=["']([^"']+)["'][^>]+name=["'](tau|go)-import["']`)

var vanityClient = &http.Client{Timeout: 20 * time.Second}

// RepoURL is where git is pointed at to fetch this module path.
func RepoURL(path string) (string, error) {
	prefix, url, err := lookupRepo(path)
	if err != nil {
		return "", err
	}
	_ = prefix
	return url, nil
}

// lookupRepo answers which prefix of an import path is the module, and where
// that module is cloned from.
func lookupRepo(path string) (prefix, url string, err error) {
	// The major suffix belongs to the module path, not to the repository: v2
	// of a library is a tag in the same place v1 was.
	_, base := PathMajor(path)

	for _, forge := range knownForges {
		if strings.HasPrefix(base, forge) {
			return base, "https://" + base, nil
		}
	}

	// Longest prefix first, so that a host serving several modules answers for
	// the one that was asked for and not for its own root.
	parts := strings.Split(base, "/")
	var firstErr error

	for n := len(parts); n >= 1; n-- {
		candidate := strings.Join(parts[:n], "/")

		p, u, err := discover(candidate)
		if err == nil {
			// The prefix it claims has to be one this path is inside of, or
			// the answer is about some other module.
			if p != candidate && !strings.HasPrefix(base, p+"/") {
				return "", "", fmt.Errorf("%s: the host answers for %q, which does not hold it", path, p)
			}
			return p, u, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	// No redirection anywhere: the path is taken to be the repository, which
	// is what it is on every forge that is not in the list above.
	if len(parts) >= 2 {
		return base, "https://" + base, nil
	}
	return "", "", fmt.Errorf("%s: no repository and no tau-import tag: %v", path, firstErr)
}

// discover asks a host where the module under this path lives.
func discover(path string) (prefix, url string, err error) {
	req, err := http.NewRequest("GET", "https://"+path+"?tau-get=1&go-get=1", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "tau-get/1")

	resp, err := vanityClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("%s: %s", path, resp.Status)
	}
	// A page, not a download: enough of it to hold a head, and no more.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", "", err
	}
	return parseMeta(string(body), path)
}

// parseMeta reads the import tags of a page and returns the one that answers
// for this path. A tau-import tag wins over a go-import one for the same
// prefix: a host that serves both means the two to differ.
func parseMeta(body, path string) (prefix, url string, err error) {
	type answer struct{ prefix, url string }
	var tau, gomod *answer

	take := func(kind, content string) {
		fields := strings.Fields(content)
		if len(fields) != 3 || fields[1] != "git" {
			return
		}
		a := &answer{fields[0], fields[2]}
		// Only a tag about this path, or about something holding it.
		if a.prefix != path && !strings.HasPrefix(path, a.prefix+"/") {
			return
		}
		if kind == "tau" {
			if tau == nil || len(a.prefix) > len(tau.prefix) {
				tau = a
			}
			return
		}
		if gomod == nil || len(a.prefix) > len(gomod.prefix) {
			gomod = a
		}
	}

	for _, m := range metaRe.FindAllStringSubmatch(body, -1) {
		take(m[1], m[2])
	}
	for _, m := range altMetaRe.FindAllStringSubmatch(body, -1) {
		take(m[2], m[1])
	}

	if tau != nil {
		return tau.prefix, tau.url, nil
	}
	if gomod != nil {
		return gomod.prefix, gomod.url, nil
	}
	return "", "", fmt.Errorf("%s: no tau-import tag", path)
}
