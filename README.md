# τ

Tau is a dynamically typed, interpreted programming language. It is small on
purpose: a handful of keywords, one way to build an object, errors that are
ordinary values, and concurrency that looks like Go's. The interpreter is
written in Go around a virtual machine and an object representation written in
C; the standard library is written in tau itself, under [stdlib](stdlib), which
is the best place to read the language as it is actually used.

- [Installing](#installing)
- [The tools](#the-tools)
- [A tour of the language](#a-tour-of-the-language)
- [The elegant bits](#the-elegant-bits)
- [Errors](#errors)
- [Modules](#modules)
- [Concurrency](#concurrency)
- [C libraries](#c-libraries)
- [Shipping a program](#shipping-a-program)
- [The standard library](#the-standard-library)
- [Builtin functions](#builtin-functions)

## Installing

You need [Go](https://golang.org/) and a C compiler. libffi is a submodule, so
clone with it.

```bash
git clone --recurse-submodules https://github.com/NicoNex/tau
cd tau
make install
```

`make` alone builds `tau`, the small C runtime `tau-rt` and the shared objects
the standard library opens; `make install` copies them into a prefix. The
default prefix is `$HOME/.local`, which needs no root: the binary goes in
`~/.local/bin` and the standard library in `~/.local/lib/tau`. For a system
wide install pass the prefix:

```bash
sudo make install PREFIX=/usr/local
```

`make uninstall` removes both. `make test` runs the Go tests and then the tau
ones. `make fmt` formats the tau sources in the tree. `make tau-lsp` builds the
language server, which speaks LSP over stdin and stdout.

Building it is one way in. Every release also carries packages, which need no
compiler: a `.deb`, an `.rpm` and an Arch package for x86_64 and aarch64, an
installer for Windows, and an archive for everything else.

```bash
tar --zstd -xf tau-v2.1.0-linux-x86_64.tar.zst
./tau/install.sh                    # /usr/local as root, ~/.local otherwise
./tau/install.sh --prefix=/opt/tau  # or wherever
./tau/install.sh --uninstall
```

Any prefix works and none of them needs a variable set afterwards: the
interpreter looks for its standard library next to itself before it looks
anywhere else, so a tree moved somewhere new keeps working. Unpacking the
archive and running `tau/bin/tau` without installing anything works for the
same reason.

A module is looked up, in order, in the directories of `TAUPATH` (separated
like `PATH`), next to the file that imports it, then in `~/.local/lib/tau`,
`/usr/local/lib/tau` and `/lib/tau`. That is how a checkout runs against its
own standard library without installing anything:

```bash
TAUPATH=$PWD/stdlib ./tau myprogram.tau
```

## The tools

Everything is one binary.

```
tau                     start the REPL
tau FILE [ARGS]         run a file, shorthand for tau run
tau run FILE [ARGS]     the same, explicitly
tau build FILE...       compile to '.tauc' bytecode
tau bundle -o app FILE  compile into a standalone executable
tau test [PATH...]      run the '*_test.tau' files found in PATH
tau fmt [-w|-l] PATH... format sources in the canonical style
tau doc [-b] MODULE     what a module exports, and the comments about it
tau version             print the version
tau help COMMAND        help for one command
```

`tau version` prints what `git describe` said when the binary was built, so a
release is a tag and nothing else, and a tree between two tags says so:

```
$ tau version
Tau v2.0.15-53-g5b05dc5 on Linux
```

`tau doc` reads a module and writes what it gives whoever imports it. Tau has
no types, so what another language documents as the methods of one is here the
fields a constructor puts on the object it returns, and a name after the module
goes one level in:

```
$ tau doc sync
module sync

    sync - the locks tau routines share state with.
    ...

Mutex = fn()
    Mutex is a lock held by one routine at a time.
...

$ tau doc sync.Mutex.Lock
module sync

Lock = fn()
    Lock takes the lock, waiting until whoever holds it gives it back.
```

It follows a return: `os.Open` hands back what an unexported `newFile` built,
so `tau doc os.Open` documents the file it gives you rather than stopping at
the call. With `-b` the same is written as a page and opened in a browser, at
the name asked for:

```bash
tau doc -b encoding/json
tau doc -b sync.WaitGroup.Add   # opens the page scrolled to Add
```

Every name on the page carries a link to the source it was read from, which is
written out beside it as a listing with a line for every line of the file, and
the link lands on the one the name is written on.

The pages are written afresh every time, so they say what the module says now,
and they land in the cache directory (`~/.cache/tau/doc` on Linux) under the
name of the module. They carry their own style and their own colouring - done
by the lexer of the language itself - so they need nothing from the network and
can be thrown away at any time.

A file can also carry a shebang and run on its own:

```python
#!/usr/bin/env tau

println("hello world")
```

The REPL is multiline: a block keeps reading until it is closed.

```
Tau v2.0.15-53-g5b05dc5 on Linux
>>> repeat = fn(n, func) {
...     for i = 0; i < n; ++i {
...         func(i)
...     }
... }
...
>>> repeat(3, fn(i) { println("hello #{i}") })
hello #0
hello #1
hello #2
>>>
```

## A tour of the language

### Values

The types are `int`, `float`, `string`, `bool`, `null`, `list`, `map`,
`object`, `bytes`, `closure`, `error` and `pipe`. `type(x)` gives the name of
one as a string.

Integers can be written in base 10, 16, 2 and 8, and underscores may be used
anywhere inside a number to group digits.

Two integers divide into an integer, the way they do in C and in Go: the
remainder is dropped. A float on either side makes it a float division, as it
does for `+`, `-` and `*`, so `float(a) / b` is how two integers give a
fraction. Dividing an integer by zero is an error, where a float division
gives `inf`.

```python
# numbers
println(255, 0xff, 0b1111_1111, 0o377, 1_000_000)
println(2.5, 1.5e3, 7 / 2, 7 % 2, 7 / 2.0, float(7) / 2)

# strings
name = "tau"
println("hello, {name}", "the answer is {6 * 7}")
println(`a raw string: {name} and \n stay as they are`)

# lists, maps, objects
xs = [1, "two", 3.0]
kv = {"a": 1, 2: true}
o = new()
o.field = "value"
println(xs[1], kv["a"], o.field, len(xs), keys(kv))
```

```
255 255 255 255 1000000
2.5 1500 3 1 3.5 3.5
hello, tau the answer is 42
a raw string: {name} and \n stay as they are
two 1 value 3 [2, a]
```

Anything between braces inside a double quoted string is an expression, and its
value is put in the string. To write a brace that stands for itself, double it
or escape it: `"{{"` and `"\{"` both give one. A string in backticks is raw:
braces and backslashes are the characters they look like, which is the easier
way to write anything with braces in it, JSON included. Since a nested string
closes the one it sits in, quote it: `"{if hot { \"warm\" } else { \"cold\" }}"`.

Comments start with `#` and run to the end of the line. A newline ends a
statement; `;` does the same in the middle of a line.

### Operators

```
=  +=  -=  *=  /=  %=  &=  |=  ^=  <<=  >>=
||  &&  !
==  !=  <  >  <=  >=
+  -  *  /  %
&  |  ^  ~  <<  >>
++  --
```

From loosest to tightest, the levels are: assignment, `||`, `&&`, `|`, `^`,
`&`, `==` and `!=`, the comparisons, `<<` and `>>`, `+` and `-`, `*` `/` `%`,
the prefix operators, calls, indexing, and `.` last. That is the order C uses,
trap included: `&` is looser than `==`, so `a & b == c` is `a & (b == c)` and
wants parentheses. The shifts are looser than the arithmetic as well, so
`1 << 2 + 3` is `1 << 5`, which is `32`.

`==` on two lists or two maps compares identity, not contents; for a structural
comparison use `cmp.Equal` from the standard library.

A list, a map, a call and a parameter list may end with a comma, so that a line
can be added to one written over several lines without touching the line above
it. It is allowed, never required.

```python
primes = [
	2,
	3,
	5,
]
```

### Control flow

`for` has three shapes and no other: an empty one that never stops, one
expression that is the condition, three that are the C header.

```python
for { break }                 # forever
for i < 10 { i++ }            # while
for i = 0; i < 10; i++ { }    # the usual
```

`break` and `continue` do what they look like.

`++` and `--` come in both forms and behave as they do in C: written in front
of what they count they give back the new value, written after it they give
back the old one. Either way the variable ends up changed by one.

```python
i = 5
println(i++)  # 5, and i is now 6
println(++i)  # 7, and i is now 7
```

### Functions

A function is a value written with `fn`. The last expression of a body is its
result, so `return` is only needed to leave early.

```python
add = fn(x, y) { x + y }

fib = fn(n) {
	if n < 2 {
		return n
	}
	fib(n-1) + fib(n-2)
}

println(add(9, 1), fib(20))
```

```
10 6765
```

#### What a function sees around it

A function reads the names of the function around it and the global ones. What
it cannot do is write to them: assigning to one of those names makes a local of
its own, starting from the value that was read on the right of the assignment,
and shadowing the name from there on.

```python
mk = fn() {
	n = 0
	return fn() { n = n + 1; return n }
}
next = mk()
println(next(), next())  # 1 1, the n outside never moves
```

A closure can change what is *inside* something it captured, because the name
goes on meaning the same thing. That is what the `ref` module is for:

```python
ref = import("ref")

mk = fn() {
	n = ref.New(0)
	return fn() { n.v = n.v + 1; return n.v }
}
next = mk()
println(next(), next())  # 1 2
```

## The elegant bits

### Objects without `self`

`new()` returns an empty object; a constructor is an ordinary function that
fills one in and returns it. Methods are functions stored in its fields, and
they reach the object the same way they reach anything else: by closing over
the variable that holds it.

```python
newQueue = fn() {
	q = new()
	q.items = []

	q.Push = fn(x) { q.items = append(q.items, x); return q }
	q.Pop = fn() {
		if len(q.items) == 0 { return error("empty queue") }
		x = q.items[0]
		q.items = slice(q.items, 1, len(q.items))
		return x
	}
	q.Len = fn() { len(q.items) }

	return q
}

q = newQueue()
q.Push("a").Push("b")
println(q.Len(), q.Pop(), q.Pop(), q.Pop())

pop = q.Pop            # a method is just a value
println(type(pop), failed(pop()))
```

```
2 a b empty queue
closure true
```

There is no `self` and no `this` because there is nothing for them to do. The
receiver is `q`, a normal local of the constructor, and a method mentions it
the way it mentions any other captured variable. That buys three things:

- one concept instead of two. Closures already exist; a receiver keyword would
  be a second, parallel way of getting at the same value.
- no implicit binding to get wrong. `pop = q.Pop` still works, because `q` was
  captured when the function was made, not looked up from whatever the call
  happened to be written against.
- methods are values like any other. They can be passed, stored, wrapped, or
  replaced after the fact: `q.Pop = fn() { ... }` and every caller sees the new
  one.

The standard library is built this way throughout: `newT` in
[stdlib/testing.tau](stdlib/testing.tau) builds the `t` a test case is given,
`newFile` in [stdlib/os.tau](stdlib/os.tau) wraps a file descriptor, `New` in
[stdlib/rand.tau](stdlib/rand.tau) holds the state of a generator.

### Assignment is an expression, and so is `if`

Both give back a value, which is what makes the error idiom below read the way
it does.

```python
n = 7
kind = if n % 2 == 0 { "even" } else { "odd" }
println(kind)

println(total = 20 + 22)
```

```
odd
42
```

An `if` with no `else` that doesn't run gives `null`.

## Errors

An error is a value, made with `error(msg)`, tested with `failed(x)`. Nothing
is thrown, nothing unwinds: a function that failed returns the error, and the
caller decides.

Because assignment is an expression, the call and the test on its result are
one line, which is the shape most of the standard library is written in:

```python
strconv = import("strconv")

parse = fn(s) {
	if failed(n = strconv.Atoi(s)) {
		return error("not a number: {s}")
	}
	return n * 2
}

println(parse("21"))
println(parse("x"))
```

```
42
not a number: x
```

`exit(err)` ends the program printing the error and returning non zero.

A mistake the VM catches itself stops the program with the line it happened on:

```python
increment = fn(n) {
	return n + 1
}

increment("this will raise a runtime error")
```

```
error in file errtest.tau at line 2:
    return n + 1
             ^
unsupported operator '+' for types string and int
runtime error
```

## Modules

`import("name")` loads a file once and gives back an object. There is no export
list: the names that start with a capital letter are the ones that end up in
that object, everything else stays private to the module. A path with slashes
reaches into a directory, as in `import("crypto/sha256")`.

```python
# greeting.tau
prefix = "hello"

Greet = fn(name) { "{prefix}, {name}" }

newGreeter = fn(prefix) {
	g = new()
	g.prefix = prefix
	g.Greet = fn(name) { "{g.prefix}, {name}" }
	return g
}

Formal = newGreeter("good evening")
```

```python
# main.tau
greeting = import("greeting")

println(greeting.Greet("world"))
println(greeting.Formal.Greet("world"))
println(keys(greeting))
```

```
hello, world
good evening, world
[Greet, Formal]
```

A module being a plain object is worth saying out loud: it is the same thing
`new()` makes, so it can be passed to a function, kept in a list, or have its
fields read with `keys`. `prefix` and `newGreeter` are not in it, and neither
is the `prefix` field of `Formal`, but the exported `Greet` closes over both
and works anyway.

Importing a file twice gives the same module: the table is keyed on the
absolute path, so two spellings of one file still run it once. A cycle is an
error rather than a crash.

### Modules from somewhere else

An import path whose first element is a host names a module that does not live
on this machine:

```python
example = import("github.com/NicoNex/example")
util = import("github.com/NicoNex/example/util")
```

That is the whole rule: a host has a dot in it, so `strings` and
`crypto/sha256` are the library, `./util` is the file next door, and
`github.com/...` is fetched. There is no registry, because an import path is
already an address.

A module says who it is in a `tau.mod` at its root:

```
module github.com/NicoNex/example

tau 2.0

require (
	github.com/x/y v1.4.0
	git.sr.ht/~z/w v0.3.1
)
```

`tau mod init PATH` writes one, `tau get PATH[@VERSION]` adds a requirement and
fetches it, `tau mod tidy` makes the file say what the source actually imports.
Versions are git tags of the form `v1.2.3`.

What a module calls itself has to be what it was required as. A repository
that moved, a fork required under the name of its origin and a v2 required
without the suffix that makes it one would all otherwise build, and build
something other than what was asked for.

#### Two versions of the same library

From v2 onwards the major version is part of the path, and the module says so
itself:

```
module github.com/NicoNex/example/v2
```

```python
one = import("github.com/NicoNex/example")
two = import("github.com/NicoNex/example/v2")
```

Two majors of a library are two modules, incompatible by definition, and a
program that reaches both through its dependencies has to be able to hold
both. A version that is not part of the path cannot do that. Tags of the wrong
major are invisible to a path: `v2.1.0` answers for `.../v2` and for nothing
else, and a path with no suffix is `v0` and `v1`.

#### A name that is not a forge

An import path is an address, but it does not have to be the address of a
repository. A host can answer for a name it keeps:

```html
<meta name="tau-import" content="tau.dev/text git https://github.com/x/text">
```

Served under the path itself, that redirects the name to wherever the code is
this year. Without it a project that changes host has to change every import
that ever mentioned it, and the name of a library ends up belonging to a forge
rather than to whoever wrote it. A `go-import` tag is read the same way: a host
already serving Go modules is answering the same question, and there is no
reason to make anyone publish it twice.

The forges whose paths are already their clone URLs are asked nothing, so the
usual import costs no request at all.

Where two modules ask for different versions of a third, the build takes the
highest of what was asked, and never a version nobody asked for. That rule is
minimum version selection, and what it buys is an answer that needs no solver
and does not change between today and next year.

Fetched modules land in `~/.tau/pkg`, or under `$TAUHOME`, one directory per
version. They are read once written, so a version is the bytes it was the day
it arrived and two projects can share the tree. `tau.sum` holds the hash of
every version the build reads: a tag moved after the fact stops the build
instead of running.

Nothing is fetched while a program runs. `tau get` and `tau mod tidy` reach the
network, a build reads what is already there, and a bundled program carries it.
Fetching goes through `git`, which is therefore needed to *get* a module and
not to build or run one.

### A module of several files

A module is one file, or the directory holding several. The files of a
directory are one scope: a name any of them defines is a name all of them see,
and the capitalised ones are what the module hands out.

```
shapes/
	area.tau         Area, and a factor it keeps to itself
	util.tau         Total, and the scale both files use
	shapes_test.tau
```

```python
shapes = import("./shapes")

println(shapes.Area(3, 4), shapes.Total([1, 2, 3]), keys(shapes))
```

```
24 12 [Area, Name, Total]
```

Splitting a file in two is then a matter of layout and nothing else. As things
were, two files sharing a helper meant exporting it, and the shape of the
source decided what a module made public.

Functions see each other whatever the order and whatever the file: a name used
before its definition reserves its place, and the definition fills it. What the
order decides is only when top level code runs, and that order is the names of
the files. Test files are about the module rather than part of it, so they are
left out.

Where both a file and a directory carry a name the directory is the module,
which is the only way round that lets `os` be a module and `os/exec` another
one inside it. The library is laid out that way throughout: every module is a
directory, `os/os.tau` and `os/exec/exec.tau`, so any of them can grow to a
second file without moving.

A lone file stays a module all the same, and is what most programs reach for:
`import("./helper")` is `helper.tau` next door, and nothing has to be a
directory until it has a reason to be.

The same holds for a module from elsewhere: `github.com/NicoNex/example` is the
root of that repository when it holds tau files, and `example.tau` at the root
otherwise.

### Tests

`tau test` runs every `*_test.tau` in the paths it is given, each in its own
process. A test file hands its cases to `testing.Main`:

```python
# double_test.tau
testing = import("testing")

double = fn(x) { x * 2 }

testing.Main([
	["double", fn(t) {
		t.AssertEq(double(21), 42)
	}],
	["double of zero", fn(t) {
		t.AssertEq(double(0), 1)
	}]
])
```

```
$ tau test double_test.tau
=== double_test.tau
--- PASS: double (0ms)
--- FAIL: double of zero (0ms)
        got 0, want 1
FAIL    1 failed, 1 passed of 2 (0ms)

1 of 1 test files failed
```

## Concurrency

`tau f(x)` runs a call in a tau-routine of its own. Values move between them
through pipes, built with `pipe()` and handled by `send`, `recv` and `close`.

Pipes can be buffered or unbuffered. On an unbuffered pipe `send` hands the
value directly to a receiver: the tau-routine that sends sleeps until another
one calls `recv`. A buffered pipe holds as many values as it was created with,
so `send` returns right away and only sleeps once the buffer is full. `recv` on
an empty pipe sleeps until something is sent. `close` closes the pipe, allowing
it to be garbage collected; the values already in it are still delivered, and
once there are none left `recv` returns `null`.

That last rule is what makes a receiving loop a one-liner: `for val = recv(p)`
assigns, tests, and stops when the pipe is drained and closed.

```python
listen = fn(p) {
	for val = recv(p) {
		println(val)
	}
	println("bye bye...")
}

p = pipe()
tau listen(p)

send(p, "hello")
send(p, "world")
send(p, 123)
close(p)
```

```
hello
world
123
bye bye...
```

The `runtime` module says how many cores the program may run on, which is the
number to size a pool of workers with:

```python
runtime = import("runtime")

work = fn(id, jobs, out) {
	for j = recv(jobs) {
		send(out, j * j)
	}
}

jobs = pipe(8)
out = pipe(8)
for i = 0; i < runtime.NumCPU(); i++ {
	tau work(i, jobs, out)
}
for i = 1; i <= 6; i++ { send(jobs, i) }
close(jobs)

squares = []
for i = 0; i < 6; i++ { squares = append(squares, recv(out)) }
close(out)
println(squares)
```

```
[1, 9, 16, 4, 25, 36]
```

The results come back in whatever order the workers finished in, which is a
different one on every run.

## C libraries

There are two ways to call C, and the second is built out of the first.

`dlopen(path)` opens a shared object. The dot on the handle is `dlsym`, so
`lib.name` is a symbol, and `lib[name]` is the same lookup with a name worked
out while the program runs. `dlopen(null)` is the handle of the program
itself, which is where the C library it was linked against can be reached.

It takes the name of a file, and that name differs from system to system, so
`ffi.Lib("m")` is the one to reach for when a program has to run on more than
one: it tries the shapes the system uses -- `libm.so.6`, `libm.dylib`,
`m.dll` -- and says what it tried when none of them opens. A name with a slash
or an extension already in it is opened as it stands.

```bash
gcc -shared -o vec.so -fPIC vec.c
```

```c
// vec.c
double dot(double a, double b) { return a * b; }

const char *greet(const char *name) {
	static char buf[64];
	snprintf(buf, sizeof buf, "hello, %s", name);
	return buf;
}

void *fill(int n) {
	char *p = malloc(n);
	memset(p, 'x', n);
	return p;
}
```

### The raw way

A symbol can be called with no declaration at all. Every argument goes as an
int64 or a double, and the result comes back as a machine word you take apart
yourself: `int(x)` for a number, `string(x)` for a `char *`, `bytes(x, n)` for
a buffer, and a pointer C gave you can be handed straight back to C.

```python
libc = dlopen("libc.so.6")

p = libc.malloc(32)
libc.memcpy(p, "hello", 6)
println(int(libc.strlen(p)), string(p), string(bytes(p, 5)))
libc.free(p)

println(int(libc.abs(-5)), int(libc.atoi("1234")))
```

```
5 hello hello
5 1234
```

It is three lines to try a library, and it takes your word for the types: pass
the wrong ones and you get a wrong answer, or a signal. One limit is worth
knowing before it bites, since nothing warns you about it: **a function that
returns a `double` does not survive this**. The call is prepared with a
pointer-sized result and reads the integer register, so `lib.dot(1.5, 4.0)`
comes back as `9.88e-324` rather than `6`. Pointers and integers do survive.

### With the types written down

The `ffi` module of the standard library gives a symbol a signature, and from
then on the arguments travel as those types and the result comes back as a tau
value.

A signature is a C declaration. The name of the function and the names of the
arguments may be there or not, so a line copied out of a header works as it
stands.

```python
ffi = import("ffi")
lib = dlopen("./vec.so")

dot = ffi.Func(lib.dot, "double dot(double a, double b)")
greet = ffi.Func(lib.greet, "const char *greet(const char *name)")
fill = ffi.Func(lib.fill, "void *fill(int n)")

println(dot(1.5, 4), greet("tau"), string(bytes(fill(4), 4)))
```

```
6 hello, tau xxxx
```

An integer where the signature says `double` is converted on the way in, so
`dot(1.5, 4)` needs no help.

`ffi.Bind` does a whole library at once, naming each function the way its
signature names it, and takes the name of a library as well as a handle -- the
same names `ffi.Lib` understands:

```python
m = ffi.Bind("m", [
	"double pow(double, double)",
	"double sqrt(double)",
])
println(m.pow(2.0, 10.0), m.sqrt(2))
```

```
1024 1.4142135623730951
```

Memory comes from the C library rather than from the language: `ffi.Alloc` and
`ffi.Free` are `malloc` and `free`, `ffi.Write` is `memcpy` into a pointer,
`ffi.String` reads a C string up to its NUL, and `ffi.Read(p, n)` is
`bytes(p, n)`.

### The other direction

`ffi.Export` turns a tau function into one C can call, which is what a library
wants when it takes a handler, a comparator or a visitor:

```python
ffi = import("ffi")
libc = dlopen("libc.so.6")

qsort = ffi.Func(libc.qsort, "void qsort(void *base, size_t n, size_t size, void *cmp)")

cmp = ffi.Export("int compare(const void *a, const void *b)", fn(a, b) {
	return bytes(a, 1)[0] - bytes(b, 1)[0]
})

buf = bytes([5, 3, 9, 1, 7])
qsort(buf, 5, 1, cmp)
println(buf)
```

```
[1, 3, 5, 7, 9]
```

What comes back is an ordinary function pointer as far as C is concerned. The
call is answered by the VM of the thread that entered C, so a handler called
from the loop of a library works; one called from a thread the library made
for itself finds no tau there and gets a zero back. A function that fails
inside a callback does not unwind through C: the failure stops there, and the
C side carries on with what it was given.

The exported function lives for as long as the program does. Whoever was
handed the pointer never says when they are finished with it, and freeing a
trampoline C still holds is the one crash this cannot have.

The types are the ones C writes, the exact width ones of `stdint.h`, and the
same widths spelled the way tau spells its own:

| written | is |
|---|---|
| `void` | nothing, as a result, and no arguments as a parameter list |
| `bool`, `_Bool` | a boolean |
| `char`, `short`, `int`, `long`, `long long` | the signed integer of that width on this machine, `unsigned` in front for the other sign |
| `size_t`, `ssize_t`, `intptr_t`, `uintptr_t` | the width they have on this machine |
| `int8_t` … `uint64_t`, `int8` … `uint64` | exactly that many bits |
| `float`, `double`, `float32`, `float64` | a float |
| `char *` | a tau string, NULL terminated on the way out and copied on the way back |
| any other `T *`, `T []`, `pointer` | an address: a `bytes` buffer, a string, a native value or an integer |

`const`, `volatile` and `restrict` are read and ignored, as is the room around
anything. A variadic signature is refused, since the call is prepared once and
a `...` says nothing about what will be passed: write the types this particular
call passes, `int(char *, double)` rather than `int(const char *, ...)`.

The call is prepared once, when `ffi.Func` is given the signature, not on every
call. It passes an error through untouched, so a failed symbol lookup says so
instead of turning into a nonsense pointer.

Underneath, `ffi.Func` reads the declaration and calls `cfunc(sym, ret, args)`,
which takes the types as numbers and prepares the call. That is the whole of
the C side: reading a declaration is tau, in `stdlib/ffi.tau`.

A pointer that came back from C is read with `bytes(ptr, n)`, which copies `n`
bytes out of it: that is how a returned buffer, or a struct, is brought into
tau. `bytes(n)` on its own allocates an empty buffer of `n` bytes, the way
`make([]byte, n)` does in Go. For a word whose width is known, `int(x, bits)`
sign-extends the right number of them.

The standard library uses all of this. [stdlib/math.tau](stdlib/math.tau) is
the shortest example and carries no C of its own: the interpreter is linked
against libm, so it declares the functions it wants off `dlopen(null)`.
[stdlib/syscall](stdlib/syscall), [stdlib/runtime](stdlib/runtime) and
[stdlib/sync/atomic](stdlib/sync/atomic) are small shared objects of their own,
opened with `dlopen`. The last of those is the case where the trick math plays
cannot work: an atomic add is an instruction the compiler writes inline, not a
function in a library with a name to look up, so there has to be C of our own
around it.

## Shipping a program

`tau bundle` writes a program and everything it imports into a single
executable: a copy of the runtime with the program appended to it. The result
runs where tau is not installed and needs nothing else, the modules and the
shared objects it uses included.

```bash
$ tau bundle -o hello main.tau
$ ./hello
hello, world
```

The modules travel compiled, so a bundled program starts without a parser and
without a compiler. The executable is built on `tau-rt`, a runtime written in C
that carries the VM, the objects and the bytecode decoder and nothing else, so
a bundled program weighs a few hundred kilobytes rather than a few megabytes
and starts in about two milliseconds. `make install` puts `tau-rt` next to the
standard library, in `PREFIX/lib/tau`, which is where `tau bundle` looks for
it.

An `import` of a path worked out at run time cannot be bundled and is refused
there: it was never in the source for the bundler to find, and there is no
compiler left to answer it.

`tau build` is the lighter option for machines that already have tau: it writes
a `.tauc` carrying the same program and its dependencies, without the runtime
in front of it.

On macOS the executable has to be signed before it will start, because
appending to one invalidates the signature it came with:

```bash
codesign -s - myapp
```

## The standard library

All of it is under [stdlib](stdlib) and all of it is readable tau. Each file
starts with a comment saying what it is for.

| module | |
|---|---|
| `buffer` | growing buffers of text and bytes |
| `bufio` | buffered reading and writing |
| `cmp` | comparison and ordering of values, `Equal` and `Compare` |
| `crypto/hmac` | the keyed hash of RFC 2104, over SHA-256 |
| `crypto/sha256` | the hash of FIPS 180-4 |
| `encoding/base64` | the encoding of RFC 4648 |
| `encoding/csv` | comma separated values, the way RFC 4180 writes them |
| `encoding/hex` | hexadecimal encoding |
| `encoding/json` | JSON, parsed and written |
| `encoding/xml` | XML, parsed into a tree of elements and written back |
| `errno` | the numbers `errno` takes |
| `errors` | errors as values: `New`, `Wrap`, `Is`, `Message` |
| `flag` | command line flags |
| `io` | moving bytes between streams |
| `list` | operations on lists |
| `log` | messages with a date on them |
| `maps` | operations on maps and objects |
| `math` | the usual functions on floats |
| `math/rand` | pseudo random numbers, and `Crypto` for the ones that matter |
| `net` | TCP and UDP sockets, in the shape of Go's net package |
| `net/http` | HTTP/1.1 client and server, in the shape of Go's net/http |
| `os` | files, environment, working directory, `os.Args` |
| `os/exec` | running other programs |
| `path` | slash separated paths |
| `ref` | a cell holding a value that a closure has to change |
| `regexp` | regular expressions, in the shape of Go's regexp package |
| `runtime` | what the program can know about the machine it runs on |
| `strconv` | conversions between strings and numbers |
| `strings` | operations on strings |
| `sync` | mutexes, wait groups and once, for state shared between routines |
| `sync/atomic` | reads and writes no other routine can see half of |
| `syscall` | the system calls underneath the rest |
| `testing` | the test runner `tau test` uses |
| `time` | clocks and pauses |
| `unicode/utf8` | text as code points |

`net/http` and `net` have their own documents: [HTTP_README.md](HTTP_README.md) and
[NET_README.md](NET_README.md).

A taste of a few of them:

```python
log = import("log")
rand = import("math/rand")
utf8 = import("unicode/utf8")
csv = import("encoding/csv")
sha256 = import("crypto/sha256")
exec = import("os/exec")

log.Print("started")
r = rand.New(1)                       # seeded, so the sequence repeats
println(r.Intn(100), r.Intn(100))
println(utf8.RuneCount("caffè"), len("caffè"), utf8.Runes("naïve"))
println(csv.Parse("a,b\n\"c,d\",e", ","))
println(sha256.Hex("abc"))
println(exec.Output(["echo", "hi"]))
```

```
2026/07/29 11:10:59 started
11 56
5 6 [110, 97, 239, 118, 101]
[[a, b], [c,d, e]]
ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
hi
```

And the flags of a program, which read like Go's:

```python
flag = import("flag")
strings = import("strings")

n = flag.Int("n", 1, "how many times")
name = flag.String("name", "world", "who to greet")
loud = flag.Bool("loud", false, "shout")

rest = flag.Parse()
for i = 0; i < n.Value; i++ {
	msg = "hello, {name.Value}"
	println(if loud.Value { strings.ToUpper(msg) } else { msg })
}
println("rest:", rest)
```

```
$ tau greet.tau -n 2 -name tau -loud extra
HELLO, TAU
HELLO, TAU
rest: [extra]
```

## Builtin functions

These are always in scope, no import needed.

- `len(x)` -- the length of a String, List, Map or Bytes.
- `println(...)` -- print the arguments separated by a space, with a newline.
- `print(...)` -- the same without the newline.
- `input([prompt])` -- read a line from standard input, after an optional prompt.
- `string(x)` -- convert `x` to a String.
- `error(s)` -- build an error carrying the message `s`.
- `type(x)` -- the name of the type of `x`, as a String.
- `int(x[, bits])` -- convert `x` to an Integer; on a native value `bits` says
  how many of its bits are meaningful (8, 16, 32 or 64).
- `float(x)` -- convert `x` to a Float.
- `exit([code | message[, code]])` -- end the program, with an optional message
  and exit code.
- `append(xs, ...)` -- a new List with the arguments added at the end.
- `new()` -- a fresh empty object.
- `failed(x)` -- true if `x` is an error.
- `dlopen(path)` -- open a C shared object, or the program itself with `null`.
- `cfunc(sym, ret, args)` -- a C function with its types given as codes. What
  `ffi.Func` is made of, see [C libraries](#c-libraries).
- `cexport(fn, ret, args)` -- a tau function C can call, the same way round.
  What `ffi.Export` is made of.
- `pipe([n])` -- a new pipe, unbuffered or holding `n` values.
- `send(p, x)` -- send `x` to the pipe `p`.
- `recv(p)` -- take the next value out of the pipe `p`.
- `close(p)` -- close the pipe `p`.
- `hex(x)`, `oct(x)`, `bin(x)` -- the base 16, 8 and 2 spellings of `x`.
- `slice(x, start, end)` -- part of a String, List or Bytes.
- `keys(x)` -- the keys of a Map, or the field names of an object, as a List.
- `delete(m, k)` -- remove the key `k` from the Map `m`.
- `bytes(x)` -- Bytes from a String, a List of integers, or an empty buffer of
  `x` bytes; `bytes(ptr, n)` copies `n` bytes from a native pointer.
