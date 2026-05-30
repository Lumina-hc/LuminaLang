package stdlib

import (
    "luminalang/runtime"
    "os"
)

var OSModule = Module{
    Name: "os",
    Funcs: map[string]runtime.BuiltinFn{
        "exit": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 1 {
                return 0, errArity("os::exit", 1, len(args))
            }
            os.Exit(int(args[0]))
            return 0, nil
        },
        "getenv": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 1 {
                return 0, errArity("os::getenv", 1, len(args))
            }
            return 0, nil
        },
        "pid": func(args []runtime.Value) (runtime.Value, error) {
            return runtime.Value(os.Getpid()), nil
        },
        "ppid": func(args []runtime.Value) (runtime.Value, error) {
            return runtime.Value(os.Getppid()), nil
        },
        "args_count": func(args []runtime.Value) (runtime.Value, error) {
            return runtime.Value(len(os.Args)), nil
        },
        "cwd": func(args []runtime.Value) (runtime.Value, error) {
            return 0, nil
        },
        "sep": func(args []runtime.Value) (runtime.Value, error) {
            return runtime.Value(os.PathSeparator), nil
        },
        "is_windows": func(args []runtime.Value) (runtime.Value, error) {
            if os.PathSeparator == '\\' {
                return 1, nil
            }
            return 0, nil
        },
    },
}
