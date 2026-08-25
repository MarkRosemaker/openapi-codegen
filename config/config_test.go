package config

import "testing"

func TestShouldGenerate_AllFalse(t *testing.T) {
	cfg := Generate{}
	for _, f := range []string{"types.gen.go", "client.gen.go", "client.gen_test.go"} {
		if !cfg.ShouldGenerate(f) {
			t.Errorf("ShouldGenerate(%q) = false, want true", f)
		}
	}

	for _, f := range []string{"server.gen.go", "other.go"} {
		if cfg.ShouldGenerate(f) {
			t.Errorf("ShouldGenerate(%q) = true, want false", f)
		}
	}
}

func TestShouldGenerate_Selective(t *testing.T) {
	cfg := Generate{Types: true, Server: true}
	want := map[string]bool{
		"types.gen.go":       true,
		"client.gen.go":      false,
		"client.gen_test.go": false,
		"server.gen.go":      true,
		"other.go":           false,
	}
	for f, w := range want {
		if got := cfg.ShouldGenerate(f); got != w {
			t.Errorf("ShouldGenerate(%q) = %v, want %v", f, got, w)
		}
	}
}
