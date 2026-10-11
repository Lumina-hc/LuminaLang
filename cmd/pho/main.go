// Command pho is the LuminaLang package manager (photon).
//
// It is decentralised: a dependency is just a git repository, usually on
// GitHub, addressed as github:owner/repo@ref. There is no central registry.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"luminalang/photon"
)

const version = "0.1.0"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}
	cmd := args[0]
	rest := args[1:]
	var err error
	switch cmd {
	case "-h", "--help", "help":
		usage()
		return
	case "-v", "--version", "version":
		fmt.Printf("photon v%s\n", version)
		return
	case "init":
		err = cmdInit(rest)
	case "add":
		err = cmdAdd(rest)
	case "remove", "rm":
		err = cmdRemove(rest)
	case "install", "i":
		err = cmdInstall(rest)
	case "list", "ls":
		err = cmdList(rest)
	case "build":
		err = cmdBuild(rest)
	default:
		fmt.Fprintf(os.Stderr, "pho: unknown command %q\n", cmd)
		usage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "pho: "+err.Error())
		os.Exit(1)
	}
}

func usage() {
	fmt.Printf(`photon v%s - the LuminaLang package manager

Usage: pho <command> [options]

Commands:
  init [name]        create lumina.toml and src/main.lum
  add <spec>         add a dependency (github:owner/repo@ref) and install
  remove <name>      remove a dependency
  install            resolve and install all dependencies
  list               list declared dependencies
  build [-- args]    compile src/main.lum with luminalang

A package is a git repository (one package per repository); dependencies are
written as github:owner/repo@ref. No central registry is involved.
`, version)
}

func die(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// findProject walks up from the current directory looking for lumina.toml.
func findProject() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if photon.FileExists(filepath.Join(dir, "lumina.toml")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", die("no lumina.toml found (run 'pho init')")
		}
		dir = parent
	}
}

func loadManifest(projDir string) (*photon.Manifest, error) {
	data, err := os.ReadFile(filepath.Join(projDir, "lumina.toml"))
	if err != nil {
		return nil, err
	}
	return photon.ParseManifest(string(data))
}

func saveManifest(projDir string, m *photon.Manifest) error {
	return os.WriteFile(filepath.Join(projDir, "lumina.toml"), []byte(m.Encode()), 0o644)
}

func newInstaller(projDir string) (*photon.Installer, error) {
	cache, err := photon.DefaultCacheDir()
	if err != nil {
		return nil, err
	}
	return &photon.Installer{ProjectDir: projDir, CacheDir: cache, Out: os.Stdout}, nil
}

func cmdInit(args []string) error {
	dir, _ := os.Getwd()
	name := filepath.Base(dir)
	if len(args) > 0 {
		name = args[0]
	}
	mfPath := filepath.Join(dir, "lumina.toml")
	if photon.FileExists(mfPath) {
		return die("lumina.toml already exists")
	}
	m := &photon.Manifest{
		Name:    name,
		Version: "0.1.0",
		Modules: []string{name},
	}
	if err := saveManifest(dir, m); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil {
		return err
	}
	mainPath := filepath.Join(dir, "src", "main.lum")
	if !photon.FileExists(mainPath) {
		main := "fn main() i32 {\n\tout \"hello from " + name + "\"\n\tback 0\n}\n"
		if err := os.WriteFile(mainPath, []byte(main), 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("created lumina.toml and src/main.lum for %q\n", name)
	return nil
}

func cmdAdd(args []string) error {
	if len(args) == 0 {
		return die("usage: pho add <spec> [--name <name>]")
	}
	spec := ""
	name := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--name":
			if i+1 >= len(args) {
				return die("--name requires a value")
			}
			name = args[i+1]
			i++
		default:
			if spec == "" {
				spec = args[i]
			}
		}
	}
	if spec == "" {
		return die("usage: pho add <spec> [--name <name>]")
	}
	src, err := photon.ParseSpec(spec)
	if err != nil {
		return err
	}
	if name == "" {
		name = src.Repo
	}
	projDir, err := findProject()
	if err != nil {
		return err
	}
	m, err := loadManifest(projDir)
	if err != nil {
		return err
	}
	if err := m.Add(name, spec); err != nil {
		return err
	}
	if err := saveManifest(projDir, m); err != nil {
		return err
	}
	fmt.Printf("added %s = %q\n", name, spec)
	return runInstall(projDir, m)
}

func cmdRemove(args []string) error {
	if len(args) == 0 {
		return die("usage: pho remove <name>")
	}
	projDir, err := findProject()
	if err != nil {
		return err
	}
	m, err := loadManifest(projDir)
	if err != nil {
		return err
	}
	if !m.Remove(args[0]) {
		return die("dependency %q not found", args[0])
	}
	if err := saveManifest(projDir, m); err != nil {
		return err
	}
	fmt.Printf("removed %s\n", args[0])
	return nil
}

func cmdInstall(_ []string) error {
	projDir, err := findProject()
	if err != nil {
		return err
	}
	m, err := loadManifest(projDir)
	if err != nil {
		return err
	}
	return runInstall(projDir, m)
}

func runInstall(projDir string, m *photon.Manifest) error {
	in, err := newInstaller(projDir)
	if err != nil {
		return err
	}
	lock, err := in.Install(m)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(projDir, "photon.lock"), []byte(lock.Encode()), 0o644); err != nil {
		return err
	}
	if len(lock.Entries) == 0 {
		fmt.Println("no dependencies")
	} else {
		fmt.Printf("installed %d package(s)\n", len(lock.Entries))
	}
	return nil
}

func cmdList(_ []string) error {
	projDir, err := findProject()
	if err != nil {
		return err
	}
	m, err := loadManifest(projDir)
	if err != nil {
		return err
	}
	if len(m.Deps) == 0 {
		fmt.Println("no dependencies")
		return nil
	}
	deps := append([]photon.Dep(nil), m.Deps...)
	sort.SliceStable(deps, func(i, j int) bool { return deps[i].Name < deps[j].Name })
	for _, d := range deps {
		fmt.Printf("%-20s %s\n", d.Name, d.Spec)
	}
	return nil
}

func cmdBuild(args []string) error {
	projDir, err := findProject()
	if err != nil {
		return err
	}
	// Collect module search directories from installed packages.
	paths := moduleDirs(filepath.Join(projDir, "packages"))
	entry := filepath.Join(projDir, "src", "main.lum")

	bin, err := findLuminalang()
	if err != nil {
		return err
	}
	cmdArgs := []string{entry}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command(bin, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	env := os.Environ()
	if len(paths) > 0 {
		sep := string(os.PathListSeparator)
		env = append(env, "LUMINA_PATH="+strings.Join(paths, sep))
	}
	cmd.Env = env
	return cmd.Run()
}

// moduleDirs returns the directories inside packages/ that may hold modules:
// each package's src/ and include/, plus the package root.
func moduleDirs(pkgsDir string) []string {
	entries, err := os.ReadDir(pkgsDir)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		base := filepath.Join(pkgsDir, e.Name())
		dirs = append(dirs, base)
		for _, sub := range []string{"src", "include"} {
			if photon.DirExists(filepath.Join(base, sub)) {
				dirs = append(dirs, filepath.Join(base, sub))
			}
		}
	}
	return dirs
}

func findLuminalang() (string, error) {
	if v := os.Getenv("LUMINA_BIN"); v != "" {
		return v, nil
	}
	name := "luminalang"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if exe, err := os.Executable(); err == nil {
		cand := filepath.Join(filepath.Dir(exe), name)
		if photon.FileExists(cand) {
			return cand, nil
		}
	}
	if p, err := exec.LookPath(name); err == nil {
		return p, nil
	}
	return "", die("cannot find %q (set LUMINA_BIN or put it on PATH)", name)
}
