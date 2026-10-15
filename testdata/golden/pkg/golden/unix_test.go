// This file is written by hand, not by the generator.
//
// It pins how dates and date-times that are integers of unix time are written among those that are strings: each
// as its own schema says, also through an embedded part and where a request leaves members out.

package golden

import (
	"encoding/json/v2"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/civil"
)

func TestUnix_AmongStrings(t *testing.T) {
	t.Parallel()

	day := civil.Date{Year: 2025, Month: time.September, Day: 30}
	created := time.Date(2025, 9, 30, 12, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		name string
		in   any
		opts json.Options
		want string
	}{
		{
			name: "own fields",
			in:   DateWrapper{Raw: day, Fmt: "2025-09-30"},
			opts: jsonOpts,
			want: `{"fmt":"2025-09-30","raw":1759190400}`,
		},
		{
			name: "a string date-time stays a string",
			in:   Meta{Created: created, Expires: day},
			opts: jsonOpts,
			want: `{"created":"2025-09-30T12:00:00Z","expires":1759190400}`,
		},
		{
			name: "through an embedded part",
			in:   Note{Meta: Meta{Created: created, Expires: day}, Note: "n"},
			opts: jsonOpts,
			want: `{"created":"2025-09-30T12:00:00Z","note":"n","expires":1759190400}`,
		},
		{
			name: "a request leaves readOnly out",
			in:   Note{Meta: Meta{Created: created, Expires: day}, Note: "n"},
			opts: jsonOptsRequest,
			want: `{"note":"n","expires":1759190400}`,
		},
		{
			name: "an unset optional date is left out",
			in:   Meta{Created: created},
			opts: jsonOpts,
			want: `{"created":"2025-09-30T12:00:00Z"}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := json.Marshal(tc.in, tc.opts)
			if err != nil {
				t.Fatal(err)
			}

			if string(b) != tc.want {
				t.Fatalf("got %s, want %s", b, tc.want)
			}

			if tc.opts == jsonOptsRequest {
				return // what it left out does not come back
			}

			back := reflect.New(reflect.TypeOf(tc.in))
			if err := json.Unmarshal(b, back.Interface(), jsonOpts); err != nil {
				t.Fatal(err)
			}

			if got := back.Elem().Interface(); !reflect.DeepEqual(got, tc.in) {
				t.Errorf("read back %+v, want %+v", got, tc.in)
			}
		})
	}
}
