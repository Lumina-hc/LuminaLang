package stdlib

import (
    "fmt"
    "luminalang/runtime"
)

func registerBits(interp *runtime.Interpreter) {
    interp.RegisterBuiltin("bits.and", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 2 {
            return runtime.VInt(0), fmt.Errorf("bits.and: expected 2 args")
        }
        return runtime.VInt(args[0].Int() & args[1].Int()), nil
    })
    interp.RegisterBuiltin("bits.or", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 2 {
            return runtime.VInt(0), fmt.Errorf("bits.or: expected 2 args")
        }
        return runtime.VInt(args[0].Int() | args[1].Int()), nil
    })
    interp.RegisterBuiltin("bits.xor", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 2 {
            return runtime.VInt(0), fmt.Errorf("bits.xor: expected 2 args")
        }
        return runtime.VInt(args[0].Int() ^ args[1].Int()), nil
    })
    interp.RegisterBuiltin("bits.not", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 1 {
            return runtime.VInt(0), fmt.Errorf("bits.not: expected 1 arg")
        }
        return runtime.VInt(^args[0].Int()), nil
    })
    interp.RegisterBuiltin("bits.shl", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 2 {
            return runtime.VInt(0), fmt.Errorf("bits.shl: expected 2 args")
        }
        return runtime.VInt(args[0].Int() << uint(args[1].Int())), nil
    })
    interp.RegisterBuiltin("bits.shr", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 2 {
            return runtime.VInt(0), fmt.Errorf("bits.shr: expected 2 args")
        }
        return runtime.VInt(args[0].Int() >> uint(args[1].Int())), nil
    })
    interp.RegisterBuiltin("bits.bit", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 2 {
            return runtime.VInt(0), fmt.Errorf("bits.bit: expected 2 args")
        }
        if (args[0].Int() >> uint(args[1].Int()))&1 == 1 {
            return runtime.VInt(1), nil
        }
        return runtime.VInt(0), nil
    })
    interp.RegisterBuiltin("bits.setbit", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 2 {
            return runtime.VInt(0), fmt.Errorf("bits.setbit: expected 2 args")
        }
        return runtime.VInt(args[0].Int() | (1 << uint(args[1].Int()))), nil
    })
    interp.RegisterBuiltin("bits.clrbit", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 2 {
            return runtime.VInt(0), fmt.Errorf("bits.clrbit: expected 2 args")
        }
        return runtime.VInt(args[0].Int() & ^(1 << uint(args[1].Int()))), nil
    })
    interp.RegisterBuiltin("bits.popcount", func(args []runtime.Value) (runtime.Value, error) {
        if len(args) != 1 {
            return runtime.VInt(0), fmt.Errorf("bits.popcount: expected 1 arg")
        }
        n := uint32(args[0].Int())
        c := 0
        for n != 0 {
            n &= n - 1
            c++
        }
        return runtime.VInt(int64(c)), nil
    })
}
