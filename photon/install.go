package photon

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Installer resolves a manifest's dependencies and installs them into a
// project's packages/ directory, fetching from decentralised git remotes.
type Installer struct {
	ProjectDir string // project root (contains lumina.toml and packages/)
	CacheDir   string // global download cache
	Out        io.Writer
}

func (in *Installer) logf(format string, args ...any) {
	if in.Out == nil {
		return
	}
	fmt.Fprintf(in.Out, format+"\n", args...)
}

// PackagesDir is where installed packages live for this project.
func (in *Installer) PackagesDir() string {
	return filepath.Join(in.ProjectDir, "packages")
}

// Install resolves the dependency graph (including transitive dependencies)
// and installs every package. It returns the resolved lockfile.
func (in *Installer) Install(root *Manifest) (*Lock, error) {
	lock := &Lock{}
	seen := map[string]LockEntry{}
	queue := append([]Dep(nil), root.Deps...)
	for len(queue) > 0 {
		dep := queue[0]
		queue = queue[1:]
		src, err := ParseSpec(dep.Spec)
		if err != nil {
			return nil, err
		}
		if prev, ok := seen[dep.Name]; ok {
			if prev.Source != dep.Spec {
				return nil, fmt.Errorf("package %q is required both as %q and %q", dep.Name, prev.Source, dep.Spec)
			}
			continue
		}
		commit, dir, err := in.fetch(src)
		if err != nil {
			return nil, fmt.Errorf("fetching %s: %w", dep.Name, err)
		}
		entry := LockEntry{Name: dep.Name, Source: dep.Spec, Ref: src.Ref, Commit: commit}
		seen[dep.Name] = entry
		lock.Set(entry)
		in.logf("installing %s (%s)", dep.Name, shortCommit(commit))
		if err := in.installPackage(dep.Name, dir); err != nil {
			return nil, err
		}
		// Read the package's own manifest to resolve transitive dependencies.
		sub, err := readManifestFile(filepath.Join(dir, "lumina.toml"))
		if err != nil {
			return nil, fmt.Errorf("reading manifest of %s: %w", dep.Name, err)
		}
		if sub != nil {
			queue = append(queue, sub.Deps...)
		}
	}
	return lock, nil
}

func shortCommit(c string) string {
	if len(c) > 8 {
		return c[:8]
	}
	return c
}

// fetch returns the resolved commit and a directory holding that revision,
// cloning into the cache when necessary.
func (in *Installer) fetch(src Source) (commit, dir string, err error) {
	if src.IsLocal() {
		if !DirExists(src.Local) {
			return "", "", fmt.Errorf("local path does not exist: %s", src.Local)
		}
		commit, err = gitRevParse(src.Local, src.Ref)
		if err != nil {
			return "", "", err
		}
		return commit, src.Local, nil
	}
	commit, err = gitLsRemote(src.URL, src.Ref)
	if err != nil {
		return "", "", err
	}
	dir = filepath.Join(in.CacheDir, src.CacheKey(), commit)
	if !DirExists(dir) {
		if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
			return "", "", err
		}
		in.logf("cloning %s", src.URL)
		if err := gitClone(src.URL, dir, commit); err != nil {
			os.RemoveAll(dir)
			return "", "", err
		}
	}
	return commit, dir, nil
}

// installPackage copies a fetched revision into <project>/packages/<name>.
func (in *Installer) installPackage(name, srcDir string) error {
	dst := filepath.Join(in.PackagesDir(), name)
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	return copyTree(srcDir, dst)
}

func readManifestFile(path string) (*Manifest, error) {
	if !FileExists(path) {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseManifest(string(data))
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		if rel == ".git" || strings.HasPrefix(rel, ".git"+string(filepath.Separator)) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return string(out), nil
}

func gitLsRemote(url, ref string) (string, error) {
	if ref == "" {
		ref = "HEAD"
	}
	out, err := runGit("ls-remote", url, ref)
	if err != nil {
		return "", err
	}
	if sha := firstSHA(out); sha != "" {
		return sha, nil
	}
	// Fall back to a full listing (handles short refs and refs with slashes).
	full, err := runGit("ls-remote", url)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(full, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		name := fields[1]
		if name == ref || strings.HasSuffix(name, "/"+ref) {
			return fields[0], nil
		}
	}
	if looksLikeCommit(ref) {
		return ref, nil
	}
	return "", fmt.Errorf("revision %q not found at %s", ref, url)
}

func gitRevParse(dir, ref string) (string, error) {
	if ref == "" {
		ref = "HEAD"
	}
	out, err := runGit("-C", dir, "rev-parse", ref)
	if err != nil {
		return "", err
	}
	sha := strings.TrimSpace(out)
	if sha == "" {
		return "", fmt.Errorf("cannot resolve revision %q in %s", ref, dir)
	}
	return sha, nil
}

func gitClone(url, dir, commit string) error {
	if _, err := runGit("clone", "--quiet", url, dir); err != nil {
		return err
	}
	if _, err := runGit("-C", dir, "checkout", "--quiet", commit); err != nil {
		return err
	}
	return nil
}

func firstSHA(out string) string {
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 1 && looksLikeCommit(fields[0]) {
			return fields[0]
		}
	}
	return ""
}

func looksLikeCommit(s string) bool {
	if len(s) < 7 || len(s) > 40 {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

// DefaultCacheDir returns the global photon cache directory.
func DefaultCacheDir() (string, error) {
	if v := os.Getenv("PHOTON_HOME"); v != "" {
		return filepath.Join(v, "cache"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".photon", "cache"), nil
}
