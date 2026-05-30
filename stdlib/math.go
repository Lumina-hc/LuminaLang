package stdlib

import "luminalang/runtime"

var MathModule = Module{
    Name: "math",
    Funcs: map[string]runtime.BuiltinFn{
        "abs": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 1 {
                return 0, errArity("math::abs", 1, len(args))
            }
            if args[0] < 0 {
                return -args[0], nil
            }
            return args[0], nil
        },
        "max": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 2 {
                return 0, errArity("math::max", 2, len(args))
            }
            if args[0] > args[1] {
                return args[0], nil
            }
            return args[1], nil
        },
        "min": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 2 {
                return 0, errArity("math::min", 2, len(args))
            }
            if args[0] < args[1] {
                return args[0], nil
            }
            return args[1], nil
        },
        "pow": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 2 {
                return 0, errArity("math::pow", 2, len(args))
            }
            base, exp := args[0], args[1]
            if exp < 0 {
                return 0, errMsg("math::pow: negative exponent not supported")
            }
            result := runtime.Value(1)
            for i := runtime.Value(0); i < exp; i++ {
                result *= base
            }
            return result, nil
        },
        "sqrt": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 1 {
                return 0, errArity("math::sqrt", 1, len(args))
            }
            n := args[0]
            if n < 0 {
                return 0, errMsg("math::sqrt: negative input")
            }
            if n == 0 {
                return 0, nil
            }
            x := n
            y := (x + 1) / 2
            for y < x {
                x = y
                y = (x + n/x) / 2
            }
            return x, nil
        },
        "gcd": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 2 {
                return 0, errArity("math::gcd", 2, len(args))
            }
            a, b := args[0], args[1]
            if a < 0 {
                a = -a
            }
            if b < 0 {
                b = -b
            }
            for b != 0 {
                a, b = b, a%b
            }
            return a, nil
        },
        "lcm": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 2 {
                return 0, errArity("math::lcm", 2, len(args))
            }
            a, b := args[0], args[1]
            if a < 0 {
                a = -a
            }
            if b < 0 {
                b = -b
            }
            if a == 0 || b == 0 {
                return 0, nil
            }
            ga, gb := a, b
            for gb != 0 {
                ga, gb = gb, ga%gb
            }
            return (a / ga) * b, nil
        },
        "mod": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 2 {
                return 0, errArity("math::mod", 2, len(args))
            }
            if args[1] == 0 {
                return 0, errMsg("math::mod: division by zero")
            }
            r := args[0] % args[1]
            if r < 0 {
                r += args[1]
            }
            return r, nil
        },
        "clamp": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 3 {
                return 0, errArity("math::clamp", 3, len(args))
            }
            val, lo, hi := args[0], args[1], args[2]
            if val < lo {
                return lo, nil
            }
            if val > hi {
                return hi, nil
            }
            return val, nil
        },
        "sign": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 1 {
                return 0, errArity("math::sign", 1, len(args))
            }
            n := args[0]
            if n > 0 {
                return 1, nil
            }
            if n < 0 {
                return -1, nil
            }
            return 0, nil
        },
    },
}
