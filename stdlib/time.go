package stdlib

import (
    "luminalang/runtime"
    "time"
)

var TimeModule = Module{
    Name: "time",
    Funcs: map[string]runtime.BuiltinFn{
        "now": func(args []runtime.Value) (runtime.Value, error) {
            return time.Now().Unix(), nil
        },
        "year": func(args []runtime.Value) (runtime.Value, error) {
            return runtime.Value(time.Now().Year()), nil
        },
        "month": func(args []runtime.Value) (runtime.Value, error) {
            return runtime.Value(time.Now().Month()), nil
        },
        "day": func(args []runtime.Value) (runtime.Value, error) {
            return runtime.Value(time.Now().Day()), nil
        },
        "hour": func(args []runtime.Value) (runtime.Value, error) {
            return runtime.Value(time.Now().Hour()), nil
        },
        "minute": func(args []runtime.Value) (runtime.Value, error) {
            return runtime.Value(time.Now().Minute()), nil
        },
        "second": func(args []runtime.Value) (runtime.Value, error) {
            return runtime.Value(time.Now().Second()), nil
        },
        "weekday": func(args []runtime.Value) (runtime.Value, error) {
            wd := time.Now().Weekday()
            if wd == 0 {
                return 7, nil
            }
            return runtime.Value(wd), nil
        },
        "sleep_ms": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 1 {
                return 0, errArity("time::sleep_ms", 1, len(args))
            }
            if args[0] < 0 {
                return 0, errMsg("time::sleep_ms: negative duration")
            }
            time.Sleep(time.Duration(args[0]) * time.Millisecond)
            return 0, nil
        },
    },
}
