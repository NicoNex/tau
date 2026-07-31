package doc

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// HTML writes the whole package as one page: everything is on it, however
// deep, and the name asked for is only where the page opens.
//
// The page carries its own style and its own script. Nothing is fetched, so
// it reads the same with no network, which is what a page written to a
// temporary file and opened straight away has to do.
func HTML(w io.Writer, p Package) error {
	return page.Execute(w, view{
		Package: p,
		Nav:     nav(p.Entries, ""),
		Blocks:  blocks(p.Doc),
		Body:    sections(p, p.Entries, "", 0),
	})
}

// Write puts the page in the cache directory under the module's name and
// returns the file.
//
// It is written afresh on every call, so a module that changed says so the
// next time it is asked about: nothing here is ever reused, and the directory
// is where the file is put rather than something consulted before writing it.
// Reading a module and writing its page takes about ten milliseconds, which
// is not worth the trouble of deciding whether it was still good.
//
// The same module lands on the same file every time, so a browser left open
// on one reloads onto the new page rather than filling up with copies. The
// file outlives the command on purpose: a browser goes on needing it after
// the tab is open, to reload, to follow an anchor, to go back. Deleting it
// would be a race with the reader. It is a few tens of kilobytes in the
// directory the system keeps for exactly this, and throwing that directory
// away costs nothing at any time.
func Write(p Package) (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	dir = filepath.Join(dir, "tau", "doc")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	// The listings first: the page links to them, and a link that lands on
	// nothing is worse than no link.
	for _, src := range p.Files {
		if err := writeSource(dir, p, src); err != nil {
			return "", err
		}
	}

	path := filepath.Join(dir, pageName(p)+".html")

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := HTML(f, p); err != nil {
		return "", err
	}
	return path, nil
}

// OpenBrowser writes the page and hands it to whatever opens files here, at
// the anchor of the name asked for.
func OpenBrowser(p Package, sym string) error {
	path, err := Write(p)
	if err != nil {
		return err
	}

	url := "file://" + filepath.ToSlash(path)
	if sym != "" {
		url += "#" + sym
	}

	fmt.Fprintln(os.Stderr, url)
	return open(url)
}

// open asks the system to open a URL, with the command each one has for it.
func open(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("tau doc: cannot open a browser: %w (the page is at %s)", err, url)
	}
	// Nothing waits for it: a browser lives longer than the command that
	// started it, and the page is already on disk.
	go cmd.Wait()
	return nil
}

// view is what the doc template is given.
type view struct {
	Package Package
	Nav     []navItem
	Blocks  []block
	Body    []section
}

// listing is what the source template is given: one file, a line at a time.
type listing struct {
	Package Package
	Name    string // the file, without the directories in front of it
	Path    string // and with them
	Doc     string // the page to go back to
	Lines   []line
}

type line struct {
	N    int
	HTML template.HTML
}

// pageName is the file a module's documentation is written to. A path has
// slashes in it and a file name cannot, so they become dashes.
func pageName(p Package) string {
	return strings.NewReplacer("/", "-", string(filepath.Separator), "-").Replace(p.Path)
}

// srcName is the file the listing of one source is written to, next to the
// page that links to it.
func srcName(p Package, src string) string {
	if src == "" {
		return ""
	}
	return pageName(p) + "." + filepath.Base(src) + ".html"
}

// writeSource writes the listing of one file and returns its name.
func writeSource(dir string, p Package, src string) error {
	text, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	lines := highlightLines(string(text))
	// A file ends in a newline, which is the end of the last line and not a
	// line of its own.
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}

	rows := make([]line, len(lines))
	for i, l := range lines {
		rows[i] = line{N: i + 1, HTML: l}
	}

	f, err := os.Create(filepath.Join(dir, srcName(p, src)))
	if err != nil {
		return err
	}
	defer f.Close()

	return source.Execute(f, listing{
		Package: p,
		Name:    filepath.Base(src),
		Path:    src,
		Doc:     pageName(p) + ".html",
		Lines:   rows,
	})
}

// navItem is one line of the index down the side.
type navItem struct {
	ID    string
	Name  string
	Kind  string
	Depth int
	Top   bool
}

// section is one name on the page together with the names under it: what a
// constructor builds is written inside the block of the constructor, since
// that is where it belongs and where a reader looks for it.
//
// The declaration is kept in pieces rather than as one string, so that the
// page can colour the name apart from what it holds.
type section struct {
	ID     string
	Name   string
	Sig    string
	Val    string
	Kind   string
	Under  string // the name this hangs off, empty at the top level
	Src    string // the line of the listing this was read from
	Blocks []block
	Kids   []section
	Depth  int
}

// block is a piece of a comment: a paragraph, or the lines of an example.
// Telling one from the other is what makes an indented example on the page
// what it is in the source. An example carries its colouring with it, done
// once here rather than on every page load in the browser.
type block struct {
	Text string
	HTML template.HTML
	Code bool
}

func nav(entries []Entry, prefix string) []navItem {
	var out []navItem

	for _, e := range entries {
		id := e.Name
		if prefix != "" {
			id = prefix + "." + e.Name
		}
		kind := "value"
		if e.Sig != "" {
			kind = "fn"
		}
		depth := strings.Count(id, ".")
		out = append(out, navItem{ID: id, Name: e.Name, Kind: kind, Depth: depth, Top: depth == 0})
		out = append(out, nav(e.Children, id)...)
	}
	return out
}

func sections(p Package, entries []Entry, prefix string, depth int) []section {
	var out []section

	for _, e := range entries {
		id := e.Name
		if prefix != "" {
			id = prefix + "." + e.Name
		}

		kind := "value"
		if e.Sig != "" {
			kind = "fn"
		}

		out = append(out, section{
			ID:     id,
			Name:   e.Name,
			Sig:    e.Sig,
			Val:    e.Val,
			Kind:   kind,
			Under:  prefix,
			Blocks: blocks(e.Doc),
			Src:    fmt.Sprintf("%s#L%d", srcName(p, e.File), e.Line),
			Kids:   sections(p, e.Children, id, depth+1),
			Depth:  depth,
		})
	}
	return out
}

// blocks cuts a comment into its paragraphs and its examples. A run of lines
// that all start with a tab is an example, which is how the standard library
// writes one; anything else is a paragraph.
func blocks(doc string) []block {
	if doc == "" {
		return nil
	}

	var (
		out   []block
		lines []string
		code  bool
	)

	flush := func() {
		if len(lines) == 0 {
			return
		}
		text := strings.Join(lines, "\n")
		if !code {
			// A paragraph was wrapped by whoever wrote it, at a width that is
			// not the width of the page: the browser wraps it again.
			text = strings.Join(lines, " ")
		}
		text = strings.TrimRight(text, "\n")
		b := block{Text: text, Code: code}
		if code {
			b.HTML = highlight(text)
		}
		out = append(out, b)
		lines = nil
	}

	for _, line := range strings.Split(doc, "\n") {
		indented := strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "    ")

		switch {
		case strings.TrimSpace(line) == "":
			flush()
		case indented != code:
			flush()
			code = indented
			lines = append(lines, dedent(line))
		default:
			lines = append(lines, dedent(line))
		}
	}
	flush()

	return out
}

func dedent(line string) string {
	if strings.HasPrefix(line, "\t") {
		return strings.TrimPrefix(line, "\t")
	}
	return strings.TrimPrefix(line, "    ")
}

// page is the template the HTML is written through. It lives in a file of its
// own rather than in a string here: it is HTML and CSS, and it is worth an
// editor that knows that.
//
//go:embed page.tmpl.html
var pageHTML string

// The style both pages share, kept as a stylesheet so that an editor treats
// it as one.
//
//go:embed page.css
var pageCSS string

// The listing of one source file, which is where a name on the doc page
// points when the reader wants to know how it is done.
//
//go:embed source.tmpl.html
var sourceHTML string

var page = template.Must(template.New("doc").Funcs(template.FuncMap{
	"indent": func(depth int) template.CSS {
		return template.CSS(fmt.Sprintf("%.2frem", float64(depth)*0.9))
	},
	"css": func() template.CSS { return template.CSS(pageCSS) },
}).Parse(pageHTML))

var source = template.Must(page.New("source").Parse(sourceHTML))
