package photon

import "testing"

func TestParseSpec(t *testing.T) {
	cases := []struct {
		spec                   string
		host, owner, repo, ref string
		url                    string
		local                  bool
	}{
		{"github:user/repo", "github.com", "user", "repo", "", "https://github.com/user/repo.git", false},
		{"github:user/repo@v1.2.3", "github.com", "user", "repo", "v1.2.3", "https://github.com/user/repo.git", false},
		{"user/repo", "github.com", "user", "repo", "", "https://github.com/user/repo.git", false},
		{"user/repo@main", "github.com", "user", "repo", "main", "https://github.com/user/repo.git", false},
		{"https://gitlab.com/grp/proj@dev", "gitlab.com", "grp", "proj", "dev", "https://gitlab.com/grp/proj.git", false},
		{"git@github.com:user/repo.git", "github.com", "user", "repo", "", "git@github.com:user/repo.git", false},
		{"git@github.com:user/repo.git@v2", "github.com", "user", "repo", "v2", "git@github.com:user/repo.git", false},
	}
	for _, c := range cases {
		s, err := ParseSpec(c.spec)
		if err != nil {
			t.Errorf("ParseSpec(%q): %v", c.spec, err)
			continue
		}
		if s.Host != c.host || s.Owner != c.owner || s.Repo != c.repo || s.Ref != c.ref || s.URL != c.url || s.IsLocal() != c.local {
			t.Errorf("ParseSpec(%q) = %+v", c.spec, s)
		}
	}
}

func TestParseSpecLocal(t *testing.T) {
	s, err := ParseSpec("./some/dir@v1")
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	if !s.IsLocal() {
		t.Fatalf("expected local, got %+v", s)
	}
	if s.Repo != "dir" {
		t.Errorf("Repo = %q", s.Repo)
	}
	if s.Ref != "v1" {
		t.Errorf("Ref = %q", s.Ref)
	}
	if s.CacheKey() == "" {
		t.Errorf("empty cache key")
	}
}

func TestParseSpecErrors(t *testing.T) {
	for _, bad := range []string{"", "   ", "github:justrepo"} {
		if _, err := ParseSpec(bad); err == nil {
			t.Errorf("ParseSpec(%q) should fail", bad)
		}
	}
}
