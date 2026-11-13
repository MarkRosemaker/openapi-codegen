package ir_test

import (
	"strings"
	"testing"

	"github.com/MarkRosemaker/openapi"
	"github.com/MarkRosemaker/openapi-codegen/ir"
)

const querySpec = `{
  "openapi": "3.1.0",
  "info": {"title": "t", "version": "1"},
  "paths": {
    "/search": {
      "get": {
        "operationId": "search",
        "parameters": [
          {"name": "user_ids", "in": "query", "schema": {"type": "array", "items": {"type": "integer"}}},
          {"name": "tags", "in": "query", "explode": false, "schema": {"type": "array", "items": {"type": "string"}}},
          {"name": "words", "in": "query", "style": "spaceDelimited", "explode": false, "schema": {"type": "array", "items": {"type": "string"}}},
          {"name": "codes", "in": "query", "style": "pipeDelimited", "explode": false, "schema": {"type": "array", "items": {"type": "integer"}}},
          {"name": "sort", "in": "query", "style": "deepObject", "explode": true, "schema": {"type": "array", "items": {"type": "string"}}},
          {"name": "filter", "in": "query", "style": "deepObject", "explode": true, "schema": {
            "type": "object", "properties": {"status": {"type": "string"}, "min": {"type": "integer"}}}},
          {"name": "point", "in": "query", "required": true, "explode": false, "schema": {
            "type": "object", "required": ["x", "z"],
            "properties": {"x": {"type": "integer"}, "z": {"oneOf": [{"type": "integer"}, {"type": "null"}]}, "label": {"type": "string"}}}},
          {"name": "range", "in": "query", "schema": {
            "type": "object", "properties": {"from": {"type": "integer"}, "to": {"type": "integer"}}}},
          {"name": "path", "in": "query", "allowReserved": true, "schema": {"type": "string"}},
          {"name": "extra", "in": "query", "schema": {"type": "object", "additionalProperties": {"type": "integer"}}},
          {"name": "fixed", "in": "query", "required": true, "schema": {"type": "string", "const": "yes"}}
        ],
        "responses": {"204": {"description": "none"}}
      }
    },
    "/items/{slug}": {
      "get": {
        "operationId": "getItem",
        "parameters": [{"name": "slug", "in": "path", "required": true, "schema": {"type": "string"}}],
        "responses": {"204": {"description": "none"}}
      }
    },
    "/items/{slug}/{n}": {
      "get": {
        "operationId": "getItemVersion",
        "parameters": [
          {"name": "slug", "in": "path", "required": true, "schema": {"type": "string"}},
          {"name": "n", "in": "path", "required": true, "schema": {"type": "integer"}}
        ],
        "responses": {"204": {"description": "none"}}
      }
    },
    "/joined": {
      "get": {
        "operationId": "joined",
        "parameters": [{"name": "tags", "in": "query", "explode": false, "schema": {"type": "array", "items": {"type": "string"}}}],
        "responses": {"204": {"description": "none"}}
      }
    },
    "/plain": {
      "get": {
        "operationId": "plain",
        "parameters": [{"name": "q", "in": "query", "schema": {"type": "string"}}],
        "responses": {"204": {"description": "none"}}
      }
    }
  }
}`

type opDoc struct{ *ir.Document }

func (d opDoc) op(name string) ir.Operation {
	for _, op := range d.Operations {
		if op.Name == name {
			return op
		}
	}

	panic("no operation " + name)
}

func queryDoc(t *testing.T) opDoc {
	t.Helper()

	spec, err := openapi.LoadFromDataJSON([]byte(querySpec))
	if err != nil {
		t.Fatal(err)
	}

	doc, err := ir.FromDocument(spec, "t", "", false, false)
	if err != nil {
		t.Fatal(err)
	}

	return opDoc{doc}
}

// TestQueryParams checks that each style is written by the client, read by the server and written by api.js alike.
func TestQueryParams(t *testing.T) {
	doc := queryDoc(t)
	search, plain := doc.op("Search"), doc.op("Plain")
	encode, parse, js := search.EncodeQuery(), search.ParseArgs(), search.JSQuery()

	for _, tc := range []struct {
		name                 string
		client, server, jsOp string // a line of each that only this style writes
	}{
		{
			"exploded array, its field named as Go names initialisms",
			`q.Add("user_ids", strconv.Itoa(v))`, `for _, s := range query["user_ids"] {`,
			`for (const v of params.userIds ?? []) q.append("user_ids", String(v));`,
		},
		{
			"form",
			`escaped["tags"] = []string{joinQuery(slices.Clone(params.Tags), ",", url.QueryEscape)}`,
			`parts, err := queryParts(req.URL.RawQuery, "tags", ",")`, `raw.push("tags=" + join(params.tags, ",", esc));`,
		},
		{
			"spaceDelimited",
			`escaped["words"] = []string{joinQuery(slices.Clone(params.Words), "%20", url.QueryEscape)}`,
			`parts, err := queryParts(req.URL.RawQuery, "words", "%20")`, `raw.push("words=" + join(params.words, "%20", esc));`,
		},
		{
			"pipeDelimited, each element formatted",
			`escaped["codes"] = []string{joinQuery(parts, "|", url.QueryEscape)}`,
			`parts, err := queryParts(req.URL.RawQuery, "codes", "|")`, `raw.push("codes=" + join(params.codes, "|", esc));`,
		},
		{
			"deepObject of an array, as form exploded",
			`q.Add("sort", v)`, `params.Sort = query["sort"]`, `for (const v of params.sort ?? []) q.append("sort", String(v));`,
		},
		{
			"deepObject",
			`q["filter[min]"] = []string{strconv.Itoa(*params.Filter.Min)}`, `if s := query.Get("filter[min]"); s != "" {`,
			"q.append(`filter[${k}]`, String(v));",
		},
		{
			"object, not exploded, as names and values in turn",
			`parts = append(parts, "x", strconv.Itoa(params.Point.X))`,
			`return params, errors.New("invalid point: not names and values in turn")`,
			`raw.push("point=" + join(parts, ",", esc));`,
		},
		{
			"exploded object, each member under its own name",
			`q["from"] = []string{strconv.Itoa(*params.Range.From)}`, `if s := query.Get("from"); s != "" {`,
			`for (const [k, v] of Object.entries(params.range ?? {}).filter(([, v]) => v != null)) q.append(k, String(v));`,
		},
		{
			"allowReserved",
			`escaped["path"] = []string{escapeReserved(params.Path)}`, `if s := query.Get("path"); s != "" {`,
			`if (params.path != null) raw.push(esc("path") + "=" + escR(params.path));`,
		},
		{
			"exploded map, holding every parameter no other does",
			`q[k] = []string{strconv.Itoa(v)}`,
			`case "user_ids", "tags", "words", "codes", "sort", "point", "from", "to", "path", "fixed":`,
			`for (const [k, v] of Object.entries(params.extra ?? {}).filter(([, v]) => v != null)) q.append(k, String(v));`,
		},
	} {
		for _, c := range []struct{ side, code, want string }{
			{"client", encode, tc.client}, {"server", parse, tc.server}, {"api.js", js, tc.jsOp},
		} {
			if !strings.Contains(c.code, c.want) {
				t.Errorf("%s: the %s lacks\n%s\nin\n%s", tc.name, c.side, c.want, c.code)
			}
		}
	}

	// an object not required sends nothing while no member is set; a required one always sends its required members
	if want := "if params.Filter.Status != \"\" || params.Filter.Min != nil {\n"; !strings.Contains(encode, want) {
		t.Errorf("the client does not guard the optional object filter:\n%s", encode)
	}

	if want := "if params.Point.X != 0"; strings.Contains(encode, want) {
		t.Errorf("the client guards the required member x of the required object point:\n%s", encode)
	}

	// a required member that may be null is sent like any other required member, null being its zero value
	if want := "parts = append(parts, \"z\", strconv.Itoa(params.Point.Z))"; !strings.Contains(encode, want) {
		t.Errorf("the client does not send the nullable member z:\n%s", encode)
	}

	// a server whose parameters are all split from the raw query declares no query values it would not use
	if parse := doc.op("Joined").ParseArgs(); strings.Contains(parse, "req.URL.Query()") {
		t.Errorf("joined reads query values it never uses:\n%s", parse)
	}

	if names, prefixes := search.OtherQueryNames(search.QueryParams[9]); len(names) != 10 || len(prefixes) != 1 || prefixes[0] != "filter[" {
		t.Errorf("OtherQueryNames = %v, %v", names, prefixes)
	}

	if got, want := search.ParseArgsFunc(), "parseSearchArgs"; got != want {
		t.Errorf("ParseArgsFunc = %q, want %q", got, want)
	}

	if !doc.EscapesQuery() || !doc.JoinsQuery() || !doc.AllowsReserved() {
		t.Errorf("the document needs every query helper: escapes %v, joins %v, reserved %v",
			doc.EscapesQuery(), doc.JoinsQuery(), doc.AllowsReserved())
	}

	if !strings.HasSuffix(encode, "u.RawQuery = encodeQuery(q, escaped)") {
		t.Errorf("an operation escaping values itself must merge them into the query:\n%s", encode)
	}

	// an operation of plain values leaves all escaping to url.Values and URLSearchParams
	if plain.EscapesQuery() || !strings.HasSuffix(plain.EncodeQuery(), "u.RawQuery = q.Encode()") {
		t.Errorf("plain escapes the query itself:\n%s", plain.EncodeQuery())
	}

	if got := plain.JSQueryString(); got != "q" {
		t.Errorf("plain's JSQueryString = %q, want q", got)
	}

	if got, want := search.JSQueryString(), `[q.toString(), ...raw].filter(Boolean).join("&")`; got != want {
		t.Errorf("search's JSQueryString = %q, want %q", got, want)
	}
}

// TestPathParams checks how the server reads path parameters: in place where they cannot fail to parse.
func TestPathParams(t *testing.T) {
	doc := queryDoc(t)

	item := doc.op("GetItem")
	if got := item.ParseArgsFunc(); got != "" {
		t.Errorf("a string path parameter needs no parse function, got %q", got)
	}

	if got, want := item.PathParams[0].ParsePath(), `r.PathValue("slug")`; got != want {
		t.Errorf("ParsePath = %q, want %q", got, want)
	}

	version := doc.op("GetItemVersion")
	if got, want := version.ParseArgsFunc(), "parseGetItemVersionArgs"; got != want {
		t.Fatalf("ParseArgsFunc = %q, want %q", got, want)
	}

	if got, want := version.ParseArgsResults(), "slug, n"; got != want {
		t.Errorf("ParseArgsResults = %q, want %q", got, want)
	}

	parse := version.ParseArgs()
	for _, want := range []string{
		"func parseGetItemVersionArgs(req *http.Request) (slug string, n int, err error) {",
		`slug = req.PathValue("slug")`,
		`return slug, n, fmt.Errorf("invalid n: %w", err)`,
	} {
		if !strings.Contains(parse, want) {
			t.Errorf("ParseArgs lacks %s in\n%s", want, parse)
		}
	}

	if strings.Contains(parse, "req.URL.Query()") {
		t.Errorf("ParseArgs reads a query it has no parameters in:\n%s", parse)
	}
}
