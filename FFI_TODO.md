# FFI: from two builtins to a library

The plan for the next change to the C interface: `plugin` and `native` stop
being builtins of the language and become an `ffi` module written in tau, on
top of a handful of primitives that stay in C because they cannot be anything
else.

Nothing here is done yet. This file is the specification and the checklist.

## Where we are

Today the whole C interface is two builtins and a piece of the dot operator.

| what | where |
| --- | --- |
| `plugin(path)` — dlopen, returns a handle as `obj_native` | `internal/obj/builtins.c:336`, search path in `internal/obj/plugin.h:41` |
| `lib.symbol` — dlsym on a handle, returns a pointer as `obj_native` | `internal/vm/vm.c:223` (`vm_exec_dot`) |
| `native(fn, "double(double, double)")` — signature, prepared cif | `internal/obj/builtins.c:361`, parser and call in `internal/obj/ffi.c` |
| the untyped call — every argument an int64 or a double, the result a word | `internal/vm/vm.c:769` and the marshalling below it |
| `bytes(ptr, n)` — copy `n` bytes out of a pointer | `internal/obj/builtins.c:633` |
| the plugins a bundle carries | `bundle.go:26` (`pluginRe`), `bundle.go:135` (`addPlugin`), `internal/rt/rt.c:96` (`write_plugins`) |

`internal/obj/ffi.c` is 474 lines and does two unrelated jobs. About 240 of
them are the mechanism: `ffi_prep_cif`, `ffi_call`, the conversion of tau
values into a `union word` and back, parking the collector around a call that
may block. The other 230 are notation: `code_of_name`, `code_of_decl`, the
split on commas, dropping the argument names and the qualifiers.

## What is wrong with it

1. **The notation is in C.** It is string handling — the part of the FFI most
   likely to grow a case (a type name nobody thought of, an attribute to
   ignore, a better error message) — and it sits in the file where a mistake
   is a segfault rather than an error value. In tau it is `strings.Split`, a
   map and about half the lines.
2. **`plugin` and `native` are builtins**, so they are in every program whether
   it opens a library or not, and they can never grow an option without
   growing the language.
3. **There is no memory to speak of.** `bytes(ptr, n)` copies out of a pointer
   and `bytes(n)` allocates a buffer tau owns. There is no `malloc`, no `free`,
   no way to write into a pointer a library handed back, no way to read a
   struct field other than by counting bytes by hand, no null check that says
   anything useful.
4. **A handle is never closed** on purpose: `dlclose` happens when the
   collector gets to the object, and there is no way to say "close it now".
   A failed open does report what `dlerror` said, but only for the last
   directory tried, so "no such file" is the message whatever the real reason
   was.
5. **A library name is a platform.** Every program that opens the C library
   writes `libc.so.6`, which is Linux and glibc: not musl, not macOS, not
   Windows. The place to fix that is a library, not a builtin.
6. **A signature can only be written, not built.** There is no way to assemble
   one from values, which is what a binding generator would want to do.

## The design

Decided: **`plugin` stays a builtin and grows**, and **`native` stays a builtin
and shrinks**. The module is written on top of both, not instead of them.

A library written in tau cannot open a shared object by itself: something has
to call `dlopen`, and that something can only be a builtin. Even a `dlopen`
that arrived from a shared object would need a shared object opened first. So
`plugin` keeps its name and its search path, and gets the things it should
have had from the start - see "What plugin should do" below.

`native` is the other way round. Of the three things it does, only one has to
be in C:

1. reading the signature, which is text handling and moves to `ffi.tau`;
2. preparing the call, `ffi_prep_cif` and an `obj_native_fn` the VM knows how
   to call, which is a VM object no C library can hand out;
3. marshalling, per call: a tau int into a four byte ABI slot, a tau string as
   a NUL terminated `char *`, the collector parked for the duration, the
   result read back out of the returned word. This one touches how values are
   represented, and a tau int has no address to hand to `memcpy` in the first
   place.

It could be done entirely in tau, over a system `libffi.so`, by building the
`ffi_type *` and argument arrays by hand with `malloc` and `memcpy`. It would
mean writing the size and the layout of `ffi_cif` into tau source, requiring a
shared libffi where it is static today, and spending a few dozen VM operations
on every call. Point 3 is the one thing in the FFI that has to be C, and it is
the only thing that stays.

**In C, the primitives.** Two, the same two there are today:

```
plugin(path [, flags])   -> a handle, or an error. Keeps the search path it
                            has, plus what is listed below.
native(fn, ret, [args])  -> a prepared call, the types given as integer codes,
                            never as text.
```

Everything else the FFI needs is already C, and we are in the business of
calling C. Memory is `malloc`, `free` and `memcpy` from the C library;
`dlsym` and `dlclose` are in the C library as well, and the dot operator on a
handle is the raw `dlsym` that bootstraps the rest. Which means no builtin for
any of it. Checked, and this runs today:

```tau
libc = plugin("libc.so.6")

malloc = native(libc.malloc, "void *malloc(size_t n)")
free   = native(libc.free, "void free(void *p)")
memcpy = native(libc.memcpy, "void *memcpy(void *dst, const void *src, size_t n)")

p = malloc(8)
memcpy(p, "abcdefg", 8)      # a tau string into memory C owns
println(string(bytes(p, 7))) # abcdefg
memcpy(p, bytes([65, 66, 67]), 3)
println(string(bytes(p, 3))) # ABC
free(p)

dlsym = native(libc.dlsym, "void *dlsym(void *handle, const char *name)")
println(type(dlsym(null, "strlen")))  # native
```

So `ffi.Alloc`, `ffi.Free`, `ffi.Write` and `lib.Close` are tau functions over
C functions, not new primitives. Reading stays `bytes(ptr, n)`, which is a
shape of a builtin that already exists.

**In tau, everything a person types.** `stdlib/ffi.tau`:

```tau
ffi = import("ffi")

libm = ffi.Open("libm")                      # the suffix of this system is added
pow = libm.Func("double pow(double, double)")
println(pow(2.0, 10.0))

libc = ffi.Open(ffi.LibC)                    # whatever the C library is here
snprintf = libc.Func("int snprintf(char *s, size_t n, const char *fmt, double x)")
buf = bytes(64)
n = snprintf(buf, 64, "pi is %.3f", 3.14159)
println(string(slice(buf, 0, n)))

libm.Close()
```

The signature stays exactly what it is now — a C declaration, names and
qualifiers optional — only the parser moves. `ffi.Func` turns the text into
`native(fn, ret_code, arg_codes)` and everything below that is unchanged.

### What plugin should do

Everything here is a thing it cannot do today.

- **`plugin(null)`** gives the handle of the program itself, `dlopen(NULL)`.
  That is what makes `dlsym`, `dlclose`, `malloc` and the rest reachable as
  ordinary C functions, so that none of them needs a builtin of its own.
- **A bare name resolves per system.** `plugin("m")` and `plugin("libm")` try
  `libm.so`, `libm.so.6`, `libm.dylib`, `m.dll` and so on, so that a program
  that opens the maths library is not a Linux program. A name that looks like
  a path (it has a separator, or an extension already) is used as it is, which
  is what keeps the bundler working: the file inside a bundle is stored under
  the name written in the source.
- **The error says what was tried.** Today the message is whatever `dlerror`
  said about the last directory attempted, so a library that exists but pulls
  in a missing dependency reports "no such file". The error should carry the
  paths tried and the reason each one failed.
- **Flags.** `plugin(path, flags)` for `RTLD_NOW` against the default lazy
  binding, and `RTLD_GLOBAL` for a library whose symbols the next one needs.
  The constants come from `ffi.tau` - `ffi.Now`, `ffi.Global` - so the builtin
  takes an integer and nothing else.
- **Opening the same library twice** returns the same handle from the loader
  anyway; what should not happen twice is the `dlclose` at collection. Whoever
  writes `lib.Close()` has to mark the handle closed.

What it should *not* grow is a method: the dot on a handle is `dlsym`, so a
field named `Close` would be a symbol named `Close`. The conveniences live on
the object `ffi.Open` returns, which wraps the handle rather than being one.

### The surface of the module

```
ffi.Open(name)          an object for a shared library. The name may be a path,
                        or a bare name, in which case the suffix of the system
                        (.so, .dylib, .dll) and the lib prefix are tried.
ffi.LibC, ffi.LibM      the name of the C and math libraries on this system.
lib.Sym(name)           the raw pointer of a symbol, error when it is not there.
lib.Func(signature)     a callable with the types the signature says.
lib.Var(name, type)     read a variable of the library: Var("errno", ffi.Int).
lib.Close()             dlclose now rather than whenever the collector runs.

ffi.Signature(text)     the parse on its own: [ret_code, [arg_codes]], so that
                        a generator can build one without going through text.
ffi.Alloc(n)            memory C owns, freed by ffi.Free, not by the collector.
ffi.Free(ptr)           Both are malloc and free of the C library, wrapped.
ffi.Read(ptr, n)        bytes out of a pointer, which is bytes(ptr, n).
ffi.Write(ptr, data)    bytes into one, which is memcpy.
ffi.String(ptr)         a C string read up to its NUL, which is strlen and a read.
ffi.Int8 … ffi.UInt64, ffi.Float32, ffi.Float64, ffi.Pointer, ffi.String_, ffi.Void
                        the type codes, for building a signature by hand.
```

Struct reading stays what it is for now — `ffi.Read` plus `int(x, bits)` — and
a struct helper is a separate question, not part of this change.

## What to change

### The C side

1. `internal/obj/ffi.c` — delete the notation: `code_of_name`, `int_code`,
   `is_qualifier`, `code_of_decl` and the text half of `new_native_obj`, about
   230 lines. Keep `type_of`, the marshalling, `native_call`, `dispose_native_obj`
   and `native_str`.
2. `new_native_obj(void *fn, char ret, char *codes, size_t n)` — same function
   without the parsing, taking the codes it used to work out.
3. `internal/obj/builtins.c:361` — `native_b` takes `(fn, ret_code, [arg_codes])`:
   an integer and a list of integers, no string anywhere. Validate the codes and
   say which one is wrong.
4. `internal/obj/builtins.c:336` — `plugin_b` keeps its name and grows what
   the section above lists: `null` for the program itself, a bare name
   resolved per system, an error that says what was tried, and an optional
   flags argument. The candidate names and the search path stay in
   `internal/obj/plugin.h`, which is where they are now.
5. `internal/vm/vm.c:223` — the `dlsym` in `vm_exec_dot` stays as it is. It is
   how a symbol is read today, it is what the untyped call uses, and it is the
   one primitive that lets everything else — `dlsym` itself, `dlclose`,
   `malloc` — be reached as ordinary C functions.
6. No builtins for memory. `malloc`, `free` and `memcpy` come from the C
   library through the FFI, and reading stays `bytes(ptr, n)`.
7. `internal/obj/object.go` — nothing to add and nothing to move. The index of
   a builtin is what a compiled `.tauc` holds, so the array stays exactly as
   it is; `plugin` and `native` keep their places and only their arguments
   change.

### The tau side

8. `stdlib/ffi.tau` — new. The parser (about 120 lines), `Open`, `Func`, `Sym`,
   `Var`, `Close`, the type constants, the memory helpers, the platform names.
   It may import `strings` and `runtime`, and must not import `syscall`:
   `syscall` will import it.
9. `stdlib/ffi_test.tau` — new, mirroring `tests/ffi_test.tau`: every type of
   the table, the optional names and qualifiers, the errors. It needs a shared
   object to call, so either it uses the one in `tests/ffi` or it grows one of
   its own under `stdlib/ffi/`.
10. `stdlib/syscall.tau:7`, `stdlib/math.tau:9`, `stdlib/runtime.tau:7` — from
    `plugin("x/lib.so")` to `ffi.Open(...)`. They use the untyped call today;
    moving them to `Func` with real signatures is a good idea and a separate
    commit, since it touches every call in `syscall.tau`.

### The bundler and the runtime

11. `bundle.go:26` — `pluginRe` looks for `plugin("...")` in the source to
    decide which shared objects to carry. It has to look for the new spelling
    as well, `ffi.Open("...")`, or a bundled program will not find its library.
    The regexp is over the source and knows nothing about names, so a program
    that renames the import (`f = import("ffi")`) escapes it: either match the
    call by its shape (`.Open("...")`) and accept the false positives, or walk
    the syntax tree, which is the honest fix and a bigger one.
12. `bundle.go:135` (`addPlugin`) and `internal/rt/rt.c:96` (`write_plugins`)
    keep working unchanged as long as the name in the source is the name of the
    file inside the bundle, which is why `ffi.Open` must pass the name through
    untouched when it looks like a path.
13. `ffi.Open` adds a suffix per system, but a bundle is built on one machine
    and may be unpacked on another, and the file inside it is stored under the
    name written in the source. Decide: either the bundle stores the resolved
    name, or `Open` tries the written name first, which is what the loader does
    today.

### Documentation and tests

14. `tests/ffi_test.tau` — rewrite against `ffi.Func`, and keep two cases on
    the raw `native` so the primitive stays covered.
15. `README.md` — the "C libraries" section: `ffi.Open`/`Func` instead of
    `plugin`/`native`, the type table stays as it is.
16. The website, `~/Documenti/tau-website`: `src/tooling.md` ("Plugins" and
    "Types across the boundary"), `src/stdlib.md` (a section for `ffi`),
    `samples/ffi_native.tau`.
17. `internal/format` — nothing to do, `ffi.Open` is an ordinary call.

## The order to do it in

1. `stdlib/ffi.tau` with the parser, calling the `native` that exists today,
   plus its tests. Nothing breaks, the new way works alongside the old one.
2. Move the notation out of `internal/obj/ffi.c` and change `native` to take
   codes. `stdlib/ffi.tau` is the only caller, so this is where the old
   spelling stops working.
3. `dlopen`/`dlsym`/`dlclose` and the memory builtins, then `ffi.Open` on top.
4. The bundler regexp, with a test that bundles a program using `ffi.Open` and
   runs it where the library is not installed.
5. The stdlib modules, one commit each.
6. `plugin` grows the flags, the null handle and the per system names. Nothing
   breaks here: every call that works today keeps working, and `ffi.Open` is
   built on top of it.

## Open questions

- **Two ways to open a library.** `plugin` stays, so both `plugin("libm.so")`
  and `ffi.Open("m")` will work. That is deliberate - the first is the
  primitive, the second is the one with the manners - but it is the kind of
  thing that ends up documented twice and explained badly. The stdlib should
  use `ffi.Open` everywhere, and the README should show `plugin` only where it
  is explaining what `ffi.Open` is made of.
- **Callbacks.** A C function that takes a function pointer cannot be called at
  all today. `ffi_closure` would fix it, and it needs a story for a C thread
  entering the VM. Out of scope here, but the shape of `ffi.Func` should not
  make it harder later: a callback is `ffi.Callback(signature, tau_function)`.
- **Structs by value**, in and out. libffi does them with a type built at
  runtime. `ffi.Struct([types])` is the natural spelling, and it is a change of
  its own.
- **Windows.** `plugin_open` splits `TAUPATH` on `:` and reads `HOME`, neither
  of which is right there. Whoever writes `ffi.Open` should decide whether the
  search path moves into tau as well, which would be the last piece of policy
  leaving the C side.
