package main

import (
    "fmt"
    "luminalang/ast"
    "luminalang/lexer"
    "luminalang/parser"
    "luminalang/runtime"
    "luminalang/stdlib"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("LuminaLang v0.7.0 interpreter")
        fmt.Println("Usage: luminalang <file.lum>")
        os.Exit(1)
    }
    filename := os.Args[1]
    source, err := os.ReadFile(filename)
    if err != nil {
        fmt.Printf("Cannot read file '%s': %v\n", filename, err)
        os.Exit(1)
    }
    l := lexer.New(string(source))
    tokens, err := l.Tokenize()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    p := parser.New(tokens)
    prog, err := p.Parse()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    interp := runtime.NewInterpreter()
    if err := registerUses(interp, prog); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    code, err := interp.Run(prog)
    if err != nil {
        if bs, ok := err.(runtime.BackSignal); ok {
            os.Exit(int(bs.Value))
        }
        fmt.Println(err)
        os.Exit(1)
    }
    os.Exit(code)
}

func registerUses(interp *runtime.Interpreter, prog *ast.Program) error {
    for _, use := range prog.Imports {
        mod, ok := stdlib.Registry[use.Module]
        if !ok {
            return fmt.Errorf("type error, unknown module '%s'", use.Module)
        }
        if use.Wildcard {
            for name, fn := range mod.Funcs {
                interp.RegisterBuiltin(name, fn)
            }
        } else {
            for _, item := range use.Items {
                fn, ok := mod.Funcs[item]
                if !ok {
                    return fmt.Errorf("type error, module '%s' has no function '%s'", use.Module, item)
                }
                interp.RegisterBuiltin(item, fn)
            }
        }
    }
    return nil
}
