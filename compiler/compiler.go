package compiler

import (
	"fmt"
	"luminalang/ast"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Compiler struct {
	functions map[string]*ast.FnDecl
	imports   map[string]string
	counter   int
}

type fnCtx struct {
	c         *Compiler
	retType   Type
	varTypes  map[string]Type
	forVars   map[string]string
	loopDepth int
}

func Compile(prog *ast.Program, outPath string) error {
	c := &Compiler{
		functions: make(map[string]*ast.FnDecl),
		imports:   resolveImports(prog.Imports),
	}
	for i := range prog.Functions {
		fn := &prog.Functions[i]
		c.functions[fn.Name] = fn
	}

	mainFn, ok := c.functions["main"]
	if !ok {
		return fmt.Errorf("compile error: main function not found")
	}
	if typeFromName(mainFn.RetType) != TInt {
		return fmt.Errorf("compile error: main function must return i32")
	}

	var body strings.Builder
	for _, fn := range c.functions {
		code, err := c.genFunction(fn)
		if err != nil {
			return err
		}
		body.WriteString(code)
		body.WriteString("\n")
	}

	source := runtimeSource + "\n" +
		body.String() +
		"func main() {\n" +
		"    os.Exit(int(lumf_main()))\n" +
		"}\n"

	dir, err := os.MkdirTemp("", "luminaout-*")
	if err != nil {
		return fmt.Errorf("cannot create build dir: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module luminaout\n\ngo 1.21\n"), 0644); err != nil {
		return fmt.Errorf("cannot write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(source), 0644); err != nil {
		return fmt.Errorf("cannot write generated source: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %v", err)
	}

	absOut, err := filepath.Abs(outPath)
	if err != nil {
		return fmt.Errorf("cannot resolve output path '%s': %v", outPath, err)
	}

	cmd := exec.Command("go", "build", "-o", absOut, ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("native build failed:\n%s", string(out))
	}
	return nil
}

func resolveImports(imports []ast.UseDecl) map[string]string {
	m := make(map[string]string)
	for _, imp := range imports {
		prefix := imp.Module + "."
		if imp.Wildcard {
			for full := range builtins {
				if strings.HasPrefix(full, prefix) {
					m[strings.TrimPrefix(full, prefix)] = full
				}
			}
		} else {
			for _, item := range imp.Items {
				full := imp.Module + "." + item
				if _, ok := builtins[full]; ok {
					m[item] = full
				}
			}
		}
	}
	return m
}
