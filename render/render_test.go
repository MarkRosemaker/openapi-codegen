package render_test

import (
	_ "embed"
	"encoding/json/v2"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/MarkRosemaker/openapi-codegen/config"
	"github.com/MarkRosemaker/openapi-codegen/ir"
	"github.com/MarkRosemaker/openapi-codegen/render"
)

//go:embed testdata/pet/ir.json
var petdocJSON []byte

var genAll = config.Generate{
	Types:      true,
	Client:     true,
	ClientTest: true,
	Server:     true,
	JS:         true,
}

// failOnReadFS wraps an fs.FS, making Open fail for files matching failName.
type failOnReadFS struct {
	fs.FS
	failName string
}

func (f failOnReadFS) Open(name string) (fs.File, error) {
	if name == f.failName {
		return nil, errors.New("intentional read error")
	}
	return f.FS.Open(name)
}

// buildPetDoc constructs a minimal IR document suitable for render testing.
func buildPetDoc(t *testing.T) *ir.Document {
	t.Helper()

	doc := &ir.Document{}
	if err := json.Unmarshal(petdocJSON, doc); err != nil {
		t.Fatal(err)
	}

	return doc
}

func TestFiles_Types(t *testing.T) {
	doc := buildPetDoc(t)

	files, err := render.Files(doc, genAll)
	if err != nil {
		t.Fatal(err)
	}

	// Find the types file.
	var typesFile *render.File
	for i := range files {
		if files[i].Name == "types.gen.go" {
			typesFile = &files[i]
			break
		}
	}
	if typesFile == nil {
		t.Fatal("types.go not found in rendered output")
	}

	content := string(typesFile.Content)

	// Must contain package declaration.
	if !contains(content, "package petapi") {
		t.Error("missing package declaration")
	}

	// Struct type.
	if !contains(content, "type ListPetsParams struct") {
		t.Error("missing ListPetsParams struct")
	}
	if !contains(content, "Limit int") {
		t.Error("missing Limit field in params struct")
	}
	if !contains(content, "type Pet struct") {
		t.Error("missing Pet struct")
	}
	if !contains(content, `json:"id,omitzero"`) {
		t.Error("missing id field with omitzero tag")
	}
	if !contains(content, `json:"name"`) {
		t.Error("missing name field")
	}

	// Enum type.
	if !contains(content, "type Status string") {
		t.Error("missing Status enum type")
	}
	// gofumpt aligns const values, so check each token separately.
	if !contains(content, "StatusActive") || !contains(content, `= "active"`) {
		t.Errorf("missing StatusActive const; content:\n%s", content)
	}

	// Array alias.
	if !contains(content, "type Tags []string") {
		t.Error("missing Tags array alias")
	}
}

func TestFiles_IntegerEnum(t *testing.T) {
	// A named integer enum schema — value 1 must render as `Value = 1`, not `Value = "1"`.
	doc := &ir.Document{
		PackageName: "pkg",
		Schemas: []ir.Schema{
			{
				Name: "Priority",
				Kind: ir.SchemaKindEnum,
				Type: "int",
				EnumValues: []ir.EnumValue{
					{GoName: "PriorityLow", Value: "1", Literal: "1"},
					{GoName: "PriorityHigh", Value: "2", Literal: "2"},
				},
			},
		},
	}

	files, err := render.Files(doc, genAll)
	if err != nil {
		t.Fatalf("Files: %v", err)
	}

	var content string
	for _, f := range files {
		if f.Name == "types.gen.go" {
			content = string(f.Content)
			break
		}
	}
	if content == "" {
		t.Fatal("types.gen.go not found")
	}
	if !contains(content, "type Priority int") {
		t.Errorf("missing Priority enum type; content:\n%s", content)
	}
	if !contains(content, "PriorityLow") || !contains(content, "Priority = 1") {
		t.Errorf("expected unquoted integer literal; content:\n%s", content)
	}
	if contains(content, `"1"`) {
		t.Errorf("integer enum value must not be quoted; content:\n%s", content)
	}
}

func TestFiles_UnionSchema(t *testing.T) {
	// A oneOf union — pointer-bag struct with exactly-one-variant unmarshal.
	doc := &ir.Document{
		PackageName: "pkg",
		Schemas: []ir.Schema{
			{
				Name:    "Payment",
				Kind:    ir.SchemaKindUnion,
				IsOneOf: true,
				UnionVariants: []ir.UnionVariant{
					{FieldName: "Card", Type: "Card"},
					{FieldName: "Bank", Type: "Bank"},
				},
			},
		},
	}

	files, err := render.Files(doc, genAll)
	if err != nil {
		t.Fatalf("Files: %v", err)
	}

	var content string
	for _, f := range files {
		if f.Name == "types.gen.go" {
			content = string(f.Content)
			break
		}
	}
	if content == "" {
		t.Fatal("types.gen.go not found")
	}
	if !contains(content, "type Payment struct") {
		t.Errorf("missing Payment struct; content:\n%s", content)
	}
	if !contains(content, "Card *Card") || !contains(content, "Bank *Bank") {
		t.Errorf("missing pointer-bag fields; content:\n%s", content)
	}
	if !contains(content, "func (v *Payment) UnmarshalJSONFrom(dec *jsontext.Decoder) error") {
		t.Errorf("missing UnmarshalJSONFrom method; content:\n%s", content)
	}
	if !contains(content, "func (v *Payment) MarshalJSONTo(enc *jsontext.Encoder) error") {
		t.Errorf("missing MarshalJSONTo method; content:\n%s", content)
	}
	if !contains(content, "expected exactly one matching variant") {
		t.Errorf("oneOf union should enforce exactly one match; content:\n%s", content)
	}
}

func TestFiles_UnionSchema_AnyOf(t *testing.T) {
	// An anyOf union enforces at least one match, not exactly one.
	doc := &ir.Document{
		PackageName: "pkg",
		Schemas: []ir.Schema{
			{
				Name: "Shape",
				Kind: ir.SchemaKindUnion,
				UnionVariants: []ir.UnionVariant{
					{FieldName: "Circle", Type: "Circle"},
					{FieldName: "Rect", Type: "Rect"},
				},
			},
		},
	}

	files, err := render.Files(doc, genAll)
	if err != nil {
		t.Fatalf("Files: %v", err)
	}

	var content string
	for _, f := range files {
		if f.Name == "types.gen.go" {
			content = string(f.Content)
			break
		}
	}
	if !contains(content, "expected at least one matching variant") {
		t.Errorf("anyOf union should enforce at least one match; content:\n%s", content)
	}
	if contains(content, "expected exactly one matching variant") {
		t.Errorf("anyOf union must not enforce exactly one match; content:\n%s", content)
	}
}

func TestFilesFromFS_RenderError(t *testing.T) {
	// Template that produces invalid Go → RenderTemplate returns error → FilesFromFS propagates it.
	memFS := fstest.MapFS{
		"types.gen.go.tmpl": &fstest.MapFile{Data: []byte("not valid go @@@@")},
	}
	_, err := render.FilesFromFS(memFS, &ir.Document{PackageName: "pkg"}, genAll)
	if err == nil {
		t.Fatal("expected error from invalid Go template")
	}
}

func TestFilesFromFS_ValidTemplate(t *testing.T) {
	// Minimal valid Go template alongside a non-tmpl file and a dir; both should be skipped.
	memFS := fstest.MapFS{
		"types.gen.go.tmpl": &fstest.MapFile{Data: []byte("package {{.PackageName}}\n")},
		"README.md":         &fstest.MapFile{Data: []byte("# docs")},
		"subdir/x":          &fstest.MapFile{Data: []byte("ignored")},
	}
	files, err := render.FilesFromFS(memFS, &ir.Document{PackageName: "mypkg"}, genAll)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Name != "types.gen.go" {
		t.Errorf("Name = %q, want out.go", files[0].Name)
	}
}

func TestRenderTemplate_ParseError(t *testing.T) {
	_, err := render.RenderTemplate("bad.go.tmpl", "{{invalid", nil)
	if err == nil {
		t.Fatal("expected error for invalid template syntax")
	}
}

func TestRenderTemplate_GoimportsError(t *testing.T) {
	// Template that produces invalid Go (not parseable by goimports).
	_, err := render.RenderTemplate("bad.go.tmpl", "not valid go at all @@@@", nil)
	if err == nil {
		t.Fatal("expected error for non-Go template output")
	}
}

func TestFiles_Client(t *testing.T) {
	doc := buildPetDoc(t)

	files, err := render.Files(doc, genAll)
	if err != nil {
		t.Fatal(err)
	}

	var clientFile *render.File
	for i := range files {
		if files[i].Name == "client.gen.go" {
			clientFile = &files[i]
			break
		}
	}
	if clientFile == nil {
		t.Fatal("client.gen.go not found in rendered output")
	}

	content := string(clientFile.Content)

	if !contains(content, "package petapi") {
		t.Error("missing package declaration")
	}
	if !contains(content, `defaultBaseURL`) {
		t.Error("missing defaultBaseURL constant")
	}
	if !contains(content, `"api.example.com"`) {
		t.Error("missing base URL field")
	}
	if !contains(content, "type Client struct") {
		t.Error("missing Client struct")
	}
	if !contains(content, "func NewClient(opts ...ClientOption) (*Client, error)") {
		t.Error("missing NewClient constructor")
	}
	if !contains(content, "func (c *Client) ListPets(ctx context.Context, params *ListPetsParams) ([]Pet, error)") {
		t.Errorf("missing ListPets method signature; content:\n%s", content)
	}
	if !contains(content, "http.MethodGet") {
		t.Error("missing http.MethodGet")
	}
	if !contains(content, `"limit"`) {
		t.Error("missing query param name")
	}
}

func TestFiles_Server(t *testing.T) {
	doc := buildPetDoc(t)

	files, err := render.Files(doc, genAll)
	if err != nil {
		t.Fatal(err)
	}

	var serverFile *render.File
	for i := range files {
		if files[i].Name == "server.gen.go" {
			serverFile = &files[i]
			break
		}
	}
	if serverFile == nil {
		t.Fatal("server.gen.go not found in rendered output")
	}

	content := string(serverFile.Content)

	if !contains(content, "package petapi") {
		t.Error("missing package declaration")
	}
	if !contains(content, "type Service interface") {
		t.Error("missing Service interface")
	}
	if !contains(content, "ListPets(ctx context.Context, params *ListPetsParams)") {
		t.Errorf("missing ListPets in Service; content:\n%s", content)
	}
	if !contains(content, "func RegisterService(svc Service, mux *http.ServeMux, pathPrefix string) {") {
		t.Error("missing RegisterService function")
	}
	if !contains(content, `fmt.Sprintf("%s%s", pathPrefix, "/pets")`) {
		t.Error("missing path pattern")
	}
	if !contains(content, `fmt.Sprintf("GET %s", path)`) {
		t.Error("missing GET pattern")
	}
	if !contains(content, `q.Get("limit")`) {
		t.Error("missing query param extraction")
	}
	if !contains(content, "svc.ListPets(ctx") {
		t.Error("missing service call")
	}
}

func TestFiles_ClientTest(t *testing.T) {
	doc := buildPetDoc(t)

	files, err := render.Files(doc, genAll)
	if err != nil {
		t.Fatal(err)
	}

	var testFile *render.File
	for i := range files {
		if files[i].Name == "client.gen_test.go" {
			testFile = &files[i]
			break
		}
	}
	if testFile == nil {
		t.Fatal("client.gen_test.go not found in rendered output")
	}

	content := string(testFile.Content)

	if !contains(content, "package petapi") {
		t.Error("missing package declaration")
	}
	if !contains(content, "func newTestServer(") {
		t.Error("missing newTestServer helper")
	}
	if !contains(content, "type roundTripFunc func(*http.Request) (*http.Response, error)") || !contains(content, "RoundTrip(") {
		t.Error("missing testRoundTripper")
	}
	if !contains(content, `"ListPets"`) {
		t.Error("missing ListPets subtest in TestClient_Error")
	}
	if !contains(content, "WithBaseURL(baseURL)") {
		t.Error("missing WithBaseURL call")
	}
	if !contains(content, "c.ListPets(t.Context()") {
		t.Error("missing ListPets call")
	}
	// No interactions in the pet doc so TestClient_Interactions should not appear.
	if contains(content, "TestClient_Interactions") {
		t.Error("unexpected TestClient_Interactions (pet doc has no interactions)")
	}
}

func TestFiles_APIJS(t *testing.T) {
	doc := buildPetDoc(t)

	files, err := render.Files(doc, genAll)
	if err != nil {
		t.Fatal(err)
	}

	var apiFile *render.File
	for i := range files {
		if files[i].Name == "api.js" {
			apiFile = &files[i]
			break
		}
	}
	if apiFile == nil {
		t.Fatal("api.js not found in rendered output")
	}

	content := string(apiFile.Content)

	if !contains(content, "window.API = {") {
		t.Error("missing window.API assignment")
	}
	if !contains(content, "async function apiFetch(") {
		t.Error("missing apiFetch function")
	}
	if !contains(content, "window.ENV") {
		t.Error("missing ENV flag")
	}
	if !contains(content, "listPets:") {
		t.Errorf("missing listPets method; content:\n%s", content)
	}
	// Pet doc has a query param on listPets.
	if !contains(content, `q.set("limit"`) {
		t.Error("missing query param set call")
	}
	if !contains(content, `new URLSearchParams()`) {
		t.Error("missing URLSearchParams construction")
	}
	if !contains(content, `operationId: "ListPets"`) {
		t.Errorf("missing operationId on generated call; content:\n%s", content)
	}
	if !contains(content, "function reportApiError(") {
		t.Error("missing reportApiError function")
	}
	if !contains(content, `"petapi:ws-message"`) {
		t.Errorf("missing package-namespaced ws-message event name; content:\n%s", content)
	}
	if !contains(content, "reportApiError(operationId, statusCode, message)") {
		t.Error("missing reportApiError call in mock branch")
	}
	if !contains(content, "reportApiError(operationId, res.status, message)") {
		t.Error("missing reportApiError call in real-fetch error branch")
	}
	if !contains(content, `reportApiError(operationId, "timeout"`) {
		t.Error("missing reportApiError call on timeout")
	}
	if !contains(content, `reportApiError(operationId, "network"`) {
		t.Error("missing reportApiError call on network error")
	}
}

func TestFiles_DateTimeOrIntJSONOpts(t *testing.T) {
	// A minimal document with one schema containing an IsDateTimeOrInt field.
	doc := &ir.Document{
		PackageName:            "pkg",
		HasDateTimeOrIntFields: true,
		Schemas: []ir.Schema{
			{
				Name: "Event",
				Kind: ir.SchemaKindStruct,
				Fields: []ir.Field{
					{
						Name:            "Timestamp",
						JSONName:        "timestamp",
						Type:            "time.Time",
						JSONTag:         `json:"timestamp"`,
						IsDateTimeOrInt: true,
					},
				},
			},
		},
	}

	files, err := render.Files(doc, genAll)
	if err != nil {
		t.Fatalf("Files: %v", err)
	}

	var content string
	for _, f := range files {
		if f.Name == "types.gen.go" {
			content = string(f.Content)
			break
		}
	}
	if content == "" {
		t.Fatal("types.gen.go not found")
	}
	if !contains(content, "jsonutil.TimeUnmarshalStringOrIntUnix") {
		t.Errorf("missing TimeUnmarshalStringOrIntUnix in jsonOpts; content:\n%s", content)
	}
	if !contains(content, "json.WithUnmarshalers") {
		t.Errorf("missing WithUnmarshalers; content:\n%s", content)
	}
}

func TestFilesFromFS_ReadFileError(t *testing.T) {
	memFS := fstest.MapFS{
		"types.gen.go.tmpl": &fstest.MapFile{Data: []byte("package {{.PackageName}}\n")},
	}
	failing := failOnReadFS{FS: memFS, failName: "types.gen.go.tmpl"}
	_, err := render.FilesFromFS(failing, &ir.Document{PackageName: "pkg"}, genAll)
	if err == nil {
		t.Fatal("expected error when reading template file fails")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
