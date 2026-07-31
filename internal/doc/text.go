package doc

import (
	"fmt"
	"io"
	"strings"
)

// Text writes the documentation of a package to w. path is the name asked
// for inside the module, empty for the module itself.
//
// What it shows is one level: asked for a module, it writes the module's own
// comment and then a line and a summary for each name; asked for a name, the
// whole comment of that name and then a line and a summary for each of the
// names it holds. Going deeper is asking for the deeper name.
func Text(w io.Writer, p Package, path string) error {
	if path == "" {
		return pkgText(w, p)
	}

	e, ok := p.Find(path)
	if !ok {
		return fmt.Errorf("doc: %s has no %s", p.Path, path)
	}

	fmt.Fprintf(w, "module %s\n\n", p.Path)
	fmt.Fprintln(w, decl(e))
	if e.Doc != "" {
		fmt.Fprint(w, indent(e.Doc, "    "))
	}

	return children(w, e.Children, p.Path+"."+path)
}

func pkgText(w io.Writer, p Package) error {
	fmt.Fprintf(w, "module %s\n\n", p.Path)
	if p.Doc != "" {
		fmt.Fprint(w, indent(p.Doc, "    "))
	}
	return children(w, p.Entries, p.Path)
}

// children writes the summary of everything one level down: the declaration
// and the first paragraph of what is written about it.
func children(w io.Writer, entries []Entry, under string) error {
	if len(entries) == 0 {
		return nil
	}
	fmt.Fprintln(w)

	for _, e := range entries {
		fmt.Fprintf(w, "%s\n", decl(e))
		if s := synopsis(e.Doc); s != "" {
			fmt.Fprint(w, indent(s, "    "))
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintf(w, "Use \"tau doc %s.NAME\" for one of these.\n", under)
	return nil
}

// decl is the one line that names an entry: its signature if it is a
// function, what it holds if that is short, its name alone otherwise.
func decl(e Entry) string {
	switch {
	case e.Sig != "":
		return e.Name + " = " + e.Sig
	case e.Val != "":
		return e.Name + " = " + e.Val
	default:
		return e.Name
	}
}

// synopsis is the first paragraph of a comment, which is the sentence that
// has to stand on its own.
func synopsis(doc string) string {
	if doc == "" {
		return ""
	}
	if i := strings.Index(doc, "\n\n"); i >= 0 {
		return doc[:i]
	}
	return doc
}

func indent(s, prefix string) string {
	var out strings.Builder

	for _, line := range strings.Split(s, "\n") {
		if line == "" {
			out.WriteByte('\n')
			continue
		}
		out.WriteString(prefix)
		out.WriteString(line)
		out.WriteByte('\n')
	}
	return out.String()
}
