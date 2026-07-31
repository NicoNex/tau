package format

import (
	"os"
	"testing"
)

// TestSource is the check the formatter needs: what comes out has to hold the
// same tokens as what went in, whatever the input looked like.
func TestSource(t *testing.T) {
	const src = `# a comment
x = 1+2
f = fn(a, b) {
	if a>b {
		return a
	}
	return -b
}
l = [1, 2, 3]
m = {"a": 1}
for i = 0; i < 3; ++i {
	println(l[i], m["a"], f(x, i))  # trailing
}
`

	out, err := Source("test.tau", src)
	t.Log("\n" + out)
	if err != nil {
		t.Fatal(err)
	}

	// Formatting what is already formatted changes nothing.
	again, err := Source("test.tau", out)
	if err != nil {
		t.Fatal(err)
	}
	if again != out {
		t.Errorf("not stable:\n%s\nbecame\n%s", out, again)
	}
}

// TestNesting is the case a line opening more than one bracket makes: the
// indentation goes up by one level, not by one per bracket, and the line that
// closes them comes back to the one that opened them.
func TestNesting(t *testing.T) {
	const src = `testing.Main([
["a", fn(t) {
t.AssertEq(1, 1)
}]
])
`
	const want = `testing.Main([
	["a", fn(t) {
		t.AssertEq(1, 1)
	}]
])
`

	out, err := Source("test.tau", src)
	if err != nil {
		t.Fatal(err)
	}
	if out != want {
		t.Errorf("got\n%s\nwant\n%s", out, want)
	}
}

// TestStdlib formats every module of the standard library: they are the
// largest tau sources around, and each of them has to survive untouched.
func TestStdlib(t *testing.T) {
	files, err := os.ReadDir("../../stdlib")
	if err != nil {
		t.Skip("no stdlib to format")
	}

	for _, f := range files {
		if f.IsDir() || len(f.Name()) < 4 || f.Name()[len(f.Name())-4:] != ".tau" {
			continue
		}

		src, err := os.ReadFile("../../stdlib/" + f.Name())
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Source(f.Name(), string(src)); err != nil {
			t.Errorf("%s: %v", f.Name(), err)
		}
	}
}
