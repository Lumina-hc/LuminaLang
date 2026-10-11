package compiler

import (
	"reflect"
	"testing"
)

func TestDeclRoundTrip(t *testing.T) {
	funcs := []lumcFunc{
		{Name: "area", Params: []lumcParam{{Name: "w", Type: "i32"}, {Name: "h", Type: "i32"}}, Ret: "i32"},
		{Name: "name", Ret: "string"},
		{Name: "total", Params: []lumcParam{{Name: "xs", Type: "[]i32"}}, Ret: "f64"},
	}
	text := WriteDecl("geometry", funcs)
	module, got, err := ParseDecl(text)
	if err != nil {
		t.Fatalf("ParseDecl: %v", err)
	}
	if module != "geometry" {
		t.Errorf("module = %q", module)
	}
	if !reflect.DeepEqual(got, funcs) {
		t.Errorf("round trip mismatch:\n got %+v\nwant %+v", got, funcs)
	}
}

func TestParseDeclErrors(t *testing.T) {
	for _, bad := range []string{
		"fn add(a i32) i32\n",
		"whatever\n",
		"fn (a: i32) i32\n",
	} {
		if _, _, err := ParseDecl(bad); err == nil {
			t.Errorf("ParseDecl(%q) should fail", bad)
		}
	}
}
