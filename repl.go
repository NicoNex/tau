package tau

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/item"
	"github.com/NicoNex/tau/internal/lexer"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/vm"
	"golang.org/x/term"
)

// The prompts: one for a statement, one for the rest of a statement that is
// not finished yet.
const (
	prompt1 = ">>> "
	prompt2 = "... "
)

// session is a REPL: the state the definitions live in, and the symbols the
// compiler needs to find them again on the next line.
//
// The two are one thing and have to be kept in step - a definition that the
// symbol table knows about and the state does not is a name that resolves to
// somebody else's value - so nothing here touches one without the other.
type session struct {
	state   vm.State
	symbols *compiler.SymbolTable
	out     io.Writer
}

func newSession(out io.Writer) *session {
	return &session{
		state:   vm.NewState(),
		symbols: loadBuiltins(compiler.NewSymbolTable()),
		out:     out,
	}
}

func (s *session) free() { s.state.Free() }

// eval runs one piece of source and returns what it came to, which is null for
// anything that is not an expression.
func (s *session) eval(src string) (obj.Object, error) {
	tree, errs := parser.Parse("<stdin>", src)
	if len(errs) > 0 {
		msgs := make([]string, len(errs))
		for i, e := range errs {
			msgs[i] = e.Error()
		}
		return obj.NullObj, errors.New(strings.Join(msgs, "\n"))
	}

	c := compiler.NewWithState(s.symbols, s.state.NumConsts())
	c.SetFileInfo("<stdin>", src)
	if err := c.Compile(tree); err != nil {
		return obj.NullObj, err
	}

	tvm := vm.NewWithState("<stdin>", c.Bytecode(), s.state)
	tvm.Run()

	res := tvm.LastPoppedStackObj()
	s.state = tvm.State()
	s.symbols.NumDefs = s.state.NumDefs()
	tvm.Free()

	return res, nil
}

// print writes what a line came to, the way a prompt should: nothing at all
// for the statements that are not expressions, since a screen full of null is
// a screen with nothing on it.
func (s *session) print(res obj.Object) {
	if res.Type() == obj.NullType {
		return
	}
	fmt.Fprintln(s.out, res)
}

// names is everything a name could be completed to here: what is defined so
// far, and what is always in scope.
func (s *session) names() []string {
	out := s.symbols.Names()
	out = append(out, obj.Builtins[:]...)
	sort.Strings(out)
	return out
}

// REPL is the prompt, on a terminal that can be put in raw mode.
func REPL() error {
	initState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("error opening terminal: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), initState)

	t := term.NewTerminal(&interrupts{rw: os.Stdin}, prompt1)
	hist := openHistory()
	if hist != nil {
		t.History = hist
		defer hist.close()
	}

	s := newSession(t)
	defer s.free()

	t.AutoCompleteCallback = completer(s)
	redirectStdout(t)
	PrintVersionInfo(t)
	fmt.Fprintln(t, `Type ":help" for the commands, ":quit" or Ctrl-D to leave.`)

	for {
		line, err := t.ReadLine()
		if err != nil {
			// Ctrl-D on an empty line, or the input ran out.
			if err == io.EOF {
				fmt.Fprintln(t)
				return nil
			}
			return err
		}

		// The rest of an unfinished statement, however many lines it takes.
		if src := strings.TrimRight(line, " \t"); unclosed(src) > 0 {
			t.SetPrompt(prompt2)
			line, err = readRest(t, src)
			t.SetPrompt(prompt1)
			if err != nil {
				if err == io.EOF {
					fmt.Fprintln(t)
					return nil
				}
				return err
			}
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		if done, err := s.command(line); done {
			if errors.Is(err, errQuit) {
				return nil
			}
			if err != nil {
				fmt.Fprintln(t, err)
			}
			continue
		}

		res, err := s.eval(line)
		if err != nil {
			fmt.Fprintln(t, err)
			continue
		}
		s.print(res)
	}
}

// readRest reads until the statement started on the first line is closed.
func readRest(t *term.Terminal, first string) (string, error) {
	var buf strings.Builder

	buf.WriteString(first)
	buf.WriteByte('\n')

	for {
		line, err := t.ReadLine()
		if err != nil {
			return "", err
		}

		buf.WriteString(strings.TrimRight(line, " \t"))
		buf.WriteByte('\n')

		src := buf.String()
		if unclosed(src) <= 0 {
			return src, nil
		}
		// An empty line gets out of something that will not close, which is
		// the way out of a bracket typed by mistake.
		if strings.TrimSpace(line) == "" {
			return src, nil
		}
	}
}

// unclosed counts the brackets a piece of source leaves open, which is what
// says whether it is a statement or the beginning of one.
//
// It is counted on the tokens rather than on the characters, so a brace inside
// a string or a comment is a brace nobody has to close.
func unclosed(src string) int {
	var (
		depth int
		ok    = true
	)

	for i := range lexer.Lex(src) {
		if i.Is(item.EOF) {
			break
		}
		// A string that is never closed is a line that is not finished, but it
		// is not a bracket either: the lexer says so and the parser will say
		// it again with a better message.
		if i.Is(item.Error) {
			ok = false
			break
		}

		switch i.Typ {
		case item.LBrace, item.LBracket, item.LParen:
			depth++
		case item.RBrace, item.RBracket, item.RParen:
			depth--
		}
	}

	if !ok {
		return 0
	}
	return depth
}

// What an interrupt is once the terminal is raw: a byte, since the driver
// that would have turned it into a signal is out of the way. The two after it
// are what the line editor understands as "go to the end" and "delete back to
// the start", which together are the line thrown away.
const (
	ctrlC = 3
	ctrlE = 5
	ctrlU = 21
)

// interrupts turns Ctrl-C into throwing away the line being typed.
//
// The line editor answers a Ctrl-C with io.EOF, the same thing it says for a
// Ctrl-D, so by the time a caller sees it there is no telling the one that
// means "never mind this line" from the one that means "I am done". It never
// gets that far: the byte is swapped out here, for the two the editor reads
// as the end of the line and the deletion of everything before it.
type interrupts struct {
	// The terminal both reads from and writes to this: on a tty they are the
	// same file, and only the reading is meddled with.
	rw   io.ReadWriter
	rest []byte
	err  error
}

func (in *interrupts) Write(p []byte) (int, error) { return in.rw.Write(p) }

func (in *interrupts) Read(p []byte) (int, error) {
	if len(in.rest) > 0 {
		n := copy(p, in.rest)
		in.rest = in.rest[n:]
		return n, nil
	}
	if in.err != nil {
		err := in.err
		in.err = nil
		return 0, err
	}

	buf := make([]byte, len(p))
	n, err := in.rw.Read(buf)

	out := make([]byte, 0, n)
	for _, b := range buf[:n] {
		if b == ctrlC {
			out = append(out, ctrlE, ctrlU)
			continue
		}
		out = append(out, b)
	}

	n = copy(p, out)
	// What did not fit waits for the next call, and so does the error: giving
	// it back now would be giving it back instead of the keys.
	if n < len(out) {
		in.rest = append(in.rest[:0], out[n:]...)
		in.err = err
		return n, nil
	}
	return n, err
}

// errQuit ends the session.
var errQuit = errors.New("quit")

// command runs one of the things that are about the prompt rather than about
// the language. They start with a colon, which no tau statement does.
func (s *session) command(line string) (bool, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, ":") {
		return false, nil
	}

	name, arg, _ := strings.Cut(line[1:], " ")
	arg = strings.TrimSpace(arg)

	switch name {
	case "quit", "q", "exit":
		return true, errQuit

	case "help", "h", "?":
		fmt.Fprint(s.out, `  :help            this
  :doc MODULE      what a module exports, as "tau doc" writes it
  :doc MODULE.NAME and one name inside it
  :vars            the names defined so far
  :load FILE       run a file in this session, keeping what it defines
  :reset           forget everything defined so far
  :quit            leave, as Ctrl-D does
`)
		return true, nil

	case "doc":
		if arg == "" {
			return true, errors.New(`:doc wants a module, as in ":doc strings"`)
		}
		return true, DocTo(s.out, arg, false)

	case "vars":
		names := s.symbols.Names()
		if len(names) == 0 {
			fmt.Fprintln(s.out, "nothing defined yet")
			return true, nil
		}
		sort.Strings(names)
		fmt.Fprintln(s.out, strings.Join(names, "\n"))
		return true, nil

	case "load":
		if arg == "" {
			return true, errors.New(`:load wants a file, as in ":load main.tau"`)
		}
		src, err := os.ReadFile(arg)
		if err != nil {
			return true, err
		}
		res, err := s.eval(string(src))
		if err != nil {
			return true, err
		}
		s.print(res)
		return true, nil

	case "reset":
		s.free()
		s.state = vm.NewState()
		s.symbols = loadBuiltins(compiler.NewSymbolTable())
		fmt.Fprintln(s.out, "everything defined here is gone")
		return true, nil

	default:
		return true, fmt.Errorf("%q is not a command, try :help", ":"+name)
	}
}

// completer completes the name being typed, on tab.
//
// One match is filled in. Several are printed and the part they all begin with
// is filled in instead, which is the behaviour of every shell and the reason
// tab is worth pressing at all.
func completer(s *session) func(string, int, rune) (string, int, bool) {
	return func(line string, pos int, key rune) (string, int, bool) {
		if key != '\t' {
			return "", 0, false
		}

		start := pos
		for start > 0 && isNameByte(line[start-1]) {
			start--
		}
		word := line[start:pos]
		if word == "" {
			// Nothing to go on: a tab is an indent, which is what it is for
			// in the middle of a function being typed.
			return line[:pos] + "\t" + line[pos:], pos + 1, true
		}

		var hits []string
		for _, n := range s.names() {
			if strings.HasPrefix(n, word) {
				hits = append(hits, n)
			}
		}
		if len(hits) == 0 {
			return "", 0, false
		}

		fill := hits[0]
		if len(hits) > 1 {
			fill = commonPrefix(hits)
			if fill == word {
				fmt.Fprintln(s.out)
				fmt.Fprintln(s.out, strings.Join(hits, "    "))
				return "", 0, false
			}
		}

		return line[:start] + fill + line[pos:], start + len(fill), true
	}
}

func isNameByte(c byte) bool {
	return c == '_' ||
		(c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9')
}

func commonPrefix(ss []string) string {
	if len(ss) == 0 {
		return ""
	}

	out := ss[0]
	for _, s := range ss[1:] {
		for !strings.HasPrefix(s, out) {
			out = out[:len(out)-1]
			if out == "" {
				return ""
			}
		}
	}
	return out
}

// history is the lines of every session, kept in a file so that the arrow keys
// reach further back than this one.
//
// The newest entry is index 0, which is the order the terminal asks for and
// the reverse of the order a file is read in.
type history struct {
	lines []string
	file  *os.File
}

const historyMax = 5000

func historyPath() (string, error) {
	dir := os.Getenv("XDG_STATE_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".local", "state")
	}
	dir = filepath.Join(dir, "tau")

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "history"), nil
}

// openHistory reads what is there and opens the file to add to it. A history
// that cannot be read or written is no history, and no reason not to start.
func openHistory() *history {
	path, err := historyPath()
	if err != nil {
		return nil
	}

	h := &history{}
	if f, err := os.Open(path); err == nil {
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if line := sc.Text(); line != "" {
				h.lines = append(h.lines, line)
			}
		}
		f.Close()
	}

	if n := len(h.lines); n > historyMax {
		h.lines = h.lines[n-historyMax:]
	}

	// 0600: a session holds whatever was typed into it.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil
	}
	h.file = f

	return h
}

func (h *history) Add(entry string) {
	entry = strings.TrimRight(entry, "\n")
	if entry == "" {
		return
	}
	// The same line twice running is one line worth remembering.
	if n := len(h.lines); n > 0 && h.lines[n-1] == entry {
		return
	}
	// A line with a newline in it would come back as several, so it is kept
	// for this session and not written down.
	if h.file != nil && !strings.Contains(entry, "\n") {
		fmt.Fprintln(h.file, entry)
	}
	h.lines = append(h.lines, entry)
}

func (h *history) Len() int { return len(h.lines) }

func (h *history) At(i int) string { return h.lines[len(h.lines)-1-i] }

func (h *history) close() {
	if h.file != nil {
		h.file.Close()
	}
}

// SimpleREPL is the same prompt where there is no terminal to put in raw mode:
// a pipe, a dumb console, Windows. No history and no completion, since both
// need a terminal that answers.
func SimpleREPL() {
	s := newSession(os.Stdout)
	defer s.free()

	reader := bufio.NewReader(os.Stdin)
	PrintVersionInfo(os.Stdout)

	for {
		fmt.Print(prompt1)
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println()
			return
		}

		if src := strings.TrimRight(line, " \t\n"); unclosed(src) > 0 {
			line, err = simpleReadRest(reader, src)
			if err != nil {
				fmt.Println()
				return
			}
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		if done, err := s.command(line); done {
			if errors.Is(err, errQuit) {
				return
			}
			if err != nil {
				fmt.Println(err)
			}
			continue
		}

		res, err := s.eval(line)
		if err != nil {
			fmt.Println(err)
			continue
		}
		s.print(res)
	}
}

func simpleReadRest(r *bufio.Reader, first string) (string, error) {
	var buf strings.Builder

	buf.WriteString(first)
	buf.WriteByte('\n')

	for {
		fmt.Print(prompt2)
		line, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}

		buf.WriteString(strings.TrimRight(line, " \t\n"))
		buf.WriteByte('\n')

		src := buf.String()
		if unclosed(src) <= 0 || strings.TrimSpace(line) == "" {
			return src, nil
		}
	}
}

func loadBuiltins(st *compiler.SymbolTable) *compiler.SymbolTable {
	for i, name := range obj.Builtins {
		st.DefineBuiltin(i, name)
	}
	return st
}
