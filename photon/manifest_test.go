package photon

import (
	"strings"
	"testing"
)

func TestParseManifest(t *testing.T) {
	src := `# a sample manifest
[package]
name = "consumer"
version = "0.1.0"
description = "demo"
license = "MIT"
authors = ["Ada", "Grace"]

[lib]
modules = ["consumer", "consumer_extra"]

[dependencies]
geometry = "github:user/geometry@v0.1.0"
util = "../local/util"
`
	m, err := ParseManifest(src)
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if m.Name != "consumer" || m.Version != "0.1.0" || m.Description != "demo" || m.License != "MIT" {
		t.Errorf("package fields wrong: %+v", m)
	}
	if len(m.Authors) != 2 || m.Authors[0] != "Ada" || m.Authors[1] != "Grace" {
		t.Errorf("authors = %v", m.Authors)
	}
	if len(m.Modules) != 2 || m.Modules[1] != "consumer_extra" {
		t.Errorf("modules = %v", m.Modules)
	}
	if spec, ok := m.Get("geometry"); !ok || spec != "github:user/geometry@v0.1.0" {
		t.Errorf("geometry spec = %q, %v", spec, ok)
	}
	if len(m.Deps) != 2 {
		t.Errorf("deps = %v", m.Deps)
	}
}

func TestManifestRoundTrip(t *testing.T) {
	m := &Manifest{
		Name:        "consumer",
		Version:     "0.1.0",
		Description: "demo",
		License:     "MIT",
		Authors:     []string{"Ada"},
		Modules:     []string{"consumer"},
	}
	m.Set("geometry", "github:user/geometry@v0.1.0")
	m.Set("util", "../local/util")
	got, err := ParseManifest(m.Encode())
	if err != nil {
		t.Fatalf("reparse: %v", err)
	}
	if got.Name != m.Name || got.Version != m.Version || len(got.Authors) != 1 {
		t.Errorf("round trip lost data: %+v", got)
	}
	if len(got.Deps) != 2 || got.Deps[0].Name != "geometry" || got.Deps[1].Name != "util" {
		t.Errorf("deps = %v", got.Deps)
	}
}

func TestManifestRemoval(t *testing.T) {
	m := &Manifest{}
	m.Set("a", "1")
	m.Set("b", "2")
	m.Set("a", "3")
	if len(m.Deps) != 2 {
		t.Fatalf("Set should replace, deps=%v", m.Deps)
	}
	if spec, _ := m.Get("a"); spec != "3" {
		t.Errorf("a = %q", spec)
	}
	if !m.Remove("a") || m.Remove("a") {
		t.Errorf("Remove semantics wrong")
	}
	if _, ok := m.Get("a"); ok {
		t.Errorf("a still present")
	}
}

func TestManifestCommentHandling(t *testing.T) {
	m, err := ParseManifest("[dependencies]\nfoo = \"bar#baz\" # trailing\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if spec, _ := m.Get("foo"); spec != "bar#baz" {
		t.Errorf("spec = %q", spec)
	}
}

func TestManifestEscaping(t *testing.T) {
	m := &Manifest{Name: "app"}
	path := `C:\Users\dev\packages\a\util`
	m.Set("util", path)
	text := m.Encode()
	got, err := ParseManifest(text)
	if err != nil {
		t.Fatalf("reparse: %v", err)
	}
	if spec, _ := got.Get("util"); spec != path {
		t.Errorf("spec = %q, want %q", spec, path)
	}
	// Re-encoding must be stable: no backslash doubling across saves.
	if again := got.Encode(); again != text {
		t.Errorf("encode not stable:\n  first:  %q\n  second: %q", text, again)
	}
}

func TestManifestAddRejectsNameClash(t *testing.T) {
	m := &Manifest{}
	if err := m.Add("util", "github:alice/util"); err != nil {
		t.Fatalf("first Add: %v", err)
	}
	// Re-adding the same spec is idempotent.
	if err := m.Add("util", "github:alice/util"); err != nil {
		t.Fatalf("idempotent Add: %v", err)
	}
	// A different source under the same name must be rejected.
	err := m.Add("util", "github:bob/util")
	if err == nil {
		t.Fatalf("Add should reject a name clash")
	}
	if !strings.Contains(err.Error(), "--name") {
		t.Errorf("error should mention --name, got %q", err)
	}
	if spec, _ := m.Get("util"); spec != "github:alice/util" {
		t.Errorf("clashing Add must not overwrite, got %q", spec)
	}
	// A distinct name is fine.
	if err := m.Add("util_b", "github:bob/util"); err != nil {
		t.Fatalf("Add with alias: %v", err)
	}
	if len(m.Deps) != 2 {
		t.Errorf("deps = %v", m.Deps)
	}
}
