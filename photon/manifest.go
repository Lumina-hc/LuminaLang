// Package photon implements the LuminaLang package manager format: the
// lumina.toml manifest, the photon.lock lockfile, and the installer that
// resolves and fetches packages from decentralised GitHub repositories.
package photon

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Dep is a single declared dependency, kept in file order.
type Dep struct {
	Name string
	Spec string
}

// Manifest is the parsed form of a lumina.toml file.
type Manifest struct {
	Name        string
	Version     string
	Description string
	License     string
	Authors     []string

	Modules []string // modules the package exports ([lib] modules)

	Deps []Dep // ordered dependencies
}

// Get returns the dependency spec for name, if declared.
func (m *Manifest) Get(name string) (string, bool) {
	for _, d := range m.Deps {
		if d.Name == name {
			return d.Spec, true
		}
	}
	return "", false
}

// Set adds or replaces a dependency, preserving order.
func (m *Manifest) Set(name, spec string) {
	for i := range m.Deps {
		if m.Deps[i].Name == name {
			m.Deps[i].Spec = spec
			return
		}
	}
	m.Deps = append(m.Deps, Dep{Name: name, Spec: spec})
}

// Add records a dependency like Set, but refuses to silently replace a name
// that is already taken by a different spec: two repositories that share a
// name must be distinguished with an explicit alias (--name).
func (m *Manifest) Add(name, spec string) error {
	if existing, ok := m.Get(name); ok && existing != spec {
		return fmt.Errorf("dependency %q already exists as %q; use --name <alias> to add %q under a different name, or remove it first", name, existing, spec)
	}
	m.Set(name, spec)
	return nil
}

// Remove deletes a dependency. It reports whether the dependency existed.
func (m *Manifest) Remove(name string) bool {
	for i := range m.Deps {
		if m.Deps[i].Name == name {
			m.Deps = append(m.Deps[:i], m.Deps[i+1:]...)
			return true
		}
	}
	return false
}

// ParseManifest parses a lumina.toml. The supported subset of TOML is
// deliberately small: quoted strings, arrays of quoted strings, and the
// sections [package], [lib] and [dependencies].
func ParseManifest(data string) (*Manifest, error) {
	m := &Manifest{}
	section := ""
	lines := strings.Split(data, "\n")
	for i, raw := range lines {
		line := stripComment(raw)
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		key, val, ok := splitKeyValue(line)
		if !ok {
			return nil, fmt.Errorf("lumina.toml:%d: expected key = value", i+1)
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch section {
		case "package":
			if key == "authors" {
				list, err := parseStringList(val)
				if err != nil {
					return nil, fmt.Errorf("lumina.toml:%d: %v", i+1, err)
				}
				m.Authors = list
				break
			}
			s, err := parseString(val)
			if err != nil {
				return nil, fmt.Errorf("lumina.toml:%d: %v", i+1, err)
			}
			switch key {
			case "name":
				m.Name = s
			case "version":
				m.Version = s
			case "description":
				m.Description = s
			case "license":
				m.License = s
			}
		case "lib":
			if key == "modules" {
				list, err := parseStringList(val)
				if err != nil {
					return nil, fmt.Errorf("lumina.toml:%d: %v", i+1, err)
				}
				m.Modules = list
			}
		case "dependencies":
			s, err := parseString(val)
			if err != nil {
				return nil, fmt.Errorf("lumina.toml:%d: %v", i+1, err)
			}
			m.Deps = append(m.Deps, Dep{Name: unquoteKey(key), Spec: s})
		default:
			// Unknown section: ignore for forward compatibility.
		}
	}
	return m, nil
}

// Encode renders the manifest back to lumina.toml text.
func (m *Manifest) Encode() string {
	var b strings.Builder
	b.WriteString("[package]\n")
	fmt.Fprintf(&b, "name = %q\n", m.Name)
	if m.Version != "" {
		fmt.Fprintf(&b, "version = %q\n", m.Version)
	}
	if m.Description != "" {
		fmt.Fprintf(&b, "description = %q\n", m.Description)
	}
	if m.License != "" {
		fmt.Fprintf(&b, "license = %q\n", m.License)
	}
	if len(m.Authors) > 0 {
		fmt.Fprintf(&b, "authors = [%s]\n", quoteList(m.Authors))
	}

	if len(m.Modules) > 0 {
		b.WriteString("\n[lib]\n")
		fmt.Fprintf(&b, "modules = [%s]\n", quoteList(m.Modules))
	}

	b.WriteString("\n[dependencies]\n")
	deps := append([]Dep(nil), m.Deps...)
	sort.SliceStable(deps, func(i, j int) bool { return deps[i].Name < deps[j].Name })
	for _, d := range deps {
		fmt.Fprintf(&b, "%s = %q\n", d.Name, d.Spec)
	}
	return b.String()
}

func stripComment(line string) string {
	inStr := false
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case '"':
			inStr = !inStr
		case '#':
			if !inStr {
				return line[:i]
			}
		}
	}
	return line
}

func splitKeyValue(line string) (string, string, bool) {
	inStr := false
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case '"':
			inStr = !inStr
		case '=':
			if !inStr {
				return line[:i], line[i+1:], true
			}
		}
	}
	return "", "", false
}

func unquoteKey(k string) string {
	if len(k) >= 2 && (k[0] == '"' || k[0] == '\'') {
		return k[1 : len(k)-1]
	}
	return k
}

func parseString(v string) (string, error) {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && v[0] == '\'' && v[len(v)-1] == '\'' {
		return v[1 : len(v)-1], nil
	}
	if s, err := strconv.Unquote(v); err == nil {
		return s, nil
	}
	return "", fmt.Errorf("expected a quoted string, got %q", v)
}

func parseStringList(v string) ([]string, error) {
	v = strings.TrimSpace(v)
	if len(v) < 2 || v[0] != '[' || v[len(v)-1] != ']' {
		return nil, fmt.Errorf("expected a list, got %q", v)
	}
	inner := strings.TrimSpace(v[1 : len(v)-1])
	if inner == "" {
		return nil, nil
	}
	var out []string
	for _, part := range strings.Split(inner, ",") {
		s, err := parseString(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func quoteList(items []string) string {
	parts := make([]string, len(items))
	for i, s := range items {
		parts[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(parts, ", ")
}
