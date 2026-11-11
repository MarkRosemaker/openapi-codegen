package ir_test

import (
	"slices"
	"testing"

	"github.com/MarkRosemaker/openapi-codegen/ir"
)

func TestResponses_Cases(t *testing.T) {
	problem := &ir.GoType{Name: "Problem"}
	ok := ir.Response{StatusCode: "200", GoConstant: "http.StatusOK", ContentType: "application/json", GoType: &ir.GoType{Name: "Page"}, IsSuccess: true}
	r401 := ir.Response{StatusCode: "401", GoConstant: "http.StatusUnauthorized", ContentType: "application/json", GoType: problem, Description: "who?"}
	r404 := ir.Response{StatusCode: "404", GoConstant: "http.StatusNotFound", ContentType: "application/json", GoType: problem, Description: "who?"}
	r409 := ir.Response{StatusCode: "409", GoConstant: "http.StatusConflict"}
	r422 := ir.Response{StatusCode: "422", GoConstant: "http.StatusUnprocessableEntity", ContentType: "application/json", GoType: &ir.GoType{Name: "Problem"}, Description: "bad"}
	other := ir.Response{StatusCode: "400", GoConstant: "http.StatusBadRequest", ContentType: "application/json", GoType: &ir.GoType{Name: "Other"}}
	r5xx := ir.Response{StatusCode: "5XX", GoConstant: "500", IsRange: true, ContentType: "application/json", GoType: problem}
	def := ir.Response{StatusCode: "default", GoConstant: "default", ContentType: "application/json", GoType: problem}

	labels := func(rs ir.Responses) (got []string) {
		for _, c := range rs.Cases() {
			if rs.HasRange() {
				got = append(got, c.Conditions())
			} else {
				got = append(got, c.GoConstants())
			}
		}

		return got
	}

	for _, tc := range []struct {
		name string
		rs   ir.Responses
		want []string
	}{
		{
			"codes alike share a case wherever they are, the default never",
			ir.Responses{ok, r401, r409, r422, other, def},
			[]string{"http.StatusOK", "http.StatusUnauthorized, http.StatusUnprocessableEntity", "http.StatusConflict", "http.StatusBadRequest", "default"},
		},
		{
			"with a range, only the response right after",
			ir.Responses{ok, r401, r404, r409, r5xx},
			[]string{
				"rsp.StatusCode == http.StatusOK",
				"rsp.StatusCode == http.StatusUnauthorized || rsp.StatusCode == http.StatusNotFound",
				"rsp.StatusCode == http.StatusConflict",
				"rsp.StatusCode >= 500 && rsp.StatusCode < 600",
			},
		},
	} {
		if got := labels(tc.rs); !slices.Equal(got, tc.want) {
			t.Errorf("%s:\ngot  %q\nwant %q", tc.name, got, tc.want)
		}
	}

	if got, want := (ir.Responses{r401, r404, r422}).Cases()[0].Descriptions(), "who?\nbad"; got != want {
		t.Errorf("descriptions: got %q, want %q", got, want)
	}
}
