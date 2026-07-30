package mod

import "testing"

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
