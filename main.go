package main

import (
    "fmt"
    "luminalang/lexer"
    "luminalang/parser"
    "luminalang/runtime"
    "luminalang/stdlib"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("LuminaLang v0.8.0 interpreter")
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
    stdlib.RegisterAll(interp)
    interp.ResolveImports(prog.Imports)
    code, err := interp.Run(prog)
    if err != nil {
        if bs, ok := err.(runtime.BackSignal); ok {
            os.Exit(int(bs.Value.Int()))
        }
        fmt.Println(err)
        os.Exit(1)
    }
    os.Exit(code)
}
