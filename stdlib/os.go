package stdlib

import (
    "fmt"
    "luminalang/runtime"
    goruntime "runtime"
    "os"
)

func registerOS(interp *runtime.Interpreter) {
    interp.RegisterBuiltin("os.exit", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 1 {
            return runtime.VInt(0), fmt.Errorf("exit: expected 1 argument")
        }
        os.Exit(int(args[0].Int()))
        return runtime.VInt(0), nil
    })
    interp.RegisterBuiltin("os.pid", func(args []runtime.Value) (runtime.Value, error) {
        return runtime.VInt(int64(os.Getpid())), nil
    })
    interp.RegisterBuiltin("os.ppid", func(args []runtime.Value) (runtime.Value, error) {
        return runtime.VInt(int64(os.Getppid())), nil
    })
    interp.RegisterBuiltin("os.args_count", func(args []runtime.Value) (runtime.Value, error) {
        return runtime.VInt(int64(len(os.Args))), nil
    })
    interp.RegisterBuiltin("os.sep", func(args []runtime.Value) (runtime.Value, error) {
        if goruntime.GOOS == "windows" {
            return runtime.VInt(int64('\\')), nil
        }
        return runtime.VInt(int64('/')), nil
    })
    interp.RegisterBuiltin("os.is_windows", func(args []runtime.Value) (runtime.Value, error) {
        if goruntime.GOOS == "windows" {
            return runtime.VInt(1), nil
        }
        return runtime.VInt(0), nil
    })
    interp.RegisterBuiltin("os.getenv", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 1 {
            return runtime.VInt(0), fmt.Errorf("os.getenv: expected 1 argument")
        }
        key := args[0].Str()
        val := os.Getenv(key)
        return runtime.VStr(val), nil
    })
    interp.RegisterBuiltin("os.cwd", func(args []runtime.Value) (runtime.Value, error) {
        dir, err := os.Getwd()
        if err != nil {
            return runtime.VStr(""), nil
        }
        return runtime.VStr(dir), nil
    })
}
