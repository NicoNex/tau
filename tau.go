package tau

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/NicoNex/tau/internal/ast"
	bundlepkg "github.com/NicoNex/tau/internal/bundle"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/format"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/vm"
)

// TauVersion is what `tau version` prints. It is not written down anywhere:
// the build sets it from the tag the working tree is on, see the Makefile,
// and a copy installed with `go install` reads it out of the module version
// instead. Only a build from a tree with no git and no module info says
// "devel", and nothing has to be edited before tagging a release.
var TauVersion = version()

// version returns what the build knows about itself.
func version() string {
	// -ldflags -X put the tag here, and there is nothing to work out.
	if tagVersion != "" {
		return tagVersion
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "devel"
	}

	// Built straight from a checkout without the Makefile: the commit says
	// more than the version go infers from the tags, which for a module path
	// with no /v2 in it is not a v2 tag at all.
	var rev, modified string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			modified = s.Value
		}
	}
	if rev != "" {
		if len(rev) > 12 {
			rev = rev[:12]
		}
		if modified == "true" {
			return "devel-" + rev + "-dirty"
		}
		return "devel-" + rev
	}

	// Installed from the module proxy, where the version is all there is.
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	return "devel"
}

// tagVersion is set at link time by the Makefile.
var tagVersion string

var ErrParseError = errors.New("error: parse error")

func readFile(fname string) []byte {
	b, err := os.ReadFile(fname)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return b
}

func writeFile(fname string, cont []byte) {
	if err := os.WriteFile(fname, cont, 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func compile(path string) (bc compiler.Bytecode, err error) {
	input := string(readFile(path))
	res, errs := parser.Parse(path, input)
	if len(errs) > 0 {
		var buf strings.Builder

		for _, e := range errs {
			buf.WriteString(e.Error())
			buf.WriteByte('\n')
		}
		return compiler.Bytecode{}, errors.New(buf.String())
	}

	c := compiler.New()
	c.SetFileInfo(path, input)
	if err = c.Compile(res); err != nil {
		return
	}

	return c.Bytecode(), nil
}

func ExecFileVM(f string) (err error) {
	var bytecode compiler.Bytecode

	if filepath.Ext(f) == ".tauc" {
		raw := readFile(f)

		if bundlepkg.Is(raw) {
			var clean func()

			if bytecode, clean, err = bundlepkg.Open(raw); err != nil {
				return err
			}
			defer clean()
		} else {
			bytecode = compiler.DecodeBytecode(raw)
		}
	} else {
		// Before anything runs: a module that cannot be found is an error the
		// author wants now, not on the branch that reaches it.
		if err = CheckImports(f); err != nil {
			return err
		}
		if bytecode, err = compile(f); err != nil {
			return err
		}
	}

	return runBytecode(f, bytecode)
}

// runBytecode runs a compiled program. What went wrong is returned rather
// than printed: the caller is the one that knows where its messages go, and
// printing here as well is how the same error ended up on the screen twice.
func runBytecode(name string, bytecode compiler.Bytecode) error {
	return bundlepkg.Run(name, bytecode)
}

// SetArgs hands the command line to the program about to run.
func SetArgs(args []string) { vm.SetArgs(args) }

// CompileFiles compiles each file into a self contained '.tauc' bundle. With
// out empty each bundle is written next to its source.
func CompileFiles(files []string, out string) error {
	if out != "" && len(files) > 1 {
		return errors.New("build: -o takes a single input file")
	}

	for _, f := range files {
		bundle, err := Bundle(f)
		if err != nil {
			return err
		}

		dst := out
		if dst == "" {
			ext := filepath.Ext(f)
			dst = f[:len(f)-len(ext)] + ".tauc"
		}
		writeFile(dst, bundle)
	}

	return nil
}

func PrintVersionInfo(w io.Writer) {
	fmt.Fprintf(w, "Tau %s on %s\n", TauVersion, strings.Title(runtime.GOOS))
}

func Parse(src string) (ast.Node, error) {
	tree, errs := parser.Parse("<input>", src)
	if len(errs) > 0 {
		var buf strings.Builder

		buf.WriteString("parser error:\n")
		for _, e := range errs {
			buf.WriteString(e.Error())
			buf.WriteByte('\n')
		}

		return nil, errors.New(buf.String())
	}

	return tree, nil
}

// TestFiles runs every *_test.tau in the given paths, a directory standing
// for the test files it holds, and reports how many of them failed. Each file
// runs in its own process the way `go test` does with its packages, so a test
// that exits or crashes takes down only itself.
func TestFiles(paths []string) error {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			files = append(files, p)
			continue
		}

		// A directory stands for everything below it, the way "./..." does
		// in Go: the tests of a module in a subdirectory are still its tests.
		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(path, "_test.tau") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	if len(files) == 0 {
		return errors.New("no test files found")
	}
	sort.Strings(files)

	self, err := os.Executable()
	if err != nil {
		self = os.Args[0]
	}

	var failed int
	for _, f := range files {
		fmt.Printf("=== %s\n", f)

		cmd := exec.Command(self, f)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			failed++
		}
	}

	fmt.Println()
	if failed > 0 {
		return fmt.Errorf("%d of %d test files failed", failed, len(files))
	}
	fmt.Printf("ok      %d test files passed\n", len(files))
	return nil
}

// FormatFiles rewrites the given tau files in the canonical style, walking
// directories for '.tau' files. With write the files are rewritten in place,
// with list only the names of the ones that differ are printed, otherwise the
// formatted source goes to standard output.
func FormatFiles(paths []string, write, list bool) error {
	var files []string

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			files = append(files, p)
			continue
		}

		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && filepath.Ext(path) == ".tau" {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	sort.Strings(files)

	var failed int
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			return err
		}

		out, err := format.Source(f, string(src))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", f, err)
			failed++
			continue
		}

		switch {
		case out == string(src):
		case list:
			fmt.Println(f)
		case write:
			if err := os.WriteFile(f, []byte(out), 0644); err != nil {
				return err
			}
			fmt.Println(f)
		}

		if !write && !list {
			fmt.Print(out)
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d of %d files could not be formatted", failed, len(files))
	}
	return nil
}
