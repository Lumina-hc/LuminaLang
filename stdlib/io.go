package stdlib

import (
    "fmt"
    "luminalang/runtime"
    "os"
    "strconv"
)

var IOModule = Module{
    Name: "io",
    Funcs: map[string]runtime.BuiltinFn{
        "println": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 1 {
                return 0, errArity("io::println", 1, len(args))
            }
            fmt.Println(args[0])
            return 0, nil
        },
        "print": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 1 {
                return 0, errArity("io::print", 1, len(args))
            }
            fmt.Print(args[0])
            return 0, nil
        },
        "readint": func(args []runtime.Value) (runtime.Value, error) {
            var s string
            _, err := fmt.Scan(&s)
            if err != nil {
                os.Exit(0)
            }
            val, err := strconv.ParseInt(s, 10, 64)
            if err != nil {
                return 0, errMsg("io::readint: input must be an integer")
            }
            return val, nil
        },
        "eprint": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 1 {
                return 0, errArity("io::eprint", 1, len(args))
            }
            fmt.Fprint(os.Stderr, args[0])
            return 0, nil
        },
        "eprintln": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 1 {
                return 0, errArity("io::eprintln", 1, len(args))
            }
            fmt.Fprintln(os.Stderr, args[0])
            return 0, nil
        },
        "flush": func(args []runtime.Value) (runtime.Value, error) {
            return 0, nil
        },
    },
}
