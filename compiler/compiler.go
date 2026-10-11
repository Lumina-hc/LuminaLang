package compiler

import (
	"fmt"
	"luminalang/ast"
	"luminalang/lexer"
	"luminalang/parser"
	"luminalang/stdlib"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type importTarget struct {
	kind    int
	builtin string
	symbol  string
}

type Compiler struct {
	functions   map[string]*ast.FnDecl
	fnModule    map[string]string
	modImports  map[string]map[string]importTarget
	loaded      map[string]bool
	fileModules map[string]bool
	libIR       map[string]string
	libMods     map[string]bool
	libDir      string
	srcDir      string
	emitModule  string
	counter     int
	globals     []globalStr
}

func newCompiler(libDir, srcDir string) *Compiler {
	return &Compiler{
		functions:   make(map[string]*ast.FnDecl),
		fnModule:    make(map[string]string),
		modImports:  make(map[string]map[string]importTarget),
		loaded:      make(map[string]bool),
		fileModules: make(map[string]bool),
		libIR:       make(map[string]string),
		libMods:     make(map[string]bool),
		libDir:      libDir,
		srcDir:      srcDir,
	}
}

func mangle(sym string) string {
	return "lumf_" + strings.ReplaceAll(sym, ".", "_")
}

func prepare(prog *ast.Program, srcDir string) (*Compiler, string, func(), string, error) {
	libDir, cleanup, err := resolveLibDir()
	if err != nil {
		return nil, "", nil, "", err
	}
	c := newCompiler(libDir, srcDir)

	for i := range prog.Functions {
		fn := &prog.Functions[i]
		c.functions[fn.Name] = fn
		c.fnModule[fn.Name] = ""
	}

	mainFn, ok := c.functions["main"]
	if !ok {
		cleanup()
		return nil, "", nil, "", fmt.Errorf("compile error: main function not found")
	}
	if typeFromName(mainFn.RetType) != TInt {
		cleanup()
		return nil, "", nil, "", fmt.Errorf("compile error: main function must return i32")
	}

	imports, err := c.resolveImports("", prog.Imports)
	if err != nil {
		cleanup()
		return nil, "", nil, "", err
	}
	c.modImports[""] = imports

	source, err := c.genProgram()
	if err != nil {
		cleanup()
		return nil, "", nil, "", err
	}
	return c, libDir, cleanup, source, nil
}

func GenerateIR(prog *ast.Program) (string, error) {
	return GenerateIRAt(prog, "")
}

func GenerateIRAt(prog *ast.Program, srcDir string) (string, error) {
	c, _, cleanup, source, err := prepare(prog, srcDir)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		return "", err
	}
	names := make([]string, 0, len(c.libIR))
	for n := range c.libIR {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	b.WriteString(source)
	for _, n := range names {
		b.WriteString("\n")
		b.WriteString(c.libIR[n])
	}
	return cleanDeclares(b.String()), nil
}

func symbolName(line string) string {
	at := strings.Index(line, "@")
	if at < 0 {
		return ""
	}
	rest := line[at+1:]
	if p := strings.IndexByte(rest, '('); p >= 0 {
		return rest[:p]
	}
	return strings.TrimSpace(rest)
}

func cleanDeclares(s string) string {
	lines := strings.Split(s, "\n")
	defined := make(map[string]bool)
	for _, ln := range lines {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "define ") {
			defined[symbolName(t)] = true
		}
	}
	seen := make(map[string]bool)
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "declare ") {
			name := symbolName(t)
			if defined[name] || seen[name] {
				continue
			}
			seen[name] = true
		}
		out = append(out, ln)
	}
	return strings.Join(out, "\n")
}

func Compile(prog *ast.Program, outPath string) error {
	return CompileFile(prog, outPath, "2", "")
}

func CompileWith(prog *ast.Program, outPath, optLevel string) error {
	return CompileFile(prog, outPath, optLevel, "")
}

func CompileFile(prog *ast.Program, outPath, optLevel, srcDir string) error {
	clang, err := findClang()
	if err != nil {
		return err
	}
	c, libDir, cleanup, source, err := prepare(prog, srcDir)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %v", err)
	}
	absOut, err := filepath.Abs(outPath)
	if err != nil {
		return fmt.Errorf("cannot resolve output path '%s': %v", outPath, err)
	}

	dir, err := os.MkdirTemp("", "luminaout-*")
	if err != nil {
		return fmt.Errorf("cannot create build dir: %v", err)
	}
	defer os.RemoveAll(dir)

	irPath := filepath.Join(dir, "main.ll")
	if err := os.WriteFile(irPath, []byte(source), 0644); err != nil {
		return fmt.Errorf("cannot write generated IR: %v", err)
	}

	libPaths, err := writeLibIR(dir, c)
	if err != nil {
		return err
	}

	cs, err := cSources(libDir)
	if err != nil {
		return err
	}

	args := []string{"-w", "-O" + optLevel, "-Wno-override-module", "-I", libDir}
	if runtime.GOOS == "windows" {
		args = append(args, "-D_CRT_SECURE_NO_WARNINGS")
	}
	args = append(args, "-o", absOut, irPath)
	args = append(args, libPaths...)
	args = append(args, cs...)
	if runtime.GOOS == "windows" {
		args = append(args, "-luser32", "-lshell32", "-ladvapi32")
	} else {
		args = append(args, "-lm")
	}

	cmd := exec.Command(clang, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("native build failed:\n%s", string(out))
	}
	return nil
}

func writeLibIR(dir string, c *Compiler) ([]string, error) {
	names := make([]string, 0, len(c.libIR))
	for n := range c.libIR {
		names = append(names, n)
	}
	sort.Strings(names)
	paths := make([]string, 0, len(names))
	for _, n := range names {
		p := filepath.Join(dir, "lib_"+n+".ll")
		if err := os.WriteFile(p, []byte(c.libIR[n]), 0644); err != nil {
			return nil, fmt.Errorf("cannot write module IR for '%s': %v", n, err)
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func cSources(libDir string) ([]string, error) {
	entries, err := os.ReadDir(libDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read standard library at '%s': %v", libDir, err)
	}
	var cs []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".c") {
			cs = append(cs, filepath.Join(libDir, e.Name()))
		}
	}
	if len(cs) == 0 {
		return nil, fmt.Errorf("standard library at '%s' contains no C sources", libDir)
	}
	sort.Strings(cs)
	return cs, nil
}

func (c *Compiler) resolveImports(module string, imports []ast.UseDecl) (map[string]importTarget, error) {
	m := make(map[string]importTarget)
	for _, imp := range imports {
		mod := imp.Module
		if err := c.loadModule(mod); err != nil {
			return nil, err
		}
		if imp.Wildcard {
			for sym := range c.functions {
				if c.fnModule[sym] == mod {
					name := sym
					if i := strings.LastIndex(sym, "."); i >= 0 {
						name = sym[i+1:]
					}
					m[name] = importTarget{kind: 1, symbol: sym}
				}
			}
			prefix := mod + "."
			for key := range builtins {
				if strings.HasPrefix(key, prefix) {
					m[strings.TrimPrefix(key, prefix)] = importTarget{kind: 0, builtin: key}
				}
			}
			continue
		}
		for _, item := range imp.Items {
			local := item.Name
			if item.Alias != "" {
				local = item.Alias
			}
			sym := mod + "." + item.Name
			if _, ok := c.functions[sym]; ok {
				m[local] = importTarget{kind: 1, symbol: sym}
			} else if _, ok := builtins[sym]; ok {
				m[local] = importTarget{kind: 0, builtin: sym}
			} else {
				return nil, fmt.Errorf("compile error: module '%s' has no item '%s'", mod, item.Name)
			}
		}
	}
	return m, nil
}

func (c *Compiler) searchDirs() []string {
	dirs := make([]string, 0, 4)
	if c.srcDir != "" {
		dirs = append(dirs, c.srcDir)
	}
	if v := os.Getenv("LUMINA_PATH"); v != "" {
		for _, d := range filepath.SplitList(v) {
			if d != "" {
				dirs = append(dirs, d)
			}
		}
	}
	dirs = append(dirs, c.libDir)
	return dirs
}

func (c *Compiler) loadModule(name string) error {
	if c.loaded[name] {
		return nil
	}
	c.loaded[name] = true
	lumcPath, lumPath, lumiPath := "", "", ""
	for _, dir := range c.searchDirs() {
		if lumcPath == "" {
			if p := filepath.Join(dir, name+".lumc"); fileExists(p) {
				lumcPath = p
			}
		}
		if lumPath == "" {
			if p := filepath.Join(dir, name+".lum"); fileExists(p) {
				lumPath = p
			}
		}
		if lumiPath == "" {
			if p := filepath.Join(dir, name+".lumi"); fileExists(p) {
				lumiPath = p
			}
		}
	}
	if lumcPath != "" {
		meta, ir, err := readLumc(lumcPath)
		if err != nil {
			return fmt.Errorf("module '%s': %v", name, err)
		}
		paired := filepath.Join(filepath.Dir(lumcPath), name+".lum")
		if fileExists(paired) && meta.SourceHash != hashFile(paired) {
			return c.loadSource(name, paired)
		}
		return c.loadLumc(name, meta, ir)
	}
	if lumPath != "" {
		return c.loadSource(name, lumPath)
	}
	if lumiPath != "" {
		data, err := os.ReadFile(lumiPath)
		if err != nil {
			return err
		}
		_, funcs, err := ParseDecl(string(data))
		if err != nil {
			return fmt.Errorf("module '%s': %v", name, err)
		}
		return c.loadDecl(name, funcs)
	}
	return nil
}

// loadDecl registers a module from a declaration-only .lumi file: its
// functions become external declarations with no IR of their own.
func (c *Compiler) loadDecl(name string, funcs []lumcFunc) error {
	for _, lf := range funcs {
		sym := name + "." + lf.Name
		if _, dup := c.functions[sym]; dup {
			continue
		}
		c.functions[sym] = toFnDecl(lf)
		c.fnModule[sym] = name
	}
	c.libMods[name] = true
	c.fileModules[name] = true
	return nil
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

func (c *Compiler) loadLumc(name string, meta lumcMeta, ir string) error {
	for _, lf := range meta.Funcs {
		sym := name + "." + lf.Name
		if _, dup := c.functions[sym]; dup {
			return fmt.Errorf("module '%s': duplicate function '%s'", name, lf.Name)
		}
		c.functions[sym] = toFnDecl(lf)
		c.fnModule[sym] = name
	}
	c.libIR[name] = ir
	c.libMods[name] = true
	c.fileModules[name] = true
	for _, dep := range meta.Uses {
		if err := c.loadModule(dep); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) loadSource(name, path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	l := lexer.New(string(src))
	toks, err := l.Tokenize()
	if err != nil {
		return fmt.Errorf("module '%s': %v", name, err)
	}
	mod, err := parser.New(toks).Parse()
	if err != nil {
		return fmt.Errorf("module '%s': %v", name, err)
	}
	for i := range mod.Functions {
		fn := &mod.Functions[i]
		if fn.Name == "main" {
			continue
		}
		sym := name + "." + fn.Name
		if _, dup := c.functions[sym]; dup {
			return fmt.Errorf("module '%s': duplicate function '%s'", name, fn.Name)
		}
		c.functions[sym] = fn
		c.fnModule[sym] = name
	}
	imports, err := c.resolveImports(name, mod.Imports)
	if err != nil {
		return err
	}
	c.modImports[name] = imports
	c.fileModules[name] = true
	return nil
}

func (c *Compiler) lookupCall(module, callee string) (importTarget, *ast.FnDecl, bool) {
	local := callee
	if module != "" {
		local = module + "." + callee
	}
	if f, ok := c.functions[local]; ok {
		return importTarget{kind: 1, symbol: local}, f, true
	}
	if t, ok := c.modImports[module][callee]; ok {
		if t.kind == 1 {
			return t, c.functions[t.symbol], true
		}
		return t, nil, true
	}
	return importTarget{}, nil, false
}

func findClang() (string, error) {
	if v := os.Getenv("LUMINA_CC"); v != "" {
		return v, nil
	}
	if p, err := exec.LookPath("clang"); err == nil {
		return p, nil
	}
	candidates := []string{
		`C:\Program Files\LLVM\bin\clang.exe`,
		`C:\Program Files (x86)\LLVM\bin\clang.exe`,
		"/usr/bin/clang",
		"/usr/local/bin/clang",
		"/opt/homebrew/opt/llvm/bin/clang",
	}
	for _, p := range candidates {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p, nil
		}
	}
	return "", fmt.Errorf("cannot find 'clang'; install LLVM or set LUMINA_CC to the clang executable")
}

func StdlibDir() (string, func(), error) {
	return resolveLibDir()
}

func resolveLibDir() (string, func(), error) {
	noop := func() {}
	var candidates []string
	if v := os.Getenv("LUMINALANG_LIB"); v != "" {
		candidates = append(candidates, v)
	}
	if exe, err := os.Executable(); err == nil {
		d := filepath.Dir(exe)
		for _, base := range []string{d, filepath.Join(d, ".."), filepath.Join(d, "..", "share", "luminalang")} {
			candidates = append(candidates, filepath.Join(base, "stdlib", "lib"), filepath.Join(base, "stdlib"))
		}
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "stdlib", "lib"), filepath.Join(wd, "stdlib"))
	}
	for _, dir := range candidates {
		if hasStdlib(dir) {
			return dir, noop, nil
		}
	}

	tmp, err := os.MkdirTemp("", "luminalib-*")
	if err != nil {
		return "", noop, fmt.Errorf("cannot create library dir: %v", err)
	}
	if err := stdlib.Extract(tmp); err != nil {
		os.RemoveAll(tmp)
		return "", noop, fmt.Errorf("cannot unpack embedded standard library: %v", err)
	}
	return tmp, func() { os.RemoveAll(tmp) }, nil
}

func hasStdlib(dir string) bool {
	for _, f := range []string{"runtime.c", "math.lum"} {
		if fi, err := os.Stat(filepath.Join(dir, f)); err != nil || fi.IsDir() {
			return false
		}
	}
	return true
}
