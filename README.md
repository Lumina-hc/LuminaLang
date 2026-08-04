# LuminaLang v0.8.0

A strongly-typed compiled language written in Go. LuminaLang source (`.lum`)
is transpiled to Go and then built into a native executable, giving you
compiler performance with a clean, readable syntax.

## Features

- Strong static typing with three core types: `i32`, `bool`, `string`
- Full control flow: `if` / `elif` / `else`, `while`, `for ... in`
- Functions with parameters and return values, including recursion
- Terminal input and output built in
- Comments: `//` line comments and `/* ... */` block comments
- A small standard library (`use` imports) covering math, bits, io, time, and os
- Compiles to a single native executable with zero runtime dependencies

## Requirements

- Go 1.21 or later

## Building the compiler

```bash
go build -o luminalang.exe .
```

## Usage

```bash
luminalang <file.lum> [-o output]
```

Compile `file.lum` into an executable. By default the output executable
shares the name of the source file. Use `-o` to pick a different name.

```bash
luminalang examples/test_bool_literals.lum
./test_bool_literals.exe

luminalang examples/test_bool_literals.lum -o greeting.exe
./greeting.exe
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

`out` prints a value.

```lum
fn main() i32 {
    out "Hello, World!"
    out 3 + 5
    out true
    back 0
}
```

### Variables and types

`let name: type = value`. The three types are `i32`, `bool`, and `string`.

```lum
fn main() i32 {
    let count: i32 = 10
    let is_ready: bool = true
    let name: string = "Lumina"
    count = count + 1
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

`for ... in` iterates over a string, binding each character in turn.

### Control keywords

`halt` stops the program immediately; `skip` jumps to the next loop iteration.

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

### Input

```lum
use io::readline

fn main() i32 {
    out "What is your name?"
    let name: string = readline()
    out "Hello, " + name
    back 0
}
```

### Comments

```lum
// This is a line comment.

/*
This is a
block comment.
*/
```

### Imports

The standard library is organized into modules. Import individual functions
with `use module::function`, then call them directly.

```lum
use os::getenv

fn main() i32 {
    let path: string = getenv("PATH")
    out path
    back 0
}
```

## Standard library

Import individual functions with `use <module>::<function>`.

| Module | Functions |
| ------ | --------- |
| `math` | `abs`, `max`, `min`, `pow`, `sqrt`, `gcd`, `lcm`, `mod`, `clamp`, `sign` |
| `bits` | `and`, `or`, `xor`, `not`, `shl`, `shr`, `bit`, `setbit`, `clrbit`, `popcount` |
| `io` | `print`, `println`, `eprint`, `eprintln`, `readint`, `readline`, `flush` |
| `time` | `now`, `year`, `month`, `day`, `hour`, `minute`, `second`, `weekday`, `sleep_ms` |
| `os` | `exit`, `pid`, `ppid`, `args_count`, `sep`, `is_windows`, `getenv`, `cwd` |

## Project structure

```
LuminaLang-v0.8.0/
├── main.go              CLI entry point
├── go.mod
├── token/               Token definitions
├── lexer/               Lexer (tokenizer)
├── ast/                 Abstract syntax tree nodes
├── parser/              Parser (source to AST)
├── compiler/            Code generator and helpers
│   ├── compiler.go      Compiler entry point, import resolution
│   ├── types.go         Type definitions and builtin signatures
│   ├── codegen.go       AST to Go source generation
│   ├── infer.go         Type inference
│   └── runtime.go       Embedded Go standard library source
├── examples/            Example programs and test cases
├── website/             Project website (HTML/CSS)
├── LICENSE              Apache License 2.0
└── README.md
```

## Examples

The `examples/` directory contains runnable programs and regression tests.
Compile and run any of them, for example:

```bash
luminalang examples/test_bool_literals.lum
./test_bool_literals.exe
```

## License

LuminaLang is licensed under the [Apache License 2.0](LICENSE).
