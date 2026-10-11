package photon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Source describes where a dependency lives and which revision to use.
type Source struct {
	Spec  string // the original spec string
	Host  string // e.g. "github.com" (empty for a local path)
	Owner string // e.g. "user" (empty for a local path)
	Repo  string // repository name, without ".git"
	URL   string // git clone URL (empty for a local path)
	Local string // local filesystem path, if this is a local dependency
	Ref   string // tag, branch or commit; empty means the remote default
}

// IsLocal reports whether the source is a local filesystem path.
func (s Source) IsLocal() bool { return s.Local != "" }

// ParseSpec parses a dependency spec into a Source. Supported forms:
//
//	github:owner/repo[@ref]
//	owner/repo[@ref]                  (assumed GitHub)
//	https://host/owner/repo[@ref]
//	git@host:owner/repo[@ref]
//	./relative/path[@ref]             (local)
func ParseSpec(spec string) (Source, error) {
	raw := strings.TrimSpace(spec)
	if raw == "" {
		return Source{}, fmt.Errorf("empty dependency spec")
	}
	base, ref := splitRef(raw)
	s := Source{Spec: raw, Ref: ref}

	switch {
	case strings.HasPrefix(base, "github:"):
		s.Host = "github.com"
		owner, repo, err := splitOwnerRepo(strings.TrimPrefix(base, "github:"))
		if err != nil {
			return Source{}, err
		}
		s.Owner, s.Repo = owner, repo
		s.URL = "https://github.com/" + owner + "/" + repo + ".git"
	case strings.HasPrefix(base, "git@"):
		// scp-like: git@host:owner/repo
		colon := strings.IndexByte(base, ':')
		if colon < 0 {
			return Source{}, fmt.Errorf("invalid git ssh spec %q", base)
		}
		s.Host = base[len("git@"):colon]
		owner, repo, err := splitOwnerRepo(base[colon+1:])
		if err != nil {
			return Source{}, err
		}
		s.Owner, s.Repo = owner, repo
		s.URL = "git@" + s.Host + ":" + owner + "/" + repo + ".git"
	case strings.Contains(base, "://"):
		s.URL = strings.TrimSuffix(base, ".git")
		host, owner, repo := parseURLParts(base)
		s.Host, s.Owner, s.Repo = host, owner, repo
		s.URL = base
		if !strings.HasSuffix(s.URL, ".git") {
			s.URL += ".git"
		}
	default:
		// owner/repo (GitHub) or a local path.
		if owner, repo, err := splitOwnerRepo(base); err == nil && looksLikeGithub(base) {
			s.Host = "github.com"
			s.Owner, s.Repo = owner, repo
			s.URL = "https://github.com/" + owner + "/" + repo + ".git"
		} else {
			abs, err := filepath.Abs(base)
			if err != nil {
				return Source{}, err
			}
			s.Local = abs
			s.Repo = filepath.Base(abs)
		}
	}
	if s.Repo == "" {
		return Source{}, fmt.Errorf("cannot determine package name from %q", raw)
	}
	return s, nil
}

func looksLikeGithub(base string) bool {
	if strings.HasPrefix(base, ".") || strings.HasPrefix(base, "/") || strings.HasPrefix(base, "\\") {
		return false
	}
	if filepath.VolumeName(base) != "" {
		return false
	}
	parts := strings.Split(base, "/")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

func splitOwnerRepo(s string) (string, string, error) {
	s = strings.Trim(s, "/")
	s = strings.TrimSuffix(s, ".git")
	parts := strings.Split(s, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("expected owner/repo, got %q", s)
	}
	return parts[len(parts)-2], parts[len(parts)-1], nil
}

func parseURLParts(raw string) (host, owner, repo string) {
	u := raw
	if i := strings.Index(u, "://"); i >= 0 {
		u = u[i+3:]
	}
	if i := strings.IndexByte(u, '/'); i >= 0 {
		host = u[:i]
		rest := u[i+1:]
		if j := strings.LastIndexByte(rest, '/'); j >= 0 {
			owner = rest[:j]
			repo = strings.TrimSuffix(rest[j+1:], ".git")
		} else {
			repo = strings.TrimSuffix(rest, ".git")
		}
	}
	return host, owner, repo
}

// splitRef separates a trailing "@ref" from a spec, taking care not to eat the
// "@" of an scp-like "git@host:..." prefix.
func splitRef(raw string) (base, ref string) {
	at := strings.LastIndexByte(raw, '@')
	if at < 0 {
		return raw, ""
	}
	if strings.HasPrefix(raw, "git@") && at == len("git") {
		return raw, ""
	}
	return raw[:at], raw[at+1:]
}

// CacheKey is a filesystem-safe key for the source inside the global cache.
func (s Source) CacheKey() string {
	if s.IsLocal() {
		return "local/" + sanitize(filepath.Base(s.Local))
	}
	return sanitize(s.Host) + "/" + sanitize(s.Owner) + "/" + sanitize(s.Repo)
}

func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			return '_'
		}
		return r
	}, s)
}

// FileExists reports whether path exists as a regular file.
func FileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

// DirExists reports whether path exists as a directory.
func DirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}
