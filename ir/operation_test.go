package ir_test

import (
	"encoding/json/jsontext"
	"testing"

	"github.com/MarkRosemaker/openapi"
	"github.com/MarkRosemaker/openapi-codegen/ir"
)

func makeNamedRef(name string) *openapi.SchemaRef {
	ref := &openapi.SchemaRef{}
	ref.Value = &openapi.Schema{Type: openapi.TypeObject}
	ref.Ref = &openapi.Reference{Identifier: "#/components/schemas/" + name}
	return ref
}

func makeParam(name string, in openapi.ParameterLocation, required bool, schema *openapi.Schema) *openapi.ParameterRef {
	p := &openapi.ParameterRef{}
	p.Value = &openapi.Parameter{
		Name:     name,
		In:       in,
		Required: required,
		Schema:   schema,
	}
	return p
}

func makeResponse(desc, contentType string, schemaRef *openapi.SchemaRef) *openapi.ResponseRef {
	r := &openapi.ResponseRef{}
	r.Value = &openapi.Response{Description: desc}
	if schemaRef != nil {
		mt := &openapi.MediaType{Schema: schemaRef}
		c := openapi.Content{}
		c.Set(openapi.MediaRange(contentType), mt)
		r.Value.Content = c
	}
	return r
}

func TestFromOperation_Basic(t *testing.T) {
	op := &openapi.Operation{
		OperationID: "listPets",
		Summary:     "List all pets",
	}
	op.Responses = openapi.OperationResponses{}
	op.Responses.Set("200", makeResponse("OK", "application/json", makeNamedRef("PetList")))

	got, err := ir.FromOperation("/pets", nil, "GET", op, nil)
	if err != nil {
		t.Fatal(err)
	}

	if got.Name != "ListPets" {
		t.Errorf("Name = %q, want ListPets", got.Name)
	}
	if got.Method != "GET" {
		t.Errorf("Method = %q, want GET", got.Method)
	}
	if got.PathTemplate != "/pets" {
		t.Errorf("PathTemplate = %q, want /pets", got.PathTemplate)
	}
	if got.Summary != "List all pets" {
		t.Errorf("Summary = %q, want 'List all pets'", got.Summary)
	}
	if got.HasParams {
		t.Error("HasParams = true, want false")
	}
	if got.SuccessReturn == nil || got.SuccessReturn.Name != "PetList" {
		t.Errorf("SuccessReturn = %v, want &GoType{Name:PetList}", got.SuccessReturn)
	}
	if len(got.Responses) != 1 {
		t.Fatalf("Responses len = %d, want 1", len(got.Responses))
	}
	if got.Responses[0].GoConstant != "http.StatusOK" {
		t.Errorf("GoConstant = %q, want http.StatusOK", got.Responses[0].GoConstant)
	}
}

func TestFromOperation_RawBytesSuccess(t *testing.T) {
	// A success response with no JSON media type (e.g. text/plain) is
	// exposed as a raw []byte read from the body, not a JSON-decoded type.
	op := &openapi.Operation{OperationID: "getLlmsTxt"}
	op.Responses = openapi.OperationResponses{}
	op.Responses.Set("200", makeResponse("OK", "text/plain", makeNamedRef("Unused")))

	got, err := ir.FromOperation("/llms.txt", nil, "GET", op, nil)
	if err != nil {
		t.Fatal(err)
	}

	if got.SuccessReturn == nil || got.SuccessReturn.String() != "[]byte" {
		t.Fatalf("SuccessReturn = %v, want []byte", got.SuccessReturn)
	}
	if !got.RawBytesSuccess {
		t.Error("RawBytesSuccess = false, want true")
	}
	if len(got.Responses) != 1 {
		t.Fatalf("Responses len = %d, want 1", len(got.Responses))
	}
	if !got.Responses[0].IsRawBytes {
		t.Error("Responses[0].IsRawBytes = false, want true")
	}
	if got.Responses[0].ContentType != "text/plain" {
		t.Errorf("Responses[0].ContentType = %q, want text/plain", got.Responses[0].ContentType)
	}
}

func TestFromOperation_RawBytesErrorOnly(t *testing.T) {
	// A non-success response with no JSON media type is raw bytes too, but
	// doesn't make the operation's SuccessReturn itself raw bytes.
	op := &openapi.Operation{OperationID: "getThing"}
	op.Responses = openapi.OperationResponses{}
	op.Responses.Set("200", makeResponse("OK", "application/json", makeNamedRef("Thing")))
	op.Responses.Set("403", makeResponse("Forbidden", "text/html", makeNamedRef("Unused")))

	got, err := ir.FromOperation("/thing", nil, "GET", op, nil)
	if err != nil {
		t.Fatal(err)
	}

	if got.SuccessReturn == nil || got.SuccessReturn.Name != "Thing" {
		t.Fatalf("SuccessReturn = %v, want &GoType{Name:Thing}", got.SuccessReturn)
	}
	if got.RawBytesSuccess {
		t.Error("RawBytesSuccess = true, want false (success response is JSON)")
	}
	if len(got.Responses) != 2 {
		t.Fatalf("Responses len = %d, want 2", len(got.Responses))
	}

	errResp := got.Responses[1]
	if !errResp.IsRawBytes {
		t.Error("Responses[1].IsRawBytes = false, want true")
	}
	if errResp.GoType == nil || errResp.GoType.String() != "[]byte" {
		t.Errorf("Responses[1].GoType = %v, want []byte", errResp.GoType)
	}
	if errResp.ContentType != "text/html" {
		t.Errorf("Responses[1].ContentType = %q, want text/html", errResp.ContentType)
	}
}

func TestFromOperation_JSONPreferredOverRawBytes(t *testing.T) {
	// When a response declares both a JSON and a non-JSON media type, JSON wins.
	op := &openapi.Operation{OperationID: "getThing"}
	op.Responses = openapi.OperationResponses{}
	r := &openapi.Response{Description: "OK"}
	r.Content = openapi.Content{}
	r.Content.Set("text/plain", &openapi.MediaType{})
	r.Content.Set("application/json", &openapi.MediaType{Schema: makeNamedRef("Thing")})
	rRef := &openapi.ResponseRef{Value: r}
	op.Responses.Set("200", rRef)

	got, err := ir.FromOperation("/thing", nil, "GET", op, nil)
	if err != nil {
		t.Fatal(err)
	}

	if got.RawBytesSuccess {
		t.Error("RawBytesSuccess = true, want false (JSON media type present)")
	}
	if got.Responses[0].IsRawBytes {
		t.Error("Responses[0].IsRawBytes = true, want false")
	}
	if got.Responses[0].ContentType != "application/json" {
		t.Errorf("Responses[0].ContentType = %q, want application/json", got.Responses[0].ContentType)
	}
}

func TestFromOperation_PathParam(t *testing.T) {
	params := openapi.ParameterList{
		makeParam("petId", openapi.ParameterLocationPath, true, &openapi.Schema{Type: openapi.TypeInteger}),
	}

	op := &openapi.Operation{
		OperationID: "getPetById",
		Parameters:  params,
	}
	op.Responses = openapi.OperationResponses{}
	op.Responses.Set("200", makeResponse("OK", "application/json", makeNamedRef("Pet")))

	got, err := ir.FromOperation("/pets/{petId}", nil, "GET", op, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(got.PathParams) != 1 {
		t.Fatalf("PathParams len = %d, want 1", len(got.PathParams))
	}
	p := got.PathParams[0]
	if p.GoName != "petID" {
		t.Errorf("GoName = %q, want petID", p.GoName)
	}
	if p.Type != "int" {
		t.Errorf("Type = %q, want int", p.Type)
	}
	if got, want := p.FormatExpr(), "strconv.Itoa(petID)"; got != want {
		t.Errorf("FormatExpr = %q, want %q", got, want)
	}
	if !p.Required {
		t.Error("Required = false, want true")
	}

	// JoinPathArgs: ["pets", strconv.Itoa(petID)]
	if len(got.JoinPathArgs) != 2 {
		t.Fatalf("JoinPathArgs len = %d, want 2", len(got.JoinPathArgs))
	}
	if got.JoinPathArgs[0] != `"pets"` {
		t.Errorf("JoinPathArgs[0] = %q, want \"pets\"", got.JoinPathArgs[0])
	}
	if got.JoinPathArgs[1] != "strconv.Itoa(petID)" {
		t.Errorf("JoinPathArgs[1] = %q, want strconv.Itoa(petID)", got.JoinPathArgs[1])
	}
}

func TestFromOperation_QueryParams(t *testing.T) {
	params := openapi.ParameterList{
		makeParam("limit", openapi.ParameterLocationQuery, false, &openapi.Schema{Type: openapi.TypeInteger}),
		makeParam("status", openapi.ParameterLocationQuery, false, &openapi.Schema{
			Type: openapi.TypeString,
			Enum: []jsontext.Value{jsontext.Value(`"active"`), jsontext.Value(`"inactive"`)},
		}),
	}

	op := &openapi.Operation{
		OperationID: "listPets",
		Parameters:  params,
	}
	op.Responses = openapi.OperationResponses{}
	op.Responses.Set("200", makeResponse("OK", "application/json", makeNamedRef("PetList")))

	got, err := ir.FromOperation("/pets", nil, "GET", op, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(got.QueryParams) != 2 {
		t.Fatalf("QueryParams len = %d, want 2", len(got.QueryParams))
	}
	if got.ParamStructName != "ListPetsParams" {
		t.Errorf("ParamStructName = %q, want ListPetsParams", got.ParamStructName)
	}
	if !got.HasParams {
		t.Error("HasParams = false, want true")
	}

	limit := got.QueryParams[0]
	if got, want := limit.NotZero(), "params.Limit != 0"; got != want {
		t.Fatalf("NotZero = %q, want %q", got, want)
	}

	status := got.QueryParams[1]
	if !status.IsEnum {
		t.Error("IsEnum = false for enum param, want true")
	}
}

func TestFromOperation_MissingOperationID(t *testing.T) {
	op := &openapi.Operation{}
	_, err := ir.FromOperation("/pets", nil, "GET", op, nil)
	if err == nil {
		t.Fatal("expected error for missing operationId")
	}
}

func TestFromOperation_PathItemParamsMerge(t *testing.T) {
	// Path-item param "limit", operation overrides with its own "limit"
	pathItemParams := openapi.ParameterList{
		makeParam("limit", openapi.ParameterLocationQuery, false, &openapi.Schema{Type: openapi.TypeInteger}),
		makeParam("offset", openapi.ParameterLocationQuery, false, &openapi.Schema{Type: openapi.TypeInteger}),
	}
	opParams := openapi.ParameterList{
		makeParam("limit", openapi.ParameterLocationQuery, true, &openapi.Schema{Type: openapi.TypeInteger}), // override
	}

	op := &openapi.Operation{
		OperationID: "listPets",
		Parameters:  opParams,
	}
	op.Responses = openapi.OperationResponses{}
	op.Responses.Set("200", makeResponse("OK", "application/json", nil))

	got, err := ir.FromOperation("/pets", pathItemParams, "GET", op, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Should have: offset (from path-item) + limit (from operation, overriding path-item)
	if len(got.QueryParams) != 2 {
		t.Fatalf("QueryParams len = %d, want 2 (offset + limit)", len(got.QueryParams))
	}
	// offset comes first (from path-item), then limit (from operation)
	if got.QueryParams[0].JSONName != "offset" {
		t.Errorf("QueryParams[0].JSONName = %q, want offset", got.QueryParams[0].JSONName)
	}
	if got.QueryParams[1].JSONName != "limit" {
		t.Errorf("QueryParams[1].JSONName = %q, want limit", got.QueryParams[1].JSONName)
	}
	// The operation's limit should be required=true (override wins)
	if !got.QueryParams[1].Required {
		t.Error("overridden limit.Required = false, want true")
	}
}

func TestFromOperation_RequestBody(t *testing.T) {
	bodySchema := makeNamedRef("NewPet")
	mt := &openapi.MediaType{Schema: bodySchema}
	c := openapi.Content{}
	c.Set("application/json", mt)

	rb := &openapi.RequestBodyRef{}
	rb.Value = &openapi.RequestBody{Content: c, Required: true}

	op := &openapi.Operation{
		OperationID: "createPet",
		RequestBody: rb,
	}
	op.Responses = openapi.OperationResponses{}
	op.Responses.Set("201", makeResponse("Created", "", nil))

	got, err := ir.FromOperation("/pets", nil, "POST", op, nil)
	if err != nil {
		t.Fatal(err)
	}

	if got.RequestBody == nil {
		t.Fatal("RequestBody = nil, want &ReqBody")
	}
	if got.RequestBody.TypeName != "NewPet" {
		t.Errorf("TypeName = %q, want NewPet", got.RequestBody.TypeName)
	}
	if !got.RequestBody.Required {
		t.Error("Required = false, want true")
	}

	if len(got.Responses) != 1 || got.Responses[0].GoConstant != "http.StatusCreated" {
		t.Errorf("Responses[0].GoConstant = %q, want http.StatusCreated", got.Responses[0].GoConstant)
	}
}

func TestStatusCodeToConst(t *testing.T) {
	tests := []struct {
		op   *openapi.Operation
		code openapi.StatusCode
		want string
	}{
		{want: "http.StatusOK"},
		{want: "http.StatusCreated"},
		{want: "http.StatusNoContent"},
		{want: "http.StatusBadRequest"},
		{want: "http.StatusUnauthorized"},
		{want: "http.StatusForbidden"},
		{want: "http.StatusNotFound"},
		{want: "http.StatusTooManyRequests"},
		{want: "http.StatusInternalServerError"},
	}

	codes := []openapi.StatusCode{"200", "201", "204", "400", "401", "403", "404", "429", "500"}
	for i, tc := range tests {
		op := &openapi.Operation{OperationID: "testOp"}
		op.Responses = openapi.OperationResponses{}
		op.Responses.Set(codes[i], makeResponse("desc", "", nil))

		got, err := ir.FromOperation("/test", nil, "GET", op, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Responses) == 0 {
			t.Fatalf("no responses for code %s", codes[i])
		}
		if got.Responses[0].GoConstant != tc.want {
			t.Errorf("code %s: GoConstant = %q, want %q", codes[i], got.Responses[0].GoConstant, tc.want)
		}
	}
}

func TestFromOperation_ParamMissingSchema(t *testing.T) {
	p := &openapi.ParameterRef{}
	p.Value = &openapi.Parameter{Name: "id", In: openapi.ParameterLocationPath, Required: true, Schema: nil}
	op := &openapi.Operation{
		OperationID: "getItem",
		Parameters:  openapi.ParameterList{p},
	}
	op.Responses = openapi.OperationResponses{}
	op.Responses.Set("200", makeResponse("OK", "", nil))
	_, err := ir.FromOperation("/items/{id}", nil, "GET", op, nil)
	if err == nil {
		t.Fatal("expected error for missing param schema")
	}
}

func TestFromOperation_RequestBodyNoSchema(t *testing.T) {
	// request body content has a media type but no schema
	mt := &openapi.MediaType{Schema: nil}
	c := openapi.Content{}
	c.Set("application/json", mt)
	rb := &openapi.RequestBodyRef{}
	rb.Value = &openapi.RequestBody{Content: c}

	op := &openapi.Operation{OperationID: "createItem", RequestBody: rb}
	op.Responses = openapi.OperationResponses{}
	op.Responses.Set("201", makeResponse("Created", "", nil))

	got, err := ir.FromOperation("/items", nil, "POST", op, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.RequestBody != nil {
		t.Errorf("RequestBody = %v, want nil (no schema)", got.RequestBody)
	}
}

func TestFromOperation_RangeStatusCode(t *testing.T) {
	op := &openapi.Operation{OperationID: "testOp"}
	op.Responses = openapi.OperationResponses{}
	op.Responses.Set("2XX", makeResponse("OK", "", nil))

	got, err := ir.FromOperation("/test", nil, "GET", op, nil)
	if err != nil {
		t.Fatal(err)
	}
	// "2XX" is not a plain integer; statusCodeToConst falls back to string
	if len(got.Responses) == 0 {
		t.Fatal("no responses")
	}
	if got.Responses[0].StatusCode != "2XX" {
		t.Errorf("StatusCode = %q, want 2XX", got.Responses[0].StatusCode)
	}
}

func TestJSPathTemplate(t *testing.T) {
	tests := []struct {
		pathTemplate string
		pathParams   []ir.Param
		want         string
	}{
		{"/pets", nil, "/pets"},
		{"/pets/{petId}", []ir.Param{{JSONName: "petId", GoName: "petID"}}, "/pets/${petID}"},
		{"/accounts/{account_id}/positions", []ir.Param{{JSONName: "account_id", GoName: "accountID"}}, "/accounts/${accountID}/positions"},
		{"/api/xbrl/companyfacts/CIK{cik}.json", []ir.Param{{JSONName: "cik", GoName: "cik"}}, "/api/xbrl/companyfacts/CIK${cik}.json"},
	}
	for _, tc := range tests {
		op := ir.Operation{PathTemplate: tc.pathTemplate, PathParams: tc.pathParams}
		if got := op.JSPathTemplate(); got != tc.want {
			t.Errorf("JSPathTemplate(%q) = %q, want %q", tc.pathTemplate, got, tc.want)
		}
	}
}

func TestFromOperation_Deprecated(t *testing.T) {
	op := &openapi.Operation{OperationID: "oldOp", Deprecated: true}
	op.Responses = openapi.OperationResponses{}
	op.Responses.Set("200", makeResponse("OK", "", nil))

	got, err := ir.FromOperation("/old", nil, "GET", op, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Deprecated {
		t.Error("Deprecated = false, want true")
	}
}

func TestFormatAndNotZeroExprs(t *testing.T) {
	type test struct {
		typ         openapi.DataType
		format      openapi.Format
		wantFmt     string
		wantNotZero string
	}
	tests := []test{
		{openapi.TypeString, "", `params.Name`, `params.Name != ""`},
		{openapi.TypeString, openapi.FormatEmail, `string(params.Name)`, `params.Name != ""`},
		{openapi.TypeString, openapi.FormatUUID, `params.Name.String()`, `params.Name != uuid.Nil`},
		{openapi.TypeString, openapi.FormatURI, `params.Name.String()`, `params.Name.Host != ""`},
		{openapi.TypeString, openapi.FormatDateTime, `params.Name.Format(time.RFC3339)`, `!params.Name.IsZero()`},
		{openapi.TypeString, openapi.FormatDate, `params.Name.String()`, `params.Name != (civil.Date{})`},
		{openapi.TypeString, openapi.FormatIPv4, `params.Name.String()`, `params.Name != nil`},
		{openapi.TypeBoolean, "", `strconv.FormatBool(params.Name)`, `params.Name`},
		{openapi.TypeInteger, "", `strconv.Itoa(params.Name)`, `params.Name != 0`},
		{openapi.TypeInteger, openapi.FormatInt32, `strconv.FormatInt(int64(params.Name), 10)`, `params.Name != 0`},
		{openapi.TypeInteger, openapi.FormatInt64, `strconv.FormatInt(params.Name, 10)`, `params.Name != 0`},
		{openapi.TypeInteger, openapi.FormatUint, `strconv.FormatUint(uint64(params.Name), 10)`, `params.Name != 0`},
		{openapi.TypeInteger, openapi.FormatUint32, `strconv.FormatUint(uint64(params.Name), 10)`, `params.Name != 0`},
		{openapi.TypeInteger, openapi.FormatUint64, `strconv.FormatUint(params.Name, 10)`, `params.Name != 0`},
		{openapi.TypeInteger, openapi.FormatDuration, `strconv.FormatInt(int64(params.Name/time.Second), 10)`, `params.Name != 0`},
		{openapi.TypeNumber, openapi.FormatFloat, `strconv.FormatFloat(float64(params.Name), 'f', -1, 32)`, `params.Name != 0`},
		{openapi.TypeNumber, "", `strconv.FormatFloat(params.Name, 'f', -1, 64)`, `params.Name != 0`},
	}

	for _, tc := range tests {
		schema := &openapi.Schema{Type: tc.typ, Format: tc.format}
		params := openapi.ParameterList{makeParam("name", openapi.ParameterLocationQuery, false, schema)}
		op := &openapi.Operation{OperationID: "testOp", Parameters: params}
		op.Responses = openapi.OperationResponses{}
		op.Responses.Set("200", makeResponse("OK", "", nil))

		got, err := ir.FromOperation("/test", nil, "GET", op, nil)
		if err != nil {
			t.Fatalf("type=%s format=%s: %v", tc.typ, tc.format, err)
		}
		if len(got.QueryParams) == 0 {
			t.Fatal("no query params")
		}
		p := got.QueryParams[0]
		if got := p.FormatExpr(); got != tc.wantFmt {
			t.Errorf("type=%s format=%s: FormatExpr = %q, want %q",
				tc.typ, tc.format, got, tc.wantFmt)
		}
		if got := p.NotZero(); got != tc.wantNotZero {
			t.Fatalf("type=%s format=%s: ZeroCheck = %q, want %q",
				tc.typ, tc.format, got, tc.wantNotZero)
		}

	}
}
