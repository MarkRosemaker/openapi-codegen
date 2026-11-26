package ir

import "testing"

func TestZeroVal(t *testing.T) {
	tests := []struct{ Type, want string }{
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
		{"uuid.UUID", "uuid.Nil()"},
		{"url.URL", "url.URL{}"},
		{"time.Time", "time.Time{}"},
		{"civil.Date", "civil.Date{}"},
		{"net.IP", "nil"},
		{"Status", `""`},
	}
	for _, tc := range tests {
		p := Param{Type: tc.Type}
		if got := p.ZeroVal(); got != tc.want {
			t.Errorf("typeZeroVal(%q) = %q, want %q", tc.Type, got, tc.want)
		}
	}
}
