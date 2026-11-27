// This file is written by hand, not by the generator.
//
// It pins what the client decodes from an XML response: the attribute, the child elements, a list whose items stand
// side by side and one wrapped in an element of its own, and the elements only an error answer holds.

package golden

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/go-api-libs/api/apitest"
)

func TestGetStatement_XML(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		body string
		want FlexStatementResponse
	}{
		{
			name: "success",
			body: `<FlexStatementResponse timestamp='27 November, 2026 10:04 AM EST'>
<Status>Success</Status>
<ReferenceCode>0877097094</ReferenceCode>
<Url>https://example.com/GetStatement</Url>
<Note>first</Note>
<Note>second</Note>
<Books><Book>a</Book><Book>b</Book></Books>
</FlexStatementResponse>`,
			want: FlexStatementResponse{
				Timestamp:     "27 November, 2026 10:04 AM EST",
				Status:        "Success",
				ReferenceCode: "0877097094",
				URL:           "https://example.com/GetStatement",
				Note:          []string{"first", "second"},
				Books:         []string{"a", "b"},
			},
		},
		{
			name: "an error answer",
			body: `<FlexStatementResponse timestamp='27 November, 2026 10:05 AM EST'>
<Status>Fail</Status>
<ErrorCode>1019</ErrorCode>
<ErrorMessage>Statement generation in progress. Please try again shortly.</ErrorMessage>
</FlexStatementResponse>`,
			want: FlexStatementResponse{
				Timestamp:    "27 November, 2026 10:05 AM EST",
				Status:       "Fail",
				ErrorCode:    "1019",
				ErrorMessage: "Statement generation in progress. Please try again shortly.",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := NewClient(WithBaseURL(apitest.BodyServer(t, http.StatusOK, "text/xml;charset=UTF-8", []byte(tc.body))))
			if err != nil {
				t.Fatal(err)
			}

			got, err := c.GetStatement(t.Context(), nil)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(*got, tc.want) {
				t.Errorf("got %+v, want %+v", *got, tc.want)
			}
		})
	}
}
