# FFI: two layers, one seam

The plan for the C interface. Nothing here is done yet; this file is the
design and the checklist.

## The shape

Two ways to reach C, on purpose, and the second is built out of the first.

**Layer 1, raw.** `dlopen` opens a shared object, the dot on a handle is
`dlsym`, and a symbol can be called straight away with no declaration at all:
the arguments go as int64 or double, the result comes back as a machine word
that you decode yourself with `int(x, bits)` or `float(x, 64)`. Three lines to
try a library, and it will lie to you if you get the types wrong. That is the
deal, and the documentation has to say it in those words.

```tau
libm = dlopen("m")
println(float(libm.pow(2.0, 10.0), 64))
```

**Layer 2, declared.** `stdlib/ffi.tau` gives a symbol a signature written as a
C declaration. Same `dlopen`, same dot, one more step:

```tau
ffi  = import("ffi")
libm = dlopen("m")

pow = ffi.Func(libm.pow, "double pow(double, double)")
println(pow(2.0, 10.0))
```

The two are not two ways of doing the same thing, the way `syscall` and `os`
are not. Layer 2 has no `dlopen` of its own, no symbol lookup of its own and no
search path of its own: it takes what layer 1 hands it and adds types.

### The two rules

They are what keeps this a layer instead of a mess.

1. **Layer 1 never learns a job of layer 2.** No C declarations parsed in
   `builtins.c`, no signature strings in a builtin. The primitives take numbers
   and pointers.
2. **Layer 2 never goes around layer 1.** `ffi.tau` calls `dlopen`, uses the
   dot for symbols, and gets its memory from the C library through layer 1.

A corollary worth keeping: `ffi.tau` is the proof that layer 1 is complete. If
writing it turns out to need a new builtin, layer 1 is missing something.

## Layer 1: what it is, what it gains

Today, all of it:

| what | where |
| --- | --- |
| `dlopen(path)` — the library, with a search path (`plugin` until the rename) | `internal/obj/builtins.c:336`, `internal/obj/plugin.h:41` |
| `lib.symbol` — dlsym, gives an `obj_native` pointer | `internal/vm/vm.c:223` (`vm_exec_dot`) |
| the untyped call — arguments guessed, result a word | `internal/vm/vm.c:769` |
| `bytes(ptr, n)` — copy `n` bytes out of a pointer | `internal/obj/builtins.c:633` |

What it should gain, all of it in `dlopen`:

- **`dlopen(null)`** — the handle of the program itself. This is
  the one that matters: it makes `malloc`, `free`, `memcpy`, `strlen`, `dlsym`
  and `dlclose` reachable as ordinary C functions, so layer 2 needs no builtin
  for memory and none for closing a library.
- **A bare name resolved per system.** `dlopen("m")` tries `libm.so`,
  `libm.so.6`, `libm.dylib`, `m.dll`. A name that already looks like a path — it
  has a separator or an extension — is used as it is, which is what keeps the
  bundler working, since a bundle stores a shared object under the name written
  in the source.
- **An error that says what was tried.** Today the message is whatever
  `dlerror` said about the last directory attempted, so a library that exists
  but pulls in a missing dependency reports "no such file". List the paths and
  the reason each one failed.
- **Flags**, `dlopen(path, flags)` for `RTLD_NOW` and `RTLD_GLOBAL`, with the
  constants in `ffi.tau` so that the builtin takes an integer. Only if
  something actually needs them.

What it must *not* gain: methods. The dot on a handle is `dlsym`, so a field
named `Close` would be a symbol named `Close`. Conveniences belong to layer 2.

## The seam: `cfunc`

The second primitive of layer 1, and the only thing in the whole FFI that
cannot be moved anywhere else:

```
cfunc(sym, ret_code, [arg_codes]) -> a callable
```

It prepares an `ffi_cif` once and returns an `obj_native_fn` the VM knows how
to call. No text: the codes are integers, and `ffi.tau` is what produces them.

It is public, like the rest of layer 1, and it needs no underscore to keep
people away: writing `cfunc(sym, 12, [12, 12])` by hand is unattractive enough
on its own, and anyone who does it is in layer 1, where the deal is already
that you know what you are doing. `ffi.Func` is the same thing with the types
written in C, and the name says so.

**It cannot be moved.** Every alternative was looked at:

- *Take it out of the builtins and hand it only to the stdlib.* Builtins are
  resolved by name at compile time, so `ffi.tau` would lose it too: it is
  ordinary tau source, not a privileged module. Resolving it only for files
  under the stdlib would be policy about file paths inside the compiler, broken
  for a vendored stdlib and for a bundle compiled elsewhere.
- *Ship it as a shared object of the stdlib.* A `.so` can only hand tau back a
  machine word: the untyped path wraps whatever comes back into `obj_native`
  (`vm.c:902`). Writing the object into a buffer instead only moves the
  problem, since turning bytes into an object would need a builtin far more
  dangerous than this one — it would let you forge any value in the VM.
- *Do the whole thing in tau over a system `libffi.so`.* `ffi_prep_cif` and
  `ffi_call` are C functions, and the `ffi_type` globals are data symbols the
  dot can fetch. It dies three times over: `sizeof(ffi_cif)` and libffi's
  internal layout would have to be written into tau source and are not stable
  across versions; nothing in libc is "store this double at this address", so
  an argument slot cannot be filled from tau; and a forwarding closure would
  need varargs, which tau does not have.

Hiding it would be theatre in any case. The untyped call of layer 1 is public,
documented and about to be improved, and it is strictly the more dangerous of
the two: it calls an arbitrary pointer with guessed types, where `cfunc` at
least checks the arity and the argument types and returns an error instead of
walking off the stack.

## Layer 2: `stdlib/ffi.tau`

```
ffi.Func(sym, signature)    a callable with the types the signature says.
ffi.Bind(lib, [signatures]) an object with one method per signature, named
                            after the function.
ffi.Sig(text)               the parse on its own: [ret_code, [arg_codes]], for
                            building a call without going through text.
ffi.Void, ffi.Bool, ffi.Int8 … ffi.UInt64, ffi.Float32, ffi.Float64,
ffi.Pointer, ffi.CString    the codes.

ffi.Alloc(n) / ffi.Free(p)  malloc and free of the C library.
ffi.Write(p, data)          memcpy into memory C owns.
ffi.String(p)               a C string read up to its NUL.
ffi.Read(p, n)              which is bytes(p, n), here so the module reads as
                            one thing.
ffi.Now, ffi.Global         the flags for dlopen, if dlopen grows them.
```

The signature is what it already is today, a C declaration: the name of the
function and the names of the arguments optional, `const`/`volatile`/`restrict`
read and ignored, the widths C leaves to the machine taken from the machine,
the exact width names of `stdint.h` and the tau spellings (`uint64`, `float64`)
accepted. The parser moves out of C with its behaviour unchanged, and the cases
in `tests/ffi_test.tau` are what says so.

The memory helpers are four calls into libc through layer 1, opened once with
`dlopen(null)`. They are in the module because otherwise everyone writes the
same three lines, not because they are primitives.

## What to change

### C

1. `internal/obj/ffi.c` — delete the notation: `code_of_name`, `int_code`,
   `is_qualifier`, `code_of_decl` and the text half of `new_native_obj`, about
   230 lines of the 474. Keep `type_of`, the marshalling, `native_call`,
   `dispose_native_obj` and `native_str`.
2. `new_native_obj(void *fn, char ret, const char *codes, size_t n)` — the same
   function without the parsing.
3. `internal/obj/builtins.c:361` — `native_b` becomes `cfunc_b`, taking
   `(sym, int, [int])` and saying which code it does not know. It keeps the
   check that passes an error argument straight through, which is what makes a
   failed symbol lookup readable.
4. `internal/obj/builtins.c:336` — `plugin_b` becomes `dlopen_b` and grows the
   null handle, the per system names, the better error and the optional flags.
   The search path stays where it is, in `internal/obj/plugin.h`, which should
   be renamed with it.
5. `internal/obj/object.go` — in `Builtins`, `plugin` becomes `dlopen` and
   `native` becomes `cfunc`, both **renamed in place**. The index of a builtin
   is what a compiled `.tauc` holds, so nothing may be inserted or moved: a
   bundle built before the rename keeps running afterwards.
6. `internal/vm/vm.c:223` and `:769` — untouched. The dot is `dlsym` and the
   untyped call is layer 1, which stays exactly as it is.

### tau

7. `stdlib/ffi.tau` — new. The signature parser (about 120 lines), `Func`,
   `Bind`, `Sig`, the codes, the memory helpers. It may import `strings`, which
   pulls only `buffer`; it must not import `syscall`, `os`, or anything else
   that opens a library of its own.
8. `stdlib/ffi_test.tau` — new: the cases of `tests/ffi_test.tau` rewritten
   against `ffi.Func`, plus the ones that only exist here — a signature that
   does not parse, `Bind` over several functions, an `Alloc`/`Write`/`String`
   round trip. It needs a shared object to call: either `tests/ffi/libffitest.so`
   or one of its own under `stdlib/ffi/`.
9. `stdlib/math.tau` — the first module to move to layer 2, because it is ten
   lines and every one of them ends in `float(x, 64)` today. The smallest proof
   that the layer works.
10. `stdlib/syscall.tau` and `stdlib/runtime.tau` — the same, later, one commit
    each. `syscall.tau` is 485 lines and about sixty calls, and the signatures
    have to be read off `stdlib/syscall/syscall.c` one at a time. Optional:
    layer 1 keeps working, so this is cleanup and not migration.

### Every place the name is written

The rename is not two lines in `object.go`: `plugin` is written in the source
of the standard library, of the tests, of the examples, of the documentation
and of the three editor vocabularies that colour it. All of it moves in the
same commit as the builtin, or the tree is left saying two different things.

11. **The calls.** Every one of them becomes `dlopen`:
    - `stdlib/math.tau:9`, `stdlib/runtime.tau:7`, `stdlib/syscall.tau:7` and
      the comment above it at `:5`;
    - `tests/ffi_test.tau:5`, `tests/test_plugin.tau:2`, `tests/test_module.tau:2`;
    - `examples/plugins.tau:1`, and the directory `examples/plugin/` it opens,
      which can keep its name or become `examples/clib/` — it is a directory,
      not a call.
12. **The lists of builtin names**, which is where a rename is forgotten and
    then shows up as a word that stops being coloured:
    - `internal/obj/object.go`, `Builtins` — renamed in place, see above;
    - `cmd/tau-lsp/builtins.go:27` — the entry and its one line description,
      plus a new one for `cfunc`;
    - `~/Documenti/tree-sitter-tau/queries/highlights.scm:61` — the
      `@function.builtin` alternation, and the same file copied into
      `~/Documenti/tau-zed/languages/tau/`, which needs the grammar rev bumped
      in `extension.toml` to ship;
    - `~/Documenti/tau-website/build.py:31` — the `BUILTINS` set the site
      highlights with.
13. **The prose.** `README.md:546`, `:580`, `:630`, `:783` (the builtin list),
    and on the site `src/index.md` (the "C without a wrapper" card, which still
    shows the letter signature as well), `src/tooling.md` and
    `samples/ffi_native.tau`.
14. **The internal vocabulary**, which is a decision rather than a rename. The
    bundle format carries what it calls plugins: `internal/bundle/codec.go`,
    `internal/bundle/bundle.go`, `internal/vm/vm.go`, `internal/rt/rt.c` (14
    mentions) and `internal/obj/utils.c`. None of it is user facing, and a
    shared object packed inside a bundle is a plugin of that bundle in a way it
    never was of the language, so the honest move is to leave the format alone
    and rename only what the language shows: `internal/obj/plugin.h` to `dl.h`
    and `plugin_open` to `dl_open`.

### Everything else

15. `bundle.go:26` — `pluginRe` looks for `plugin("...")` in the source to
    decide which shared objects a bundle carries. It has to look for
    `dlopen("...")` instead, and it has to change in the same commit as the
    builtin or every bundled program stops finding its library. Layer 2 adds
    nothing to look for, since it opens nothing of its own. The other thing to
    keep true: the per system names must never change the name a bundle stores,
    which is why a name that looks like a path is passed through untouched.
16. `README.md` — the "C libraries" section becomes two: layer 1 as the quick
    and unsafe way, with `int(x, bits)` shown and the trade said plainly, and
    layer 2 as the one to use. `cfunc` is shown once, where `ffi.Func` is
    explained, and not in the list of things to reach for.
17. The website, `~/Documenti/tau-website`: `src/tooling.md` (the "Plugins"
    section, same split), `src/stdlib.md` (a section for `ffi`), and
    `samples/ffi_native.tau`.

## The order

Every step leaves the tree working.

1. `stdlib/ffi.tau` and its tests, calling the `native` that exists today. The
   new way is available, nothing has changed.
2. Move the parser out of `internal/obj/ffi.c`; `native` becomes `cfunc` and
   takes codes. `ffi.tau` is the only caller, so this is where the old spelling
   stops working — and the old spelling is one day old and in no release.
3. `plugin` becomes `dlopen` everywhere at once — the builtin, the bundler
   regexp, the stdlib, the tests, the examples, the LSP list, the two
   highlighting queries, the site — in a single commit, since a half done
   rename is a tree that contradicts itself. This is the breaking one: it goes
   in a release of its own and in the changelog, `plugin` being what every
   program that opens a library has written until now.
4. `dlopen(null)`, then the memory helpers in `ffi.tau` on top of it, then the
   per system names and the better error.
5. `math.tau` to layer 2, tests green.
6. `syscall.tau` and `runtime.tau`, one commit each, if and when it is worth
   it.
7. README and website, once the shape has stopped moving.

## Open questions

- **Flags.** `RTLD_NOW` and `RTLD_GLOBAL` are two lines in `dlopen_b` and a
  constant each in `ffi.tau`. Nothing needs them yet. Add them when something
  does.
- **Callbacks.** A C function that takes a function pointer cannot be called at
  all, in either layer. `ffi_closure` would fix it, and it needs a story for a C
  thread entering the VM: which routine it runs on, and what the collector does
  meanwhile. Out of scope here, but `ffi.Callback(signature, f)` is the spelling
  it should get, so nothing in `ffi.Func` should make it harder.
- **Structs by value**, in and out. libffi builds a type at runtime for them,
  and `ffi.Struct([codes])` is the natural spelling. A change of its own; today
  a struct is read with `bytes(ptr, n)` and written with `ffi.Write`.
- **Windows.** `plugin_open` (`dl_open` after the rename) splits `TAUPATH` on
  `:` and reads `HOME`, neither
  of which is right there. Whoever writes the per system names will be looking
  at exactly that code.
