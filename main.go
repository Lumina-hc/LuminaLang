package main

import (
	"fmt"
	"luminalang/compiler"
	"luminalang/lexer"
	"luminalang/parser"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
)

func main() {
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println("LuminaLang v0.9.0 (native compiler)")
		fmt.Println("Usage: luminalang <file.lum> [-o output]")
		os.Exit(1)
	}

	filename := args[0]

	outPath := ""
	if len(args) >= 3 && args[len(args)-2] == "-o" {
		outPath = args[len(args)-1]
	}

	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("cannot read file '%s': %v\n", filename, err)
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

	if outPath == "" {
		ext := filepath.Ext(filename)
		base := strings.TrimSuffix(filename, ext)
		if goruntime.GOOS == "windows" {
			outPath = base + ".exe"
		} else {
			outPath = base
		}
	}

	if err := compiler.Compile(prog, outPath); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
