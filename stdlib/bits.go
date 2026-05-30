package stdlib

import "luminalang/runtime"

var BitsModule = Module{
    Name: "bits",
    Funcs: map[string]runtime.BuiltinFn{
        "and": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 2 {
                return 0, errArity("bits::and", 2, len(args))
            }
            return args[0] & args[1], nil
        },
        "or": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 2 {
                return 0, errArity("bits::or", 2, len(args))
            }
            return args[0] | args[1], nil
        },
        "xor": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 2 {
                return 0, errArity("bits::xor", 2, len(args))
            }
            return args[0] ^ args[1], nil
        },
        "not": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 1 {
                return 0, errArity("bits::not", 1, len(args))
            }
            return ^args[0], nil
        },
        "shl": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 2 {
                return 0, errArity("bits::shl", 2, len(args))
            }
            return args[0] << uint(args[1]), nil
        },
        "shr": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 2 {
                return 0, errArity("bits::shr", 2, len(args))
            }
            return args[0] >> uint(args[1]), nil
        },
        "bit": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 2 {
                return 0, errArity("bits::bit", 2, len(args))
            }
            return (args[0] >> uint(args[1])) & 1, nil
        },
        "setbit": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 2 {
                return 0, errArity("bits::setbit", 2, len(args))
            }
            return args[0] | (1 << uint(args[1])), nil
        },
        "clrbit": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 2 {
                return 0, errArity("bits::clrbit", 2, len(args))
            }
            return args[0] & ^(1 << uint(args[1])), nil
        },
        "popcount": func(args []runtime.Value) (runtime.Value, error) {
            if len(args) != 1 {
                return 0, errArity("bits::popcount", 1, len(args))
            }
            n := args[0]
            count := runtime.Value(0)
            for n != 0 {
                count++
                n &= n - 1
            }
            return count, nil
        },
    },
}
