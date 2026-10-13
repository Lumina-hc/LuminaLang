# LuminaLang v1.2.0 (LLVM backend)

A small strongly-typed compiled language. LuminaLang source (`.lum`) is
translated to LLVM IR and compiled to a native executable with `clang` — there
is no interpreter and no Go runtime in the produced binary. Programs link only
against a tiny C runtime, so a "hello world" is roughly 100 KB instead of the
~2.4 MB a Go build would produce.

## Features

- Strong static typing with four core types: `i32`, `f64`, `bool`, `string`
- Dynamically-sized slices (`[]i32`, `[]f64`, `[]bool`, `[]string`) with
  literals, indexing, `len` and `push`/`append`
- Block-scoped variables with shadowing
- Full control flow: `if` / `elif` / `else`, `while`, `for ... in`
- Arithmetic (integer and floating-point), comparison, logical and bitwise operators
- Functions with parameters and return values, including recursion
- Modules: the standard library plus user modules in sibling `.lum` files
- A standard library written in LuminaLang itself (`math`, `bits`, `strings`,
  `strconv`, `collections`, `encoding`, `json`, `os`, `time`, `file`, `fs`,
  `random`, `process`, `io`, `win`)
  on top of a thin `sys` C layer
- A decentralised package manager (`pho`) with a manifest, lockfile and
  precompiled `.lumc` libraries
- Diagnostics with `line:column` positions
- Compiles to a single native executable that depends only on the OS runtime

## Requirements

- [LLVM](https://releases.llvm.org/) / `clang` on your `PATH`
  (set `LUMINA_CC` to override the `clang` executable)

Building the compiler itself needs Go 1.21+.

## Building the compiler

```bash
go build -o luminalang.exe .
```

## Usage

```bash
luminalang <file.lum> [options]
```

| Option | Description |
| ------ | ----------- |
| `-o <path>` | output executable path |
| `-S`, `--emit-llvm` | write LLVM IR instead of compiling |
| `-c <module.lum>` | compile a module to a `.lumc` library |
| `--build-lib` | compile every stdlib `.lum` module to `.lumc` |
| `-I <dir>` | add a module search directory (repeatable) |
| `-O <level>` | clang optimisation level `0`–`3` (default `2`) |
| `--version` | print the version |
| `-h`, `--help` | print help |

```bash
luminalang examples/test_bool_literals.lum
./test_bool_literals.exe

luminalang examples/test_bool_literals.lum -o greeting.exe
./greeting.exe

luminalang -S examples/test_bool_literals.lum   # emit .ll IR
```

## Syntax overview

Every program is a set of functions. Execution starts at `main`, which
returns an `i32` exit code via `back`.

```lum
fn main() i32 {
    out "Hello, World!"
    back 0
}
```

### Output

`out` prints a value followed by a newline.

```lum
fn main() i32 {
    out "Hello, World!"
    out 3 + 5
    out true
    back 0
}
```

### Variables and types

`let name: type = value`. The types are `i32`, `f64`, `bool`, and `string`, plus
slices `[]i32`, `[]f64`, `[]bool`, and `[]string`.
Declarations are block-scoped: a name may be redeclared in a nested block
(shadowing), but not twice in the same block.

```lum
fn main() i32 {
    let count: i32 = 10
    let ratio: f64 = 3.5
    let is_ready: bool = true
    let name: string = "Lumina"
    count = count + 1
    back 0
}
```

### Operators

| Kind | Operators |
| ---- | --------- |
| Arithmetic (`i32`) | `+` `-` `*` `/` `%` |
| Arithmetic (`f64`) | `+` `-` `*` `/` (operands must both be `f64`) |
| Unary | `-x`, `~x` (bitwise not), `not x` |
| Comparison | `==` `!=` `<` `>` `<=` `>=` (yield `bool`) |
| Logical (`bool`) | `and`, `or` (short-circuit), `not` |
| Bitwise (`i32`) | `&` `|` `^` `~` `<<` `>>` (arithmetic shift) |
| String | `+` (concatenation) |

Parentheses group sub-expressions.

```lum
fn main() i32 {
    out 7 % 3
    out 6 & 3
    out 1 << 4
    out (1 < 2) and (3 > 2)
    out "a" + "b"
    back 0
}
```

### Conditionals

```lum
fn main() i32 {
    let score: i32 = 85
    if score >= 90 {
        out "Excellent"
    } elif score >= 60 {
        out "Pass"
    } else {
        out "Fail"
    }
    back 0
}
```

### Loops

```lum
fn main() i32 {
    let i: i32 = 0
    while i < 5 {
        out i
        i = i + 1
    }

    for ch in "hello" {
        out ch
    }
    back 0
}
```

`for ... in` iterates over a string (by Unicode code point), binding each
character in turn.

### Control keywords

Inside a loop `halt` breaks out and `skip` jumps to the next iteration.
Outside a loop `halt` exits the program.

```lum
fn main() i32 {
    let i: i32 = 0
    while i < 5 {
        i = i + 1
        if i == 3 {
            skip
        }
        out i
        if i == 5 {
            halt
        }
    }
    back 0
}
```

### Functions

Parameters are written as `name: type`; the return type follows the closing
parenthesis. Functions return a value with `back`.

```lum
fn add(a: i32, b: i32) i32 {
    back a + b
}

fn is_positive(n: i32) bool {
    back n > 0
}

fn main() i32 {
    out add(2, 3)
    out is_positive(5)
    back 0
}
```

### Slices

A slice is a dynamically-sized, ordered sequence of one element type, written
`[]i32`, `[]f64`, `[]bool` or `[]string`. Slices are heap-allocated and grow
automatically; a slice variable holds a reference, so reassign it after a `push`.

- `[a, b, c]` builds a slice literal. An empty literal `[]` needs an explicit
  type from its declaration or return type.
- `s[i]` reads an element (0-based) and `s[i] = v` writes one.
- `len(s)` returns the number of elements (also works on `string`).
- `push(s, v)` / `append(s, v)` return a new slice with `v` added; assign it
  back to the variable.
- Indexing outside `0 .. len(s)-1` is a runtime error that stops the program.

```lum
fn sum(a: []i32) i32 {
    let total: i32 = 0
    let i: i32 = 0
    while i < len(a) {
        total = total + a[i]
        i = i + 1
    }
    back total
}

fn main() i32 {
    let nums: []i32 = [1, 2, 3]
    nums[1] = 9
    nums = push(nums, 4)
    out len(nums)
    out sum(nums)
    back 0
}
```

### Imports

The standard library is organized into modules. Import individual functions
with `use module::function` (separate several with commas) and call them
directly. A function may be renamed locally with `as`, and `module::*` imports
everything the module exports.

```lum
use os::getenv
use io::readline, println
use strings::len as length

fn main() i32 {
    let path: string = getenv("PATH")
    out path
    out length("hello")
    back 0
}
```

#### User modules

A `use name::...` first looks for `name.lum` next to the file being compiled;
if none exists it falls back to the standard library. So a project can be split
across files:

```lum
// geometry.lum
fn area(w: i32, h: i32) i32 {
    back w * h
}
```

```lum
// main.lum
use geometry::area

fn main() i32 {
    out area(3, 4)
    back 0
}
```

Modules may import other modules (including each other's dependencies); a
module's `main` (if any) is ignored.

#### Compiled libraries (`.lumc`)

A module can be compiled ahead of time into a `.lumc` binary library. The
compiler reads only the embedded signature table and links the module's LLVM IR,
so downstream programs never re-parse the module's source.

```bash
luminalang -c geometry.lum -o geometry.lumc   # compile one module
luminalang --build-lib                        # compile every stdlib module
```

When resolving `use name::...`, the compiler looks for `name.lumc` before
`name.lum`, first next to the file being compiled, then in the standard library.
The shipped stdlib is already precompiled (`.lumc` alongside `.lum`), and the
`.lum` sources remain as the fallback and as the build input for `--build-lib`.

A `.lumc` file holds: the `LUMC` magic, a version, a JSON signature table
(module name, a hash of the source it was built from, its `use` dependencies,
and each exported function's parameter names/types and return type), and the
module's target-independent LLVM IR.

The stored source hash keeps libraries honest: if a `.lum` sits next to its
`.lumc` and no longer matches the hash, the compiler ignores the stale `.lumc`
and compiles from source instead (so editing a module always takes effect, even
without rebuilding). Run `-c`/`--build-lib` again to refresh the `.lumc`.

`-S` emits a single self-contained `.ll`: the compiled modules' IR is inlined
(with their string constants namespaced and duplicate declarations removed), so
the output links on its own against the stdlib C sources.

Each `-c`/`--build-lib` run also writes a sibling `name.lumi` declaration file
(a plain-text `module`/`fn` signature list, useful for editors and tooling).
When no `.lumc` or `.lum` is found, the compiler can load a module from its
`.lumi` alone: calls type-check against the declared signatures and expect an
external implementation at link time.

### Comments

```lum
// This is a line comment.

/*
This is a
block comment.
*/
```

## Standard library

Most modules are pure LuminaLang (`stdlib/lib/<module>.lum`). Modules that need
the operating system or C (`os`, `time`, `file`, `win`, `math` f64 helpers,
`encoding`, the reading half of `io`) are thin LuminaLang wrappers over
low-level `sys` C primitives.

```lum
use strings::upper
use math::sqrt
```

| Module | Functions |
| ------ | --------- |
| `math` | `abs`, `max`, `min`, `pow`, `sqrt`, `gcd`, `lcm`, `mod`, `clamp`, `sign`, `absdiff`, and f64: `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `atan2`, `sqrt_f`, `pow_f`, `fabs`, `ln`, `log10`, `log2`, `exp`, `floor`, `ceil`, `round`, `trunc`, `fmod`, `hypot`, `pi`, `e`, `tau` |
| `bits` | `bit`, `setbit`, `clrbit`, `toggle`, `popcount`, `parity` |
| `strings` | `len`, `len_bytes`, `char_at`, `upper`, `lower`, `trim`, `trim_left`, `trim_right`, `index`, `contains`, `startswith`, `endswith`, `repeat`, `substr`, `replace`, `reverse`, `count`, `split`, `split_lines`, `pad_left`, `pad_right` |
| `strconv` | `itoa`, `atoi`, `atof`, `ftoa`, `fstr`, `hex`, `bin`, `from_base`, `parse_hex`, `boolstr` |
| `collections` | `zeros`, `filled`, `sum`, `product`, `min`, `max`, `index_of`, `contains`, `count`, `reverse`, `sort`, `unique`, `range`, `sum_f`, `min_f`, `max_f`, `avg_f`, `index_of_s`, `contains_s`, `reverse_s`, `sort_s`, `join` |
| `encoding` | `md5`, `sha1`, `sha256`, `crc32`, `base64_encode`, `base64_decode` |
| `json` | `escape`, `unescape`, `quote`, `valid`, `kind`, `parse_int`, `parse_f64`, `parse_bool`, `num`, `numf`, `boolean`, `array_int`, `array_f64`, `array_bool`, `array_string` |
| `io` | `print`, `println`, `eprint`, `eprintln`, `write`, `newline`, `flush`, `readline`, `readint`, `readchar`, `readall` |
| `os` | `getenv`, `setenv`, `cwd`, `chdir`, `pid`, `ppid`, `args_count`, `arg`, `exit`, `sep`, `is_windows`, `hostname`, `temp_dir`, `username`, `arch`, `home_dir`, `is_admin`, `page_size`, `uname` |
| `time` | `now`, `now_ms`, `year`, `month`, `day`, `hour`, `minute`, `second`, `weekday`, `yearday`, `sleep_ms`, `unix_f`, `now_str`, `utc_offset` |
| `file` | `exists`, `read`, `write`, `append`, `remove`, `rename`, `size` |
| `fs` | `is_dir`, `is_file`, `mkdir`, `rmdir`, `list`, `join`, `abs` |
| `random` | `seed`, `int`, `float` |
| `process` | `system`, `exit`, `args_count`, `arg`, `getenv`, `args`, `exec`, `which` |
| `win` | `msgbox`, `beep`, `tick_ms`, `console_title`, `last_error`, `sleep_ms` |

Note: the bitwise helpers `and` / `or` / `xor` / `not` / `shl` / `shr` are now
language operators (`&` `|` `^` `~` `<<` `>>`) rather than functions, and
`strconv.bool` is available as `strconv.boolstr` (the former name is a keyword).

## Packages (photon)

`photon` is the package manager. It is decentralised: a package is just a git
repository (usually on GitHub), and there is no central registry. The command is
`pho`.

```bash
pho init myapp                     # create lumina.toml and src/main.lum
pho add github:user/geometry@v0.1.0  # add a dependency and install it
pho install                        # resolve and install everything
pho build                          # compile src/main.lum with luminalang
pho list                           # show declared dependencies
pho remove geometry                # drop a dependency
```

A dependency spec is one of `github:owner/repo[@ref]`, `owner/repo[@ref]`
(assumed GitHub), `https://host/owner/repo[@ref]`, `git@host:owner/repo[@ref]`,
or a local path. `@ref` is a tag, branch or commit; without it the repository's
default branch is used.

A package is one repository, laid out as:

```
geometry/                  # = one git repository
├── lumina.toml            # manifest
├── include/geometry.lumi  # declaration-only signatures
└── src/geometry.lum       # source
```

`lumina.toml`:

```toml
[package]
name = "geometry"
version = "0.1.0"
description = "2D geometry helpers"
license = "MIT"
authors = ["Ada"]

[lib]
modules = ["geometry"]

[dependencies]
util = "github:user/util@v1.0.0"
```

`pho install` fetches each dependency into a global cache (`~/.photon/cache`,
overridable with `PHOTON_HOME`), copies it into `packages/<name>/`, and writes
`photon.lock` pinning every dependency to an exact commit so builds are
reproducible. Transitive dependencies (a package's own `[dependencies]`) are
resolved recursively.

A dependency's identity is its name (the manifest key), and installed packages
live at `packages/<name>`, so names must be unique in a project. Two
repositories that share a name are not merged: adding a second source under an
already-used name is refused, and you give it an alias instead:

```bash
pho add github:alice/util
pho add github:bob/util --name util_b   # otherwise: "util already exists ... use --name"
```

If two different packages in the graph require the same name from different
sources, the install aborts with an explicit error rather than guessing.

`pho build` runs `luminalang` on `src/main.lum` with every installed package's
`src/`, `include/` and root directory on the module search path, so `use
geometry::area` finds the installed package. (The same search path can be set by
hand with `LUMINA_PATH` or `luminalang -I <dir>`.)

If you want to **publish** a package of your own, see the developer guide:
[PACKAGES.md](PACKAGES.md).

## Diagnostics

Errors are reported with a `line:column` prefix:

```
3:14: compile error: variable 'x' declared as i32 but initialised with string
5:4: syntax error, unexpected character '!'
```

## Testing

Golden tests live under `tests/`: each `tests/<name>.lum` (or
`tests/<name>/main.lum`) is compiled and run, and its output is compared with
the matching `.out` file. Programs under `tests/fail/` must fail to compile.

```bash
go test ./...
```

The test harness needs `clang` on `PATH`.

## Project structure

```
LuminaLang/
├── main.go              CLI entry point
├── go.mod
├── golden_test.go       Golden/regression test runner
├── token/               Token definitions
├── lexer/               Lexer (tokenizer)
├── ast/                 Abstract syntax tree nodes
├── parser/              Parser (source to AST)
├── compiler/            LLVM IR code generator
│   ├── compiler.go      Entry point, .lum/.lumc/.lumi module loader
│   ├── lumc.go          .lumc compiled-library format and builder
│   ├── decl.go          .lumi declaration file reader/writer
│   ├── types.go         Types, sys builtins, LLVM type mapping
│   ├── codegen.go       AST to LLVM IR generation
│   └── infer.go         Type inference
├── photon/              Package manager format and installer
│   ├── manifest.go      lumina.toml parsing/encoding
│   ├── spec.go          dependency spec parsing (github/URL/ssh/local)
│   ├── lock.go          photon.lock parsing/encoding
│   └── install.go       git fetch, cache, install into packages/
├── cmd/pho/             `pho` CLI (init/add/install/build/list/remove)
├── docs/                Documentation
│   └── index.html       GitHub Pages landing page
├── PACKAGES.md          Guide for third-party package authors (EN/中文)
├── stdlib/              Standard library
│   ├── lumina.toml      Package manifest for the standard library
│   ├── embed.go         Embeds lib/ for a self-contained binary
│   └── lib/
│       ├── *.lum        Modules written in LuminaLang
│       ├── *.lumc       Precompiled module libraries (--build-lib)
│       ├── *.lumi       Declaration-only signatures
│       ├── runtime.c    Entry point and core primitives
│       └── *.c          sys C layer (io, os, time, file, fs, random,
│                        process, win)
├── tests/               Golden tests and expected outputs
├── examples/            Example programs and regression tests
├── LICENSE              Apache License 2.0
└── README.md
```

## Building with a different clang

```bash
LUMINA_CC=/usr/bin/clang luminalang prog.lum
```

The standard library is embedded in the binary and extracted on demand, but an
on-disk `stdlib/lib` next to the compiler (or selected with `LUMINALANG_LIB`)
takes precedence — handy while developing the library itself.

## Examples

The `examples/` directory contains runnable programs and regression tests.
Compile and run any of them, for example:

```bash
luminalang examples/test_bool_literals.lum
./test_bool_literals.exe
```

## License

LuminaLang is licensed under the [Apache License 2.0](LICENSE).
