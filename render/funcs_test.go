package render

import (
	"bytes"
	"testing"
	"text/template"
)

func TestIsExported(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"Foo", true},
		{"fOO", false},
		{"", false},
		{"ABC", true},
	}
	for _, tc := range tests {
		if got := isExported(tc.in); got != tc.want {
			t.Errorf("isExported(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestIsGoKeyword(t *testing.T) {
	keywords := []string{
		"break", "case", "chan", "const", "continue",
		"default", "defer", "else", "fallthrough", "for",
		"func", "go", "goto", "if", "import",
		"interface", "map", "package", "range", "return",
		"select", "struct", "switch", "type", "var",
	}
	for _, kw := range keywords {
		if !isGoKeyword(kw) {
			t.Errorf("isGoKeyword(%q) = false, want true", kw)
		}
	}
	nonKeywords := []string{"Foo", "bar", "for1", ""}
	for _, s := range nonKeywords {
		if isGoKeyword(s) {
			t.Errorf("isGoKeyword(%q) = true, want false", s)
		}
	}
}

func TestDeref(t *testing.T) {
	tests := []struct{ in, want string }{
		{"*Foo", "Foo"},
		{"Foo", "Foo"},
		{"**Foo", "*Foo"},
	}
	for _, tc := range tests {
		if got := deref(tc.in); got != tc.want {
			t.Errorf("deref(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestTypeZeroVal(t *testing.T) {
	tests := []struct{ in, want string }{
		{"string", `""`},
		{"types.Email", `""`},
		{"bool", "false"},
		{"int", "0"},
		{"int32", "0"},
		{"int64", "0"},
		{"uint", "0"},
		{"uint32", "0"},
		{"uint64", "0"},
		{"float32", "0"},
		{"float64", "0"},
		{"time.Duration", "0"},
		{"uuid.UUID", "uuid.Nil"},
		{"url.URL", "url.URL{}"},
		{"time.Time", "time.Time{}"},
		{"civil.Date", "civil.Date{}"},
		{"net.IP", "nil"},
		{"Status", `""`},
	}
	for _, tc := range tests {
		if got := typeZeroVal(tc.in); got != tc.want {
			t.Errorf("typeZeroVal(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestTemplateFuncs exercises all registered template functions through a template.
func TestTemplateFuncs(t *testing.T) {
	tmplSrc2 := `{{lower "FOO"}} {{upper "foo"}} {{trimSpace "  hi  "}} {{hasPrefix "abc" "ab"}} {{hasSuffix "abc" "bc"}} {{contains "abc" "b"}} {{replace "a-b" "-" "_"}} {{isExported "Foo"}} {{isKeyword "func"}} {{deref "*T"}} {{add 3 2}} {{sub 5 2}} {{last 2 3}} {{titleCase "GET"}} {{titleCase ""}}`

	tmpl, err := template.New("test").Funcs(templateFuncs()).Parse(tmplSrc2)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	// spot-check a few results
	if got[:3] != "foo" {
		t.Errorf("lower result: %q", got[:3])
	}
}
