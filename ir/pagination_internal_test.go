package ir

import (
	"encoding/json/jsontext"
	"strings"
	"testing"
)

func TestReadPaginationExt(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		ext     string
		want    paginationExt
		wantErr bool
	}{
		{"", paginationExt{}, false},
		{`{"x-other":1}`, paginationExt{}, false},
		{`{"x-pagination":false}`, paginationExt{set: true, off: true}, false},
		{`{"x-pagination":{"cursor":"after","items":"entries"}}`, paginationExt{set: true, Cursor: "after", Items: "entries"}, false},
		{`{"x-pagination":{"cursr":"after"}}`, paginationExt{}, true},
	} {
		got, err := readPaginationExt(jsontext.Value(tc.ext))
		if (err != nil) != tc.wantErr || got != tc.want {
			t.Errorf("%s: got %+v, %v; want %+v, error %v", tc.ext, got, err, tc.want, tc.wantErr)
		}
	}
}

// listOp is an operation that takes a cursor of the type as a query parameter and answers with a list L, whose next
// cursor is a string.
func listOp(cursorType string, ext paginationExt) Operation {
	return Operation{
		Name:            "ListThings",
		ParamStructName: "ListThingsParams",
		QueryParams:     Params{{JSONName: "start_cursor", FieldName: "StartCursor", Type: cursorType}},
		SuccessReturn:   &GoType{Name: "L"},
		pagination:      ext,
	}
}

var listSchemas = []Schema{{Name: "L", Kind: SchemaKindStruct, Fields: []Field{
	{Name: "Results", JSONName: "results", Type: "[]Thing"},
	{Name: "NextCursor", JSONName: "next_cursor", Type: "string"},
	{Name: "HasMore", JSONName: "has_more", Type: "bool"},
}}}

func TestPaginate(t *testing.T) {
	t.Parallel()

	ops := []Operation{listOp("string", paginationExt{})}
	if err := paginate(ops, listSchemas); err != nil {
		t.Fatal(err)
	}

	p := ops[0].Pagination
	if p == nil {
		t.Fatal("a list of results, a next cursor and has_more is not paged through")
	}

	if want := (Pagination{
		CursorType: "string", NextField: "NextCursor", MoreField: "HasMore", ItemsField: "Results", ItemType: "Thing",
		Prepare: "var p ListThingsParams\nif params != nil {\np = *params\n}\np.StartCursor = cursor", Args: ", &p",
	}); *p != want {
		t.Errorf("got %+v\nwant %+v", *p, want)
	}

	// a cursor of another type than the list's is no cursor of the list, unless x-pagination says it is
	ops = []Operation{listOp("int", paginationExt{})}
	if err := paginate(ops, listSchemas); err != nil || ops[0].Pagination != nil {
		t.Errorf("a cursor of another type: got %+v, %v; want none", ops[0].Pagination, err)
	}

	ops = []Operation{listOp("int", paginationExt{set: true})}
	if err := paginate(ops, listSchemas); err == nil || !strings.Contains(err.Error(), "ListThings: x-pagination") {
		t.Errorf("x-pagination on a cursor of another type: got %v, want its error", err)
	}

	ops = []Operation{listOp("string", paginationExt{set: true, off: true})}
	if err := paginate(ops, listSchemas); err != nil || ops[0].Pagination != nil {
		t.Errorf("x-pagination false: got %+v, %v; want none", ops[0].Pagination, err)
	}

	// a named list of items is followed to its items
	named := []Schema{
		{Name: "L", Kind: SchemaKindStruct, Fields: []Field{
			{Name: "Things", JSONName: "things", Type: "Things"},
			{Name: "NextCursor", JSONName: "next_cursor", Type: "string"},
		}},
		{Name: "Things", Kind: SchemaKindAlias, Type: "[]Thing"},
	}

	ops = []Operation{listOp("string", paginationExt{set: true, Items: "things"})}
	if err := paginate(ops, named); err != nil || ops[0].Pagination == nil || ops[0].Pagination.ItemType != "Thing" {
		t.Errorf("a named list: got %+v, %v; want items of Thing", ops[0].Pagination, err)
	}

	ops = []Operation{listOp("string", paginationExt{}), {Name: "ListThingsAll"}}
	if err := paginate(ops, listSchemas); err != nil || ops[0].Pagination != nil {
		t.Errorf("a name taken: got %+v, %v; want none", ops[0].Pagination, err)
	}
}
