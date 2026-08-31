package ir_test

import (
	"encoding/json/jsontext"
	"testing"

	"github.com/MarkRosemaker/openapi"
	"github.com/MarkRosemaker/openapi-codegen/ir"
)

func TestGoTypeZeroValue(t *testing.T) {
	tests := []struct {
		t    ir.GoType
		want string
	}{
		{ir.GoType{Name: "Pet", IsPointer: true}, "nil"},
		{ir.GoType{Name: "Pet", IsSlice: true}, "nil"},
		{ir.GoType{Name: "string"}, `""`},
		{ir.GoType{Name: "bool"}, "false"},
		{ir.GoType{Name: "int"}, "0"},
		{ir.GoType{Name: "int32"}, "0"},
		{ir.GoType{Name: "int64"}, "0"},
		{ir.GoType{Name: "uint"}, "0"},
		{ir.GoType{Name: "float32"}, "0"},
		{ir.GoType{Name: "float64"}, "0"},
		{ir.GoType{Name: "Pet"}, "Pet{}"},
		{ir.GoType{Name: "uuid.UUID"}, "uuid.Nil"},
	}
	for _, tc := range tests {
		if got := tc.t.ZeroValue(); got != tc.want {
			t.Errorf("GoType{%q, ptr=%v, slice=%v}.ZeroValue() = %q, want %q",
				tc.t.Name, tc.t.IsPointer, tc.t.IsSlice, got, tc.want)
		}
	}
}

func TestSchemaGoType(t *testing.T) {
	tests := []struct {
		name    string
		schema  openapi.Schema
		want    string
		wantErr bool
	}{
		// boolean
		{name: "bool", schema: openapi.Schema{Type: openapi.TypeBoolean}, want: "bool"},
		// integer
		{name: "int", schema: openapi.Schema{Type: openapi.TypeInteger}, want: "int"},
		{name: "int32", schema: openapi.Schema{Type: openapi.TypeInteger, Format: openapi.FormatInt32}, want: "int32"},
		{name: "int64", schema: openapi.Schema{Type: openapi.TypeInteger, Format: openapi.FormatInt64}, want: "int64"},
		{name: "uint", schema: openapi.Schema{Type: openapi.TypeInteger, Format: openapi.FormatUint}, want: "uint"},
		{name: "uint32", schema: openapi.Schema{Type: openapi.TypeInteger, Format: openapi.FormatUint32}, want: "uint32"},
		{name: "uint64", schema: openapi.Schema{Type: openapi.TypeInteger, Format: openapi.FormatUint64}, want: "uint64"},
		{name: "duration", schema: openapi.Schema{Type: openapi.TypeInteger, Format: openapi.FormatDuration}, want: "time.Duration"},
		// number
		{name: "float64", schema: openapi.Schema{Type: openapi.TypeNumber}, want: "float64"},
		{name: "float64-double", schema: openapi.Schema{Type: openapi.TypeNumber, Format: openapi.FormatDouble}, want: "float64"},
		{name: "float32", schema: openapi.Schema{Type: openapi.TypeNumber, Format: openapi.FormatFloat}, want: "float32"},
		// string
		{name: "string", schema: openapi.Schema{Type: openapi.TypeString}, want: "string"},
		{name: "string-password", schema: openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatPassword}, want: "string"},
		{name: "string-byte", schema: openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatByte}, want: "string"},
		{name: "string-binary", schema: openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatBinary}, want: "string"},
		{name: "string-zipcode", schema: openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatZipCode}, want: "string"},
		{name: "uuid", schema: openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatUUID}, want: "uuid.UUID"},
		{name: "uri", schema: openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatURI}, want: "url.URL"},
		{name: "uriref", schema: openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatURIRef}, want: "url.URL"},
		{name: "email", schema: openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatEmail}, want: "types.Email"},
		{name: "datetime", schema: openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatDateTime}, want: "time.Time"},
		{name: "date", schema: openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatDate}, want: "civil.Date"},
		{name: "ipv4", schema: openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatIPv4}, want: "net.IP"},
		{name: "ipv6", schema: openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatIPv6}, want: "net.IP"},
		// errors
		{name: "unknown-type", schema: openapi.Schema{Type: "unknown"}, wantErr: true},
		{name: "unknown-integer-format", schema: openapi.Schema{Type: openapi.TypeInteger, Format: "base36"}, wantErr: true},
		{name: "unknown-number-format", schema: openapi.Schema{Type: openapi.TypeNumber, Format: "complex128"}, wantErr: true},
		{name: "unknown-string-format", schema: openapi.Schema{Type: openapi.TypeString, Format: "base58"}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ir.SchemaGoType(&tc.schema)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("SchemaGoType() expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("SchemaGoType() unexpected error: %v", err)
			}
			if got.String() != tc.want {
				t.Errorf("SchemaGoType() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSchemaGoType_Array(t *testing.T) {
	items := &openapi.SchemaRef{}
	items.Value = &openapi.Schema{Type: openapi.TypeString}

	s := &openapi.Schema{Type: openapi.TypeArray, Items: items}
	got, err := ir.SchemaGoType(s)
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "[]string" {
		t.Errorf("got %q, want []string", got)
	}
}

func TestSchemaGoType_ArrayRef(t *testing.T) {
	items := &openapi.SchemaRef{}
	items.Value = &openapi.Schema{Type: openapi.TypeObject}
	items.Ref = &openapi.Reference{Identifier: "#/components/schemas/MyObject"}

	s := &openapi.Schema{Type: openapi.TypeArray, Items: items}
	got, err := ir.SchemaGoType(s)
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "[]MyObject" {
		t.Errorf("got %q, want []MyObject", got)
	}
}

func TestSchemaGoType_MapAdditional(t *testing.T) {
	addl := &openapi.SchemaRef{}
	addl.Value = &openapi.Schema{Type: openapi.TypeString}

	s := &openapi.Schema{Type: openapi.TypeObject, AdditionalProperties: addl}
	got, err := ir.SchemaGoType(s)
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "map[string]string" {
		t.Errorf("got %q, want map[string]string", got)
	}
}

func TestFromComponentSchemas_Struct(t *testing.T) {
	props := openapi.SchemaRefs{}
	idRef := &openapi.SchemaRef{}
	idRef.Value = &openapi.Schema{Type: openapi.TypeInteger}
	props.Set("id", idRef)
	nameRef := &openapi.SchemaRef{}
	nameRef.Value = &openapi.Schema{Type: openapi.TypeString}
	props.Set("name", nameRef)

	schemas := openapi.Schemas{}
	schemas.Set("Pet", &openapi.Schema{
		Type:        openapi.TypeObject,
		Properties:  props,
		Required:    []string{"id"},
		Description: "A pet",
	})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(got))
	}

	s := got[0]
	if s.Name != "Pet" {
		t.Errorf("Name = %q, want Pet", s.Name)
	}
	if s.Kind != ir.SchemaKindStruct {
		t.Errorf("Kind = %v, want SchemaKindStruct", s.Kind)
	}
	if s.Description != "A pet" {
		t.Errorf("Description = %q, want 'A pet'", s.Description)
	}
	if len(s.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(s.Fields))
	}

	// id field (required) — strcase.ToGoPascal expands "id" to the initialism "ID"
	id := s.Fields[0]
	if id.Name != "ID" {
		t.Errorf("Fields[0].Name = %q, want ID", id.Name)
	}
	if id.JSONName != "id" {
		t.Errorf("Fields[0].JSONName = %q, want id", id.JSONName)
	}
	if id.Type != "int" {
		t.Errorf("Fields[0].Type = %q, want int", id.Type)
	}
	if !id.Required {
		t.Error("Fields[0].Required = false, want true")
	}
	if id.JSONTag != `json:"id"` {
		t.Errorf("Fields[0].JSONTag = %q, want json:\"id\"", id.JSONTag)
	}

	// name field (optional string)
	name := s.Fields[1]
	if name.Name != "Name" {
		t.Errorf("Fields[1].Name = %q, want Name", name.Name)
	}
	if name.JSONTag != `json:"name,omitzero"` {
		t.Errorf("Fields[1].JSONTag = %q, want json:\"name\"", name.JSONTag)
	}
}

func TestFromComponentSchemas_Enum(t *testing.T) {
	schemas := openapi.Schemas{}
	schemas.Set("Status", &openapi.Schema{
		Type: openapi.TypeString,
		Enum: []jsontext.Value{jsontext.Value(`"active"`), jsontext.Value(`"inactive"`), jsontext.Value(`"4xx"`)},
	})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(got))
	}

	s := got[0]
	if s.Kind != ir.SchemaKindEnum {
		t.Errorf("Kind = %v, want SchemaKindEnum", s.Kind)
	}
	if s.Type != "string" {
		t.Errorf("EnumType = %q, want string", s.Type)
	}
	if len(s.EnumValues) != 3 {
		t.Fatalf("expected 3 enum values, got %d", len(s.EnumValues))
	}
	if s.EnumValues[0].GoName != "StatusActive" {
		t.Errorf("EnumValues[0].GoName = %q, want StatusActive", s.EnumValues[0].GoName)
	}
	if s.EnumValues[0].Value != "active" {
		t.Errorf("EnumValues[0].Value = %q, want active", s.EnumValues[0].Value)
	}
	if s.EnumValues[0].Literal != `"active"` {
		t.Errorf("EnumValues[0].Literal = %q, want %q", s.EnumValues[0].Literal, `"active"`)
	}
	// "4xx" starts with digit → "FourXx"
	if s.EnumValues[2].GoName != "StatusFourXx" {
		t.Errorf("EnumValues[2].GoName = %q, want StatusFourXx", s.EnumValues[2].GoName)
	}
}

func TestFromComponentSchemas_EnumInteger(t *testing.T) {
	schemas := openapi.Schemas{}
	schemas.Set("Priority", &openapi.Schema{
		Type: openapi.TypeInteger,
		Enum: []jsontext.Value{jsontext.Value(`1`), jsontext.Value(`2`), jsontext.Value(`3`)},
	})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(got))
	}

	s := got[0]
	if s.Kind != ir.SchemaKindEnum {
		t.Errorf("Kind = %v, want SchemaKindEnum", s.Kind)
	}
	if s.Type != "int" {
		t.Errorf("EnumType = %q, want int", s.Type)
	}
	if len(s.EnumValues) != 3 {
		t.Fatalf("expected 3 enum values, got %d", len(s.EnumValues))
	}
	if s.EnumValues[0].Value != "1" {
		t.Errorf("EnumValues[0].Value = %q, want 1", s.EnumValues[0].Value)
	}
	if s.EnumValues[0].Literal != "1" {
		t.Errorf("EnumValues[0].Literal = %q, want 1", s.EnumValues[0].Literal)
	}
	if s.EnumValues[0].GoName != "PriorityOne" {
		t.Errorf("EnumValues[0].GoName = %q, want PriorityOne", s.EnumValues[0].GoName)
	}
}

func TestFromComponentSchemas_EnumBoolean(t *testing.T) {
	schemas := openapi.Schemas{}
	schemas.Set("Flag", &openapi.Schema{
		Type: openapi.TypeBoolean,
		Enum: []jsontext.Value{jsontext.Value(`true`), jsontext.Value(`false`)},
	})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(got))
	}

	s := got[0]
	if s.Type != "bool" {
		t.Errorf("EnumType = %q, want bool", s.Type)
	}
	if len(s.EnumValues) != 2 {
		t.Fatalf("expected 2 enum values, got %d", len(s.EnumValues))
	}
	if s.EnumValues[0].Literal != "true" {
		t.Errorf("EnumValues[0].Literal = %q, want true", s.EnumValues[0].Literal)
	}
	if s.EnumValues[1].Literal != "false" {
		t.Errorf("EnumValues[1].Literal = %q, want false", s.EnumValues[1].Literal)
	}
}

func TestFromComponentSchemas_ArrayAlias(t *testing.T) {
	items := &openapi.SchemaRef{}
	items.Value = &openapi.Schema{Type: openapi.TypeString}

	schemas := openapi.Schemas{}
	schemas.Set("Tags", &openapi.Schema{
		Type:  openapi.TypeArray,
		Items: items,
	})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(got))
	}

	s := got[0]
	if s.Kind != ir.SchemaKindAlias {
		t.Errorf("Kind = %v, want SchemaKindAlias", s.Kind)
	}
	if s.Type != "[]string" {
		t.Errorf("Type = %q, want []string", s.Type)
	}
}

func TestFromComponentSchemas_Scalars(t *testing.T) {
	// A named scalar component is declared like any other: a $ref resolves to
	// its name, and a response body decoded into it can carry an Error method,
	// which a bare string or int cannot.
	schemas := openapi.Schemas{}
	schemas.Set("Count", &openapi.Schema{Type: openapi.TypeInteger})
	schemas.Set("Name", &openapi.Schema{Type: openapi.TypeString})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 schemas, got %d", len(got))
	}

	for i, want := range []struct{ name, goType string }{
		{"Count", "int"},
		{"Name", "string"},
	} {
		if got[i].Name != want.name {
			t.Errorf("[%d].Name = %q, want %q", i, got[i].Name, want.name)
		}
		if got[i].Kind != ir.SchemaKindAlias {
			t.Errorf("[%d].Kind = %v, want %v", i, got[i].Kind, ir.SchemaKindAlias)
		}
		if got[i].Type != want.goType {
			t.Errorf("[%d].Type = %q, want %q", i, got[i].Type, want.goType)
		}
	}
}

func TestFromComponentSchemas_AllOf(t *testing.T) {
	// Base schema: type object with one property
	baseRef := &openapi.SchemaRef{}
	baseRef.Ref = &openapi.Reference{Identifier: "#/components/schemas/Base"}

	nameRef := &openapi.SchemaRef{}
	nameRef.Value = &openapi.Schema{Type: openapi.TypeString}
	inlineProps := openapi.SchemaRefs{}
	inlineProps.Set("name", nameRef)
	inlineEntry := &openapi.SchemaRef{}
	inlineEntry.Value = &openapi.Schema{
		Type:       openapi.TypeObject,
		Properties: inlineProps,
		Required:   []string{"name"},
	}

	schemas := openapi.Schemas{}
	schemas.Set("Child", &openapi.Schema{
		AllOf: openapi.SchemaRefList{baseRef, inlineEntry},
	})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(got))
	}

	s := got[0]
	if s.Kind != ir.SchemaKindAllOf {
		t.Errorf("Kind = %v, want SchemaKindAllOf", s.Kind)
	}
	if len(s.Fields) != 2 {
		t.Fatalf("expected 2 fields (1 embedded + 1 regular), got %d", len(s.Fields))
	}

	embed := s.Fields[0]
	if !embed.Embedded {
		t.Error("first field: Embedded = false, want true")
	}
	if embed.Type != "Base" {
		t.Errorf("first field Type = %q, want %q", embed.Type, "Base")
	}

	regular := s.Fields[1]
	if regular.Embedded {
		t.Error("second field: Embedded = true, want false")
	}
	if regular.Name != "Name" {
		t.Errorf("second field Name = %q, want %q", regular.Name, "Name")
	}
	if regular.Type != "string" {
		t.Errorf("second field Type = %q, want %q", regular.Type, "string")
	}
}

func TestFromComponentSchemas_MapObject(t *testing.T) {
	// A named map component is declared whatever its value type. A property
	// referencing it resolves to the component's name -- SchemaRefGoType takes
	// the name straight from the $ref -- so skipping the declaration would
	// leave that name undefined.
	valRef := &openapi.SchemaRef{}
	valRef.Value = &openapi.Schema{Type: openapi.TypeString}

	schemas := openapi.Schemas{}
	schemas.Set("StringMap", &openapi.Schema{
		Type:                 openapi.TypeObject,
		AdditionalProperties: valRef,
	})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(got))
	}

	s := got[0]
	if s.Name != "StringMap" {
		t.Errorf("Name = %q, want StringMap", s.Name)
	}
	if s.Kind != ir.SchemaKindMap {
		t.Errorf("Kind = %v, want %v", s.Kind, ir.SchemaKindMap)
	}
	if s.MapKey != "string" || s.MapValue != "string" {
		t.Errorf("map[%s]%s, want map[string]string", s.MapKey, s.MapValue)
	}
}

func TestFromComponentSchemas_AllOf_OuterRequired(t *testing.T) {
	// Required field on the outer allOf schema must be respected.
	nameRef := &openapi.SchemaRef{}
	nameRef.Value = &openapi.Schema{Type: openapi.TypeString}
	inlineProps := openapi.SchemaRefs{}
	inlineProps.Set("name", nameRef)
	inlineEntry := &openapi.SchemaRef{}
	inlineEntry.Value = &openapi.Schema{
		Type:       openapi.TypeObject,
		Properties: inlineProps,
	}

	schemas := openapi.Schemas{}
	schemas.Set("Child", &openapi.Schema{
		AllOf:    openapi.SchemaRefList{inlineEntry},
		Required: []string{"name"},
	})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0].Fields) != 1 {
		t.Fatalf("expected 1 schema with 1 field, got %v", got)
	}
	if !got[0].Fields[0].Required {
		t.Error("expected field to be required (from outer schema Required list)")
	}
}

func TestFromComponentSchemas_EmptyTypeNoAllOf(t *testing.T) {
	// A schema with no type and no allOf should be skipped.
	schemas := openapi.Schemas{}
	schemas.Set("Unknown", &openapi.Schema{})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 schemas, got %d", len(got))
	}
}

func TestFromComponentSchemas_AllOf_NilValueEntry(t *testing.T) {
	// An allOf entry with Ref==nil and Value==nil should be silently skipped.
	nilEntry := &openapi.SchemaRef{} // Ref=nil, Value=nil

	schemas := openapi.Schemas{}
	schemas.Set("Child", &openapi.Schema{
		AllOf: openapi.SchemaRefList{nilEntry},
	})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(got))
	}
	if got[0].Kind != ir.SchemaKindAllOf {
		t.Errorf("Kind = %v, want SchemaKindAllOf", got[0].Kind)
	}
	if len(got[0].Fields) != 0 {
		t.Errorf("expected 0 fields, got %d", len(got[0].Fields))
	}
}

func TestGoTypeString(t *testing.T) {
	tests := []struct {
		gt   ir.GoType
		want string
	}{
		{ir.GoType{Name: "string"}, "string"},
		{ir.GoType{Name: "Pet", IsPointer: true}, "*Pet"},
		{ir.GoType{Name: "Pet", IsSlice: true}, "[]Pet"},
	}
	for _, tc := range tests {
		if got := tc.gt.String(); got != tc.want {
			t.Errorf("GoType%+v.String() = %q, want %q", tc.gt, got, tc.want)
		}
	}
}

func TestGoTypeNilable(t *testing.T) {
	tests := []struct {
		gt   ir.GoType
		want string
	}{
		{ir.GoType{Name: "Pet"}, "*Pet"},
		{ir.GoType{Name: "Pet", IsSlice: true}, "[]Pet"},
		{ir.GoType{Name: "Pet", IsPointer: true}, "*Pet"},
		{ir.GoType{Name: "Pet", IsArrayOfSize: 3}, "[3]Pet"},
	}
	for _, tc := range tests {
		if got := tc.gt.Nilable(); got != tc.want {
			t.Errorf("GoType%+v.Nilable() = %q, want %q", tc.gt, got, tc.want)
		}
	}
}

func TestSchemaGoType_ArrayNilItems(t *testing.T) {
	s := &openapi.Schema{Type: openapi.TypeArray, Items: nil}
	got, err := ir.SchemaGoType(s)
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "[]any" {
		t.Errorf("got %q, want []any", got)
	}
}

func TestSchemaGoType_ObjectNoProps(t *testing.T) {
	s := &openapi.Schema{Type: openapi.TypeObject}
	got, err := ir.SchemaGoType(s)
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "struct{}" {
		t.Errorf("got %q, want any", got)
	}
}

func TestFromComponentSchemas_StructWithArrayField(t *testing.T) {
	items := &openapi.SchemaRef{}
	items.Value = &openapi.Schema{Type: openapi.TypeString}
	tagsRef := &openapi.SchemaRef{}
	tagsRef.Value = &openapi.Schema{Type: openapi.TypeArray, Items: items}

	props := openapi.SchemaRefs{}
	props.Set("tags", tagsRef)

	schemas := openapi.Schemas{}
	schemas.Set("Pet", &openapi.Schema{
		Type:       openapi.TypeObject,
		Properties: props,
	})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 || len(got[0].Fields) == 0 {
		t.Fatal("no fields")
	}
	f := got[0].Fields[0]
	if f.Type != "[]string" {
		t.Errorf("Type = %q, want []string", f.Type)
	}
	if want := `json:"tags,omitzero"`; f.JSONTag != want {
		t.Errorf("JSONTag = %q, want %q", f.JSONTag, want)
	}
}

func TestFromComponentSchemas_StructRequiredNonString(t *testing.T) {
	countRef := &openapi.SchemaRef{}
	countRef.Value = &openapi.Schema{Type: openapi.TypeInteger}
	props := openapi.SchemaRefs{}
	props.Set("count", countRef)

	schemas := openapi.Schemas{}
	schemas.Set("Stats", &openapi.Schema{
		Type:       openapi.TypeObject,
		Properties: props,
		Required:   []string{"count"},
	})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 || len(got[0].Fields) == 0 {
		t.Fatal("no fields")
	}
	f := got[0].Fields[0]
	if f.JSONTag != `json:"count"` {
		t.Errorf("JSONTag = %q, want json:\"count\"", f.JSONTag)
	}
}

func TestFieldGoName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"id", "ID"}, // initialism
		{"user_name", "UserName"},
		{"_id", "UnderscoreID"},
		{"created-at", "CreatedAt"},
		{"content-type", "ContentType"},
		{"C#", "CSharp"},
		{"1080p", "OneZeroEightZeroP"},
		{"4K", "FourK"},
		{"x+y", "XPlusY"},
		{"url.path", "URLDotPath"}, // initialism
		{"pdf", "PDF"},
		{"error", "Err"}, // avoid colliding with the error interface's Error() method
		{"Error", "Err"},
		{"ERROR", "Err"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			schemas := openapi.Schemas{}
			propRef := &openapi.SchemaRef{}
			propRef.Value = &openapi.Schema{Type: openapi.TypeString}
			props := openapi.SchemaRefs{}
			props.Set(tc.in, propRef)
			schemas.Set("Test", &openapi.Schema{
				Type:       openapi.TypeObject,
				Properties: props,
			})
			got, err := ir.FromComponentSchemas(schemas)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) == 0 || len(got[0].Fields) == 0 {
				t.Fatal("no fields generated")
			}
			if got[0].Fields[0].Name != tc.want {
				t.Errorf("field name for %q = %q, want %q", tc.in, got[0].Fields[0].Name, tc.want)
			}
		})
	}
}

// dateTimeOrIntSchema returns a schema representing a oneOf of a date-time
// string and an integer, in the given order (0 = date-time first, 1 = int first).
func dateTimeOrIntSchema(intFirst bool) *openapi.Schema {
	dt := &openapi.SchemaRef{}
	dt.Value = &openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatDateTime}
	num := &openapi.SchemaRef{}
	num.Value = &openapi.Schema{Type: openapi.TypeInteger}
	if intFirst {
		return &openapi.Schema{OneOf: openapi.SchemaRefList{num, dt}}
	}
	return &openapi.Schema{OneOf: openapi.SchemaRefList{dt, num}}
}

func TestSchemaGoType_DateTimeOrIntOneOf(t *testing.T) {
	for _, intFirst := range []bool{false, true} {
		got, err := ir.SchemaGoType(dateTimeOrIntSchema(intFirst))
		if err != nil {
			t.Fatalf("intFirst=%v: %v", intFirst, err)
		}
		if got.Name != "time.Time" {
			t.Errorf("intFirst=%v: got %q, want time.Time", intFirst, got.Name)
		}
	}
}

func TestSchemaGoType_UnrelatedOneOfFallsBackToAny(t *testing.T) {
	// An inline (unnamed) oneOf that doesn't match the date-time+int pattern
	// has nowhere to generate a named union struct, so it falls back to any.
	// A named oneOf in this shape instead gets a real pointer-bag type — see
	// TestFromComponentSchemas_OneOfEmitsUnionType.
	a := &openapi.SchemaRef{}
	a.Value = &openapi.Schema{Type: openapi.TypeString}
	b := &openapi.SchemaRef{}
	b.Value = &openapi.Schema{Type: openapi.TypeBoolean}
	s := &openapi.Schema{OneOf: openapi.SchemaRefList{a, b}}
	got, err := ir.SchemaGoType(s)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "any" {
		t.Errorf("got %q, want any", got.Name)
	}
}

func TestFromComponentSchemas_DateTimeOrIntField(t *testing.T) {
	prop := &openapi.SchemaRef{Value: dateTimeOrIntSchema(false)}
	props := openapi.SchemaRefs{}
	props.Set("timestamp", prop)

	schemas := openapi.Schemas{}
	schemas.Set("Event", &openapi.Schema{
		Type:       openapi.TypeObject,
		Properties: props,
		Required:   []string{"timestamp"},
	})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0].Fields) != 1 {
		t.Fatalf("unexpected schemas/fields: %+v", got)
	}

	f := got[0].Fields[0]
	if f.Type != "time.Time" {
		t.Errorf("Type = %q, want time.Time", f.Type)
	}
	if !f.IsDateTimeOrInt {
		t.Error("IsDateTimeOrInt = false, want true")
	}
}

func TestSchemaRefGoType_DateTimeOrIntOneOfViaRef(t *testing.T) {
	// After flattening, a date-time-or-int oneOf may live in components and be
	// reached by $ref. It should still resolve to time.Time, not a synthetic name.
	ref := &openapi.SchemaRef{}
	ref.Ref = &openapi.Reference{Identifier: "#/components/schemas/SomeDate"}
	ref.Value = dateTimeOrIntSchema(false)

	got, err := ir.SchemaRefGoType(ref)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "time.Time" {
		t.Errorf("got %q, want time.Time", got.Name)
	}
}

// anyOfUnionSchema returns an untagged-union schema like:
//
//	anyOf: [A, B, C]
//
// with no type/allOf/oneOf of its own. As a named component schema, this
// generates a pointer-bag union type (see fromUnionSchema); as an inline
// schema reached only via SchemaGoType (no name to give a struct), it falls
// back to `any`.
func anyOfUnionSchema() *openapi.Schema {
	a := &openapi.SchemaRef{Value: &openapi.Schema{Type: openapi.TypeObject}}
	a.Ref = &openapi.Reference{Identifier: "#/components/schemas/A"}
	b := &openapi.SchemaRef{Value: &openapi.Schema{Type: openapi.TypeObject}}
	b.Ref = &openapi.Reference{Identifier: "#/components/schemas/B"}
	return &openapi.Schema{AnyOf: openapi.SchemaRefList{a, b}}
}

func TestSchemaGoType_AnyOfOnly(t *testing.T) {
	// Inline (unnamed) anyOf: nowhere to generate a struct, falls back to any.
	got, err := ir.SchemaGoType(anyOfUnionSchema())
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "any" {
		t.Errorf("got %q, want any", got.Name)
	}
}

func TestSchemaRefGoType_AnyOfOnlyViaRef(t *testing.T) {
	// A named anyOf union reached via $ref resolves to its own generated
	// pointer-bag type name, not `any` — see TestFromComponentSchemas_AnyOfEmitsUnionType.
	ref := &openapi.SchemaRef{}
	ref.Ref = &openapi.Reference{Identifier: "#/components/schemas/SomeUnion"}
	ref.Value = anyOfUnionSchema()

	got, err := ir.SchemaRefGoType(ref)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "SomeUnion" {
		t.Errorf("got %q, want SomeUnion", got.Name)
	}
}

func TestFromComponentSchemas_AnyOfEmitsUnionType(t *testing.T) {
	schemas := openapi.Schemas{}
	schemas.Set("SomeUnion", anyOfUnionSchema())

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 schema, got %d: %+v", len(got), got)
	}

	s := got[0]
	if s.Kind != ir.SchemaKindUnion {
		t.Errorf("Kind = %v, want SchemaKindUnion", s.Kind)
	}
	if s.IsOneOf {
		t.Error("IsOneOf = true, want false for anyOf")
	}
	if len(s.UnionVariants) != 2 {
		t.Fatalf("expected 2 variants, got %d: %+v", len(s.UnionVariants), s.UnionVariants)
	}
	if s.UnionVariants[0].FieldName != "A" || s.UnionVariants[0].Type != "A" {
		t.Errorf("UnionVariants[0] = %+v, want {FieldName:A Type:A}", s.UnionVariants[0])
	}
	if s.UnionVariants[1].FieldName != "B" || s.UnionVariants[1].Type != "B" {
		t.Errorf("UnionVariants[1] = %+v, want {FieldName:B Type:B}", s.UnionVariants[1])
	}
}

func TestFromComponentSchemas_OneOfEmitsUnionType(t *testing.T) {
	a := &openapi.SchemaRef{Value: &openapi.Schema{Type: openapi.TypeObject}}
	a.Ref = &openapi.Reference{Identifier: "#/components/schemas/Card"}
	b := &openapi.SchemaRef{Value: &openapi.Schema{Type: openapi.TypeObject}}
	b.Ref = &openapi.Reference{Identifier: "#/components/schemas/Bank"}

	schemas := openapi.Schemas{}
	schemas.Set("Payment", &openapi.Schema{OneOf: openapi.SchemaRefList{a, b}})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 schema, got %d: %+v", len(got), got)
	}

	s := got[0]
	if s.Kind != ir.SchemaKindUnion {
		t.Errorf("Kind = %v, want SchemaKindUnion", s.Kind)
	}
	if !s.IsOneOf {
		t.Error("IsOneOf = false, want true for oneOf")
	}
	if len(s.UnionVariants) != 2 {
		t.Fatalf("expected 2 variants, got %d: %+v", len(s.UnionVariants), s.UnionVariants)
	}
	if s.UnionVariants[0].FieldName != "Card" {
		t.Errorf("UnionVariants[0].FieldName = %q, want Card", s.UnionVariants[0].FieldName)
	}
	if s.UnionVariants[1].FieldName != "Bank" {
		t.Errorf("UnionVariants[1].FieldName = %q, want Bank", s.UnionVariants[1].FieldName)
	}
}

func TestFromComponentSchemas_OneOfVariantFieldNameCollision(t *testing.T) {
	// Two variants that resolve to the same base field name get deduped.
	a := &openapi.SchemaRef{Value: &openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatUUID}}
	b := &openapi.SchemaRef{Value: &openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatUUID}}

	schemas := openapi.Schemas{}
	schemas.Set("Either", &openapi.Schema{OneOf: openapi.SchemaRefList{a, b}})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0].UnionVariants) != 2 {
		t.Fatalf("unexpected schemas: %+v", got)
	}
	if got[0].UnionVariants[0].FieldName != "UUID" {
		t.Errorf("UnionVariants[0].FieldName = %q, want UUID", got[0].UnionVariants[0].FieldName)
	}
	if got[0].UnionVariants[1].FieldName != "UUID2" {
		t.Errorf("UnionVariants[1].FieldName = %q, want UUID2", got[0].UnionVariants[1].FieldName)
	}
}

func TestFromSchema_DateTimeOrIntOneOfNamedSchemaEmitsNothing(t *testing.T) {
	// A named component whose whole schema is the date-time-or-int oneOf
	// pattern still emits no schema of its own — $ref resolution redirects
	// straight to time.Time (see TestSchemaRefGoType_DateTimeOrIntOneOfViaRef).
	schemas := openapi.Schemas{}
	schemas.Set("SomeDate", dateTimeOrIntSchema(false))

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no named schema for the date-time-or-int oneOf, got %+v", got)
	}
}

func TestSchemaGoType_AnyOfWithAllOfIsNotAnyOfOnly(t *testing.T) {
	// allOf alongside anyOf means this isn't a pure untagged union; the
	// existing "unsupported schema type" error should still surface.
	s := anyOfUnionSchema()
	s.AllOf = openapi.SchemaRefList{&openapi.SchemaRef{Value: &openapi.Schema{Type: openapi.TypeObject}}}
	if _, err := ir.SchemaGoType(s); err == nil {
		t.Fatal("expected error for anyOf combined with allOf")
	}
}

func TestFromComponentSchemas_PlainDateTimeFieldNotFlagged(t *testing.T) {
	// A regular string+date-time property must not set IsDateTimeOrInt.
	prop := &openapi.SchemaRef{}
	prop.Value = &openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatDateTime}
	props := openapi.SchemaRefs{}
	props.Set("createdAt", prop)

	schemas := openapi.Schemas{}
	schemas.Set("Event", &openapi.Schema{
		Type:       openapi.TypeObject,
		Properties: props,
	})

	got, err := ir.FromComponentSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0].Fields) != 1 {
		t.Fatalf("unexpected schemas/fields: %+v", got)
	}
	if got[0].Fields[0].IsDateTimeOrInt {
		t.Error("plain date-time field: IsDateTimeOrInt = true, want false")
	}
}
