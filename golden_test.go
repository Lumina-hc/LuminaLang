package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"luminalang/compiler"
	"luminalang/lexer"
	"luminalang/parser"
)

func buildAndRun(t *testing.T, srcPath string) string {
	t.Helper()
	source, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("read %s: %v", srcPath, err)
	}
	toks, err := lexer.New(string(source)).Tokenize()
	if err != nil {
		t.Fatalf("lex %s: %v", srcPath, err)
	}
	prog, err := parser.New(toks).Parse()
	if err != nil {
		t.Fatalf("parse %s: %v", srcPath, err)
	}
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	out := filepath.Join(t.TempDir(), "prog"+ext)
	if err := compiler.CompileFile(prog, out, "2", filepath.Dir(srcPath)); err != nil {
		t.Fatalf("compile %s: %v", srcPath, err)
	}
	cmd := exec.Command(out)
	cmd.Stdin = strings.NewReader("")
	data, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run %s: %v\n%s", srcPath, err, data)
	}
	return string(data)
}

func normalize(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

func TestGolden(t *testing.T) {
	var matches []string
	for _, pat := range []string{"tests/*.lum", "tests/*/main.lum"} {
		m, err := filepath.Glob(pat)
		if err != nil {
			t.Fatal(err)
		}
		matches = append(matches, m...)
	}
	if len(matches) == 0 {
		t.Skip("no golden tests found")
	}
	for _, src := range matches {
		src := src
		t.Run(filepath.Base(src), func(t *testing.T) {
			got := buildAndRun(t, src)
			wantPath := strings.TrimSuffix(src, ".lum") + ".out"
			want, err := os.ReadFile(wantPath)
			if err != nil {
				t.Fatalf("missing expected output %s", wantPath)
			}
			if normalize(got) != normalize(string(want)) {
				t.Errorf("output mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", src, got, want)
			}
		})
	}
}

func TestGoldenFailures(t *testing.T) {
	matches, err := filepath.Glob("tests/fail/*.lum")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Skip("no failure tests found")
	}
	for _, src := range matches {
		src := src
		t.Run(filepath.Base(src), func(t *testing.T) {
			source, err := os.ReadFile(src)
			if err != nil {
				t.Fatal(err)
			}
			toks, err := lexer.New(string(source)).Tokenize()
			if err != nil {
				return
			}
			prog, err := parser.New(toks).Parse()
			if err != nil {
				return
			}
			out := filepath.Join(t.TempDir(), "prog"+map[bool]string{true: ".exe"}[runtime.GOOS == "windows"])
			if err := compiler.CompileFile(prog, out, "2", filepath.Dir(src)); err == nil {
				t.Errorf("%s: expected a compile error, got none", src)
			}
		})
	}
}
