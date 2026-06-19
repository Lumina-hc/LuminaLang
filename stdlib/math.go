package stdlib

import (
    "fmt"
    "luminalang/runtime"
)

func registerMath(interp *runtime.Interpreter) {
    interp.RegisterBuiltin("math.abs", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 1 {
            return runtime.VInt(0), fmt.Errorf("abs: expected 1 argument")
        }
        v := args[0].Int()
        if v < 0 {
            v = -v
        }
        return runtime.VInt(v), nil
    })
    interp.RegisterBuiltin("math.max", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 2 {
            return runtime.VInt(0), fmt.Errorf("max: expected 2 arguments")
        }
        if args[0].Int() > args[1].Int() {
            return args[0], nil
        }
        return args[1], nil
    })
    interp.RegisterBuiltin("math.min", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 2 {
            return runtime.VInt(0), fmt.Errorf("min: expected 2 arguments")
        }
        if args[0].Int() < args[1].Int() {
            return args[0], nil
        }
        return args[1], nil
    })
    interp.RegisterBuiltin("math.pow", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 2 {
            return runtime.VInt(0), fmt.Errorf("pow: expected 2 arguments")
        }
        base, exp := args[0].Int(), args[1].Int()
        if exp < 0 {
            return runtime.VInt(0), fmt.Errorf("pow: negative exponent")
        }
        r := int64(1)
        for i := int64(0); i < exp; i++ {
            r *= base
        }
        return runtime.VInt(r), nil
    })
    interp.RegisterBuiltin("math.sqrt", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 1 {
            return runtime.VInt(0), fmt.Errorf("sqrt: expected 1 argument")
        }
        n := args[0].Int()
        if n < 0 {
            return runtime.VInt(0), fmt.Errorf("sqrt: negative number")
        }
        lo, hi := int64(0), n+1
        for lo < hi {
            mid := (lo + hi + 1) / 2
            if mid*mid <= n {
                lo = mid
            } else {
                hi = mid - 1
            }
        }
        return runtime.VInt(lo), nil
    })
    interp.RegisterBuiltin("math.gcd", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 2 {
            return runtime.VInt(0), fmt.Errorf("gcd: expected 2 arguments")
        }
        a, b := args[0].Int(), args[1].Int()
        if a < 0 {
            a = -a
        }
        if b < 0 {
            b = -b
        }
        for b != 0 {
            a, b = b, a%b
        }
        return runtime.VInt(a), nil
    })
    interp.RegisterBuiltin("math.lcm", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 2 {
            return runtime.VInt(0), fmt.Errorf("lcm: expected 2 arguments")
        }
        a, b := args[0].Int(), args[1].Int()
        if a == 0 || b == 0 {
            return runtime.VInt(0), nil
        }
        ga, gb := a, b
        if ga < 0 {
            ga = -ga
        }
        if gb < 0 {
            gb = -gb
        }
        for gb != 0 {
            ga, gb = gb, ga%gb
        }
        return runtime.VInt((a / ga) * b), nil
    })
    interp.RegisterBuiltin("math.mod", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 2 {
            return runtime.VInt(0), fmt.Errorf("mod: expected 2 arguments")
        }
        return runtime.VInt(args[0].Int() % args[1].Int()), nil
    })
    interp.RegisterBuiltin("math.clamp", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 3 {
            return runtime.VInt(0), fmt.Errorf("clamp: expected 3 arguments")
        }
        v, lo, hi := args[0].Int(), args[1].Int(), args[2].Int()
        if v < lo {
            v = lo
        }
        if v > hi {
            v = hi
        }
        return runtime.VInt(v), nil
    })
    interp.RegisterBuiltin("math.sign", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 1 {
            return runtime.VInt(0), fmt.Errorf("sign: expected 1 argument")
        }
        v := args[0].Int()
        if v > 0 {
            return runtime.VInt(1), nil
        }
        if v < 0 {
            return runtime.VInt(-1), nil
        }
        return runtime.VInt(0), nil
    })
}
