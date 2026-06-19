package stdlib

import (
    "bufio"
    "fmt"
    "luminalang/runtime"
    "os"
    "strings"
)

func registerIO(interp *runtime.Interpreter) {
    interp.RegisterBuiltin("io.println", func(args []runtime.Value) (runtime.Value, error) {
        for _, a := range args {
            fmt.Print(runtime.ValueString(a))
        }
        fmt.Println()
        return runtime.VInt(0), nil
    })
    interp.RegisterBuiltin("io.print", func(args []runtime.Value) (runtime.Value, error) {
        for _, a := range args {
            fmt.Print(runtime.ValueString(a))
        }
        return runtime.VInt(0), nil
    })
    interp.RegisterBuiltin("io.readint", func(args []runtime.Value) (runtime.Value, error) {
        var v int64
        _, err := fmt.Scanf("%d", &v)
        if err != nil {
            return runtime.VInt(0), err
        }
        return runtime.VInt(v), nil
    })
    interp.RegisterBuiltin("io.eprint", func(args []runtime.Value) (runtime.Value, error) {
        for _, a := range args {
            fmt.Fprint(os.Stderr, runtime.ValueString(a))
        }
        return runtime.VInt(0), nil
    })
    interp.RegisterBuiltin("io.eprintln", func(args []runtime.Value) (runtime.Value, error) {
        for _, a := range args {
            fmt.Fprint(os.Stderr, runtime.ValueString(a))
        }
        fmt.Fprintln(os.Stderr)
        return runtime.VInt(0), nil
    })
    interp.RegisterBuiltin("io.flush", func(args []runtime.Value) (runtime.Value, error) {
        return runtime.VInt(0), nil
    })
    interp.RegisterBuiltin("io.readline", func(args []runtime.Value) (runtime.Value, error) {
        reader := bufio.NewReader(os.Stdin)
        line, _ := reader.ReadString('\n')
        line = strings.TrimRight(line, "\r\n")
        return runtime.VStr(line), nil
    })
}
