package stdlib

import (
    "fmt"
    "luminalang/runtime"
    "time"
)

func registerTime(interp *runtime.Interpreter) {
    interp.RegisterBuiltin("time.now", func(args []runtime.Value) (runtime.Value, error) {
        return runtime.VInt(time.Now().Unix()), nil
    })
    interp.RegisterBuiltin("time.year", func(args []runtime.Value) (runtime.Value, error) {
        return runtime.VInt(int64(time.Now().Year())), nil
    })
    interp.RegisterBuiltin("time.month", func(args []runtime.Value) (runtime.Value, error) {
        return runtime.VInt(int64(time.Now().Month())), nil
    })
    interp.RegisterBuiltin("time.day", func(args []runtime.Value) (runtime.Value, error) {
        return runtime.VInt(int64(time.Now().Day())), nil
    })
    interp.RegisterBuiltin("time.hour", func(args []runtime.Value) (runtime.Value, error) {
        return runtime.VInt(int64(time.Now().Hour())), nil
    })
    interp.RegisterBuiltin("time.minute", func(args []runtime.Value) (runtime.Value, error) {
        return runtime.VInt(int64(time.Now().Minute())), nil
    })
    interp.RegisterBuiltin("time.second", func(args []runtime.Value) (runtime.Value, error) {
        return runtime.VInt(int64(time.Now().Second())), nil
    })
    interp.RegisterBuiltin("time.weekday", func(args []runtime.Value) (runtime.Value, error) {
        return runtime.VInt(int64(time.Now().Weekday())), nil
    })
    interp.RegisterBuiltin("time.sleep_ms", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 1 {
            return runtime.VInt(0), fmt.Errorf("sleep_ms: expected 1 argument")
        }
        time.Sleep(time.Duration(args[0].Int()) * time.Millisecond)
        return runtime.VInt(0), nil
    })
}
