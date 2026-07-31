package mod

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsRemote(t *testing.T) {
	remote := []string{
		"github.com/NicoNex/example",
		"github.com/NicoNex/example/util",
		"git.sr.ht/~z/w",
		"example.org/thing",
	}
	local := []string{
		"strings", "crypto/sha256", "encoding/json",
		"./util", "../lib", "/abs/path", "",
	}

	for _, p := range remote {
		if !IsRemote(p) {
			t.Errorf("IsRemote(%q) = false, want true", p)
		}
	}
	for _, p := range local {
		if IsRemote(p) {
			t.Errorf("IsRemote(%q) = true, want false", p)
		}
	}
}

func TestVersions(t *testing.T) {
	valid := []string{"v0.0.0", "v1.2.3", "v10.20.30"}
	invalid := []string{"1.2.3", "v1.2", "v1.2.3.4", "v1.2.x", "v01.2.3", "", "latest"}

	for _, v := range valid {
		if !ValidVersion(v) {
			t.Errorf("ValidVersion(%q) = false, want true", v)
		}
	}
	for _, v := range invalid {
		if ValidVersion(v) {
			t.Errorf("ValidVersion(%q) = true, want false", v)
		}
	}

	// A version is ordered by its numbers and not by its text, or v1.10.0
	// would come before v1.9.0 and the selection would go backwards.
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.2.3", "v1.2.3", 0},
		{"v1.9.0", "v1.10.0", -1},
		{"v1.10.0", "v1.9.0", 1},
		{"v2.0.0", "v1.99.99", 1},
		{"v1.2.3", "nonsense", 1},
	}
	for _, c := range cases {
		if got := CompareVersions(c.a, c.b); got != c.want {
			t.Errorf("CompareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestParse(t *testing.T) {
	src := `# a comment
module github.com/NicoNex/example

tau 2.0

require (
	github.com/x/y v1.4.0
	git.sr.ht/~z/w v0.3.1
)
`
	f, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if f.Module != "github.com/NicoNex/example" {
		t.Errorf("module = %q", f.Module)
	}
	if f.Tau != "2.0" {
		t.Errorf("tau = %q", f.Tau)
	}
	if len(f.Require) != 2 {
		t.Fatalf("require = %v", f.Require)
	}

	// What is written out has to parse back to the same thing, or tidy would
	// slowly rewrite a manifest into one it can no longer read.
	back, err := Parse(f.String())
	if err != nil {
		t.Fatal(err)
	}
	if back.Module != f.Module || len(back.Require) != len(f.Require) {
		t.Errorf("roundtrip changed the manifest: %+v", back)
	}

	// A single requirement is written on one line and read back the same.
	one, err := Parse("module m\n\nrequire github.com/x/y v1.0.0\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(one.Require) != 1 || one.Require[0].Version != "v1.0.0" {
		t.Errorf("single require = %v", one.Require)
	}
}

func TestParseErrors(t *testing.T) {
	bad := []string{
		"require github.com/x/y v1.0.0\n",        // no module line
		"module m\nrequire github.com/x/y\n",     // no version
		"module m\nrequire github.com/x/y 1.0\n", // not a version
		"module m\nrequire (\n",                  // block left open
		"module m\nnonsense yes\n",               // unknown verb
	}
	for _, src := range bad {
		if _, err := Parse(src); err == nil {
			t.Errorf("Parse(%q) = nil error, want one", src)
		}
	}
}

func TestSetRequire(t *testing.T) {
	f := &File{Module: "m"}
	f.SetRequire("github.com/x/y", "v1.2.0")
	f.SetRequire("github.com/x/y", "v1.4.0")
	if len(f.Require) != 1 || f.Require[0].Version != "v1.4.0" {
		t.Fatalf("require = %v, want one at v1.4.0", f.Require)
	}

	// Never downwards: the build takes the highest anybody asked for, so
	// asking again for an older one changes nothing.
	f.SetRequire("github.com/x/y", "v1.1.0")
	if f.Require[0].Version != "v1.4.0" {
		t.Errorf("version went back to %q", f.Require[0].Version)
	}
}

func TestSplit(t *testing.T) {
	known := []string{"github.com/a/b", "github.com/a/b/deep"}

	cases := []struct {
		path, mod, sub string
		ok             bool
	}{
		{"github.com/a/b", "github.com/a/b", "", true},
		{"github.com/a/b/util", "github.com/a/b", "util", true},
		// The longest requirement wins, so a module may hold another one's
		// path as a prefix without swallowing its imports.
		{"github.com/a/b/deep/x", "github.com/a/b/deep", "x", true},
		{"github.com/other/thing", "", "", false},
		// A prefix that is not a path element boundary is not a match.
		{"github.com/a/bb", "", "", false},
	}
	for _, c := range cases {
		mod, sub, ok := Split(c.path, known)
		if ok != c.ok || mod != c.mod || sub != c.sub {
			t.Errorf("Split(%q) = %q, %q, %v; want %q, %q, %v",
				c.path, mod, sub, ok, c.mod, c.sub, c.ok)
		}
	}
}

func TestUsed(t *testing.T) {
	imports := []string{"github.com/a/b/util", "github.com/c/d"}
	required := []string{"github.com/a/b", "github.com/c/d", "github.com/e/f"}

	used := Used(imports, required)
	if !used["github.com/a/b"] || !used["github.com/c/d"] {
		t.Errorf("an imported module is not used: %v", used)
	}
	if used["github.com/e/f"] {
		t.Errorf("a module nothing imports is used: %v", used)
	}
}

func TestFiles(t *testing.T) {
	dir := t.TempDir()

	for _, name := range []string{"b.tau", "a.tau", "z_test.tau", "notes.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x = 1\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}

	files, err := Files(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Sorted, because an order there has to be and that is the only one a
	// reader can predict. A test file is about the module, not part of it,
	// and neither a subdirectory nor a file that is not tau belongs.
	want := []string{filepath.Join(dir, "a.tau"), filepath.Join(dir, "b.tau")}
	if len(files) != len(want) {
		t.Fatalf("Files = %v, want %v", files, want)
	}
	for i := range want {
		if files[i] != want[i] {
			t.Errorf("Files[%d] = %s, want %s", i, files[i], want[i])
		}
	}

	if !IsDirModule(dir) {
		t.Error("a directory of tau files is not a module")
	}
	if IsDirModule(filepath.Join(dir, "sub")) {
		t.Error("a directory with no tau file is a module")
	}
	if IsDirModule(filepath.Join(dir, "a.tau")) {
		t.Error("a file is a directory module")
	}
}

func TestPathMajor(t *testing.T) {
	cases := []struct {
		path  string
		major int
		base  string
	}{
		{"github.com/x/y", 0, "github.com/x/y"},
		{"github.com/x/y/v2", 2, "github.com/x/y"},
		{"github.com/x/y/v10", 10, "github.com/x/y"},
		// v0 and v1 are the path with no suffix, so those are directories.
		{"github.com/x/y/v1", 0, "github.com/x/y/v1"},
		{"github.com/x/y/v0", 0, "github.com/x/y/v0"},
		{"github.com/x/y/v02", 0, "github.com/x/y/v02"},
		{"github.com/x/y/view", 0, "github.com/x/y/view"},
	}
	for _, c := range cases {
		major, base := PathMajor(c.path)
		if major != c.major || base != c.base {
			t.Errorf("PathMajor(%q) = %d, %q; want %d, %q", c.path, major, base, c.major, c.base)
		}
	}

	// A tag belongs to the path that names its major, and to no other.
	yes := [][2]string{
		{"github.com/x/y", "v0.3.0"},
		{"github.com/x/y", "v1.0.0"},
		{"github.com/x/y/v2", "v2.1.0"},
	}
	no := [][2]string{
		{"github.com/x/y", "v2.0.0"},
		{"github.com/x/y/v2", "v1.9.0"},
		{"github.com/x/y/v2", "v3.0.0"},
	}
	for _, c := range yes {
		if !MatchesMajor(c[0], c[1]) {
			t.Errorf("MatchesMajor(%q, %q) = false, want true", c[0], c[1])
		}
	}
	for _, c := range no {
		if MatchesMajor(c[0], c[1]) {
			t.Errorf("MatchesMajor(%q, %q) = true, want false", c[0], c[1])
		}
	}
}

func TestParseMeta(t *testing.T) {
	page := `<html><head>
		<meta name="go-import" content="tau.dev/text git https://github.com/other/text">
		<meta name="tau-import" content="tau.dev/text git https://github.com/x/text">
	</head></html>`

	// Both tags answer, and the tau one is the one that means this language.
	prefix, url, err := parseMeta(page, "tau.dev/text")
	if err != nil {
		t.Fatal(err)
	}
	if prefix != "tau.dev/text" || url != "https://github.com/x/text" {
		t.Errorf("parseMeta = %q, %q", prefix, url)
	}

	// A host serving only Go modules answers the same question.
	prefix, url, err = parseMeta(
		`<meta name="go-import" content="tau.dev/text git https://github.com/other/text">`,
		"tau.dev/text")
	if err != nil || url != "https://github.com/other/text" {
		t.Errorf("parseMeta go-import = %q, %q, %v", prefix, url, err)
	}

	// The attributes the other way round are as valid, and as common.
	_, url, err = parseMeta(
		`<meta content="tau.dev/text git https://example.org/t.git" name="tau-import"/>`,
		"tau.dev/text")
	if err != nil || url != "https://example.org/t.git" {
		t.Errorf("parseMeta reversed = %q, %v", url, err)
	}

	// A tag about some other module is not an answer.
	if _, _, err := parseMeta(
		`<meta name="tau-import" content="other.dev/thing git https://x/y">`,
		"tau.dev/text"); err == nil {
		t.Error("a tag for another prefix was accepted")
	}

	// Neither is a page with no tag at all.
	if _, _, err := parseMeta("<html><body>hi</body></html>", "tau.dev/text"); err == nil {
		t.Error("a page with no tag was accepted")
	}
}
