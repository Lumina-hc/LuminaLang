package photon

import "testing"

func TestLockRoundTrip(t *testing.T) {
	var l Lock
	l.Set(LockEntry{Name: "geometry", Source: "github:user/geometry@v0.1.0", Ref: "v0.1.0", Commit: "5b095d88"})
	l.Set(LockEntry{Name: "util", Source: `C:\Users\dev\a\util`, Commit: "abc123"})
	l.Set(LockEntry{Name: "geometry", Source: "changed", Commit: "newest"})

	got, err := ParseLock(l.Encode())
	if err != nil {
		t.Fatalf("ParseLock: %v", err)
	}
	if len(got.Entries) != 2 {
		t.Fatalf("entries = %v", got.Entries)
	}
	e, ok := got.Get("geometry")
	if !ok || e.Commit != "newest" {
		t.Errorf("geometry = %+v", e)
	}
	if u, ok := got.Get("util"); !ok || u.Source != `C:\Users\dev\a\util` {
		t.Errorf("util source = %q", u.Source)
	}
	if again := got.Encode(); again != l.Encode() {
		t.Errorf("lock encode not stable:\n  first:  %q\n  second: %q", l.Encode(), again)
	}
}

func TestParseLockIgnoresNoise(t *testing.T) {
	src := `# photon.lock
stray = "value"

[[package]]
name = "a"
commit = "deadbeef"

garbage line

[[package]]
name = "b"
commit = "cafef00d"
`
	l, err := ParseLock(src)
	if err != nil {
		t.Fatalf("ParseLock: %v", err)
	}
	if len(l.Entries) != 2 || l.Entries[0].Name != "a" || l.Entries[1].Name != "b" {
		t.Errorf("entries = %v", l.Entries)
	}
}
