package compiler

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"luminalang/ast"
	"luminalang/lexer"
	"luminalang/parser"
)

const lumcMagic = "LUMC"
const lumcVersion uint16 = 1

type lumcParam struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type lumcFunc struct {
	Name   string      `json:"name"`
	Params []lumcParam `json:"params"`
	Ret    string      `json:"ret"`
}

type lumcMeta struct {
	Module     string     `json:"module"`
	SourceHash string     `json:"source_hash,omitempty"`
	Uses       []string   `json:"uses"`
	Funcs      []lumcFunc `json:"funcs"`
}

func hashFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func writeLumc(path string, meta lumcMeta, ir string) error {
	jb, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.WriteString(lumcMagic)
	var tag [2]byte
	binary.LittleEndian.PutUint16(tag[:], lumcVersion)
	buf.Write(tag[:])
	var u4 [4]byte
	binary.LittleEndian.PutUint32(u4[:], uint32(len(jb)))
	buf.Write(u4[:])
	buf.Write(jb)
	binary.LittleEndian.PutUint32(u4[:], uint32(len(ir)))
	buf.Write(u4[:])
	buf.WriteString(ir)
	return os.WriteFile(path, buf.Bytes(), 0644)
}

func readLumc(path string) (lumcMeta, string, error) {
	var meta lumcMeta
	data, err := os.ReadFile(path)
	if err != nil {
		return meta, "", err
	}
	if len(data) < 6 || string(data[:4]) != lumcMagic {
		return meta, "", fmt.Errorf("not a LuminaLang compiled library")
	}
	pos := 4
	ver := binary.LittleEndian.Uint16(data[pos : pos+2])
	pos += 2
	if ver != lumcVersion {
		return meta, "", fmt.Errorf("unsupported .lumc version %d", ver)
	}
	if len(data) < pos+4 {
		return meta, "", fmt.Errorf("truncated .lumc header")
	}
	jsonLen := int(binary.LittleEndian.Uint32(data[pos : pos+4]))
	pos += 4
	if len(data) < pos+jsonLen+4 {
		return meta, "", fmt.Errorf("truncated .lumc metadata")
	}
	if err := json.Unmarshal(data[pos:pos+jsonLen], &meta); err != nil {
		return meta, "", fmt.Errorf("corrupt .lumc metadata: %v", err)
	}
	pos += jsonLen
	irLen := int(binary.LittleEndian.Uint32(data[pos : pos+4]))
	pos += 4
	if len(data) < pos+irLen {
		return meta, "", fmt.Errorf("truncated .lumc IR payload")
	}
	return meta, string(data[pos : pos+irLen]), nil
}

func toFnDecl(lf lumcFunc) *ast.FnDecl {
	fn := &ast.FnDecl{Name: lf.Name, RetType: lf.Ret}
	for _, p := range lf.Params {
		fn.Params = append(fn.Params, ast.Param{Name: p.Name, Type: p.Type})
	}
	return fn
}

func BuildModule(srcPath, outPath string) error {
	libDir, cleanup, err := resolveLibDir()
	if err != nil {
		return err
	}
	defer cleanup()

	src, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("cannot read '%s': %v", srcPath, err)
	}
	toks, err := lexer.New(string(src)).Tokenize()
	if err != nil {
		return fmt.Errorf("module '%s': %v", srcPath, err)
	}
	prog, err := parser.New(toks).Parse()
	if err != nil {
		return fmt.Errorf("module '%s': %v", srcPath, err)
	}

	name := strings.TrimSuffix(filepath.Base(srcPath), ".lum")
	c := newCompiler(libDir, filepath.Dir(srcPath))
	for i := range prog.Functions {
		fn := &prog.Functions[i]
		if fn.Name == "main" {
			continue
		}
		sym := name + "." + fn.Name
		c.functions[sym] = fn
		c.fnModule[sym] = name
	}

	imports, err := c.resolveImports(name, prog.Imports)
	if err != nil {
		return err
	}
	c.modImports[name] = imports

	c.emitModule = name
	ir, err := c.genProgram()
	if err != nil {
		return err
	}

	meta := lumcMeta{Module: name}
	sum := sha256.Sum256(src)
	meta.SourceHash = hex.EncodeToString(sum[:])
	var syms []string
	for sym := range c.functions {
		if c.fnModule[sym] == name {
			syms = append(syms, sym)
		}
	}
	sort.Strings(syms)
	for _, sym := range syms {
		fn := c.functions[sym]
		lf := lumcFunc{Name: fn.Name, Ret: fn.RetType}
		for _, p := range fn.Params {
			lf.Params = append(lf.Params, lumcParam{Name: p.Name, Type: p.Type})
		}
		meta.Funcs = append(meta.Funcs, lf)
	}
	for m := range c.fileModules {
		meta.Uses = append(meta.Uses, m)
	}
	sort.Strings(meta.Uses)

	if outPath == "" {
		outPath = strings.TrimSuffix(srcPath, ".lum") + ".lumc"
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %v", err)
	}
	if err := writeLumc(outPath, meta, ir); err != nil {
		return err
	}
	declPath := filepath.Join(filepath.Dir(outPath), name+".lumi")
	return os.WriteFile(declPath, []byte(WriteDecl(name, meta.Funcs)), 0644)
}

func BuildStdlib(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("cannot read '%s': %v", dir, err)
	}
	var mods []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".lum") {
			mods = append(mods, e.Name())
		}
	}
	sort.Strings(mods)
	for _, m := range mods {
		src := filepath.Join(dir, m)
		out := filepath.Join(dir, strings.TrimSuffix(m, ".lum")+".lumc")
		if err := BuildModule(src, out); err != nil {
			return err
		}
	}
	return nil
}
