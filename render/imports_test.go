package render

import "testing"

func TestWriteImports(t *testing.T) {
	for _, tc := range []struct {
		name, src, want string
	}{
		{
			name: "adds what is used, the standard library first, each group sorted",
			src: `package p

func f() (*url.URL, error) {
	_ = api.NewErrCustom
	_ = civil.Date{}
	return nil, errors.New(strconv.Itoa(http.StatusOK))
}
`,
			want: `package p

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"cloud.google.com/go/civil"
	"github.com/go-api-libs/api"
)

func f() (*url.URL, error) {
	_ = api.NewErrCustom
	_ = civil.Date{}
	return nil, errors.New(strconv.Itoa(http.StatusOK))
}
`,
		},
		{
			name: "drops what is not used, and replaces imports that guessed wrong",
			src: `package p

import (
	"encoding/json"
	"fmt"
	"strings"
)

import "os"

func f(b []byte, v any) error { return json.Unmarshal(b, v) }
`,
			want: `package p

import (
	"encoding/json/v2"
)

func f(b []byte, v any) error { return json.Unmarshal(b, v) }
`,
		},
		{
			name: "a variable of a package's name is no use of it",
			src: `package p

func f(url struct{ Host string }) string { return url.Host }
`,
			want: `package p

func f(url struct{ Host string }) string { return url.Host }
`,
		},
		{
			name: "a name it does not know is left to the compiler",
			src: `package p

func f() { other.Do() }
`,
			want: `package p

func f() { other.Do() }
`,
		},
		{
			name: "keeps the comments around",
			src: `// Code generated. DO NOT EDIT.

// Package p is p.
package p

// f is f.
func f() string { return fmt.Sprint(1) }
`,
			want: `// Code generated. DO NOT EDIT.

// Package p is p.
package p

import (
	"fmt"
)

// f is f.
func f() string { return fmt.Sprint(1) }
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := writeImports("p.go", []byte(tc.src))
			if err != nil {
				t.Fatal(err)
			}

			if string(got) != tc.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tc.want)
			}
		})
	}

	if _, err := writeImports("p.go", []byte("package p\n\nfunc {")); err == nil {
		t.Error("source that does not parse: got no error")
	}
}
