package ir_test

import (
	"testing"

	"github.com/MarkRosemaker/openapi"
	"github.com/MarkRosemaker/openapi-codegen/ir"
)

func buildMinimalDoc() *openapi.Document {
	// Build a minimal valid document:
	//   GET /pets            → listPets (returns PetList)
	//   GET /pets/{petId}    → getPet   (returns Pet)
	//   POST /pets           → createPet (body: NewPet, returns Pet)

	petSchema := &openapi.Schema{
		Type:        openapi.TypeObject,
		Description: "A pet",
		Properties:  openapi.SchemaRefs{},
		Required:    []string{"id", "name"},
	}
	idRef := &openapi.SchemaRef{}
	idRef.Value = &openapi.Schema{Type: openapi.TypeInteger}
	petSchema.Properties.Set("id", idRef)
	nameRef := &openapi.SchemaRef{}
	nameRef.Value = &openapi.Schema{Type: openapi.TypeString}
	petSchema.Properties.Set("name", nameRef)

	petListRef := makeNamedRef("Pet")
	petListSchema := &openapi.Schema{Type: openapi.TypeArray, Items: petListRef}

	newPetSchema := &openapi.Schema{
		Type:       openapi.TypeObject,
		Properties: openapi.SchemaRefs{},
		Required:   []string{"name"},
	}
	newPetNameRef := &openapi.SchemaRef{}
	newPetNameRef.Value = &openapi.Schema{Type: openapi.TypeString}
	newPetSchema.Properties.Set("name", newPetNameRef)

	doc := &openapi.Document{
		OpenAPI: "3.1.0",
		Info:    &openapi.Info{Title: "Pet API", Version: "1.0"},
		Servers: openapi.Servers{
			{URL: "https://api.example.com/v1"},
		},
	}

	// Add schemas to components.
	doc.Components.Schemas = openapi.Schemas{}
	doc.Components.Schemas.Set("Pet", petSchema)
	doc.Components.Schemas.Set("PetList", petListSchema)
	doc.Components.Schemas.Set("NewPet", newPetSchema)

	// GET /pets → listPets
	listPetsOp := &openapi.Operation{OperationID: "listPets", Summary: "List pets"}
	listPetsOp.Responses = openapi.OperationResponses{}
	listPetsOp.Responses.Set("200", makeResponse("OK", "application/json", makeNamedRef("PetList")))

	// GET /pets/{petId} → getPet
	petIDParam := makeParam("petId", openapi.ParameterLocationPath, true, &openapi.Schema{Type: openapi.TypeInteger})
	getPetOp := &openapi.Operation{
		OperationID: "getPet",
		Parameters:  openapi.ParameterList{petIDParam},
	}
	getPetOp.Responses = openapi.OperationResponses{}
	getPetOp.Responses.Set("200", makeResponse("OK", "application/json", makeNamedRef("Pet")))

	// POST /pets → createPet
	newPetMT := &openapi.MediaType{Schema: makeNamedRef("NewPet")}
	newPetContent := openapi.Content{}
	newPetContent.Set("application/json", newPetMT)
	createRB := &openapi.RequestBodyRef{}
	createRB.Value = &openapi.RequestBody{Content: newPetContent, Required: true}
	createPetOp := &openapi.Operation{OperationID: "createPet", RequestBody: createRB}
	createPetOp.Responses = openapi.OperationResponses{}
	createPetOp.Responses.Set("201", makeResponse("Created", "application/json", makeNamedRef("Pet")))

	doc.Paths = openapi.Paths{}
	doc.Paths.Set("/pets", &openapi.PathItem{Get: listPetsOp, Post: createPetOp})
	doc.Paths.Set("/pets/{petId}", &openapi.PathItem{Get: getPetOp})

	return doc
}

func TestFromDocument_Basic(t *testing.T) {
	doc := buildMinimalDoc()

	irDoc, err := ir.FromDocument(doc, "petapi", "")
	if err != nil {
		t.Fatal(err)
	}

	if irDoc.PackageName != "petapi" {
		t.Errorf("PackageName = %q, want petapi", irDoc.PackageName)
	}
	if irDoc.BaseURL.Scheme != "https" {
		t.Errorf("BaseURL.Scheme = %q, want https", irDoc.BaseURL.Scheme)
	}
	if irDoc.BaseURL.Host != "api.example.com" {
		t.Errorf("BaseURL.Host = %q, want api.example.com", irDoc.BaseURL.Host)
	}
	if irDoc.BaseURL.Path != "/v1/pets" {
		t.Errorf("BaseURL.Path = %q, want /v1", irDoc.BaseURL.Path)
	}
	if irDoc.UserAgent != "Pet API" {
		t.Errorf("UserAgent = %q, want 'Pet API'", irDoc.UserAgent)
	}

	// 3 named schemas: Pet, PetList, NewPet
	if len(irDoc.Schemas) != 3 {
		t.Errorf("Schemas len = %d, want 3", len(irDoc.Schemas))
	}

	// 3 operations: listPets, getPet, createPet
	if len(irDoc.Operations) != 3 {
		t.Errorf("Operations len = %d, want 3", len(irDoc.Operations))
	}
}

func TestFromDocument_OperationDetails(t *testing.T) {
	doc := buildMinimalDoc()
	irDoc, err := ir.FromDocument(doc, "petapi", "test-agent")
	if err != nil {
		t.Fatal(err)
	}

	// Find operations by name.
	opsByName := make(map[string]ir.Operation, len(irDoc.Operations))
	for _, op := range irDoc.Operations {
		opsByName[op.Name] = op
	}

	listPets, ok := opsByName["ListPets"]
	if !ok {
		t.Fatal("ListPets operation not found")
	}
	if listPets.Method != "GET" {
		t.Errorf("ListPets.Method = %q, want GET", listPets.Method)
	}
	if listPets.SuccessReturn == nil || listPets.SuccessReturn.Name != "PetList" {
		t.Errorf("ListPets.SuccessReturn = %v, want &GoType{Name:PetList}", listPets.SuccessReturn)
	}

	getPet, ok := opsByName["GetPet"]
	if !ok {
		t.Fatal("GetPet operation not found")
	}
	if len(getPet.PathParams) != 1 {
		t.Fatalf("GetPet.PathParams len = %d, want 1", len(getPet.PathParams))
	}
	if getPet.PathParams[0].GoName != "petID" {
		t.Errorf("GetPet.PathParams[0].GoName = %q, want petID", getPet.PathParams[0].GoName)
	}

	createPet, ok := opsByName["CreatePet"]
	if !ok {
		t.Fatal("CreatePet operation not found")
	}
	if createPet.RequestBody == nil {
		t.Fatal("CreatePet.RequestBody = nil")
	}
	if createPet.RequestBody.TypeName != "NewPet" {
		t.Errorf("CreatePet.RequestBody.TypeName = %q, want NewPet", createPet.RequestBody.TypeName)
	}

	if irDoc.UserAgent != "test-agent" {
		t.Errorf("UserAgent = %q, want test-agent", irDoc.UserAgent)
	}
}

func TestFromDocument_NoServers(t *testing.T) {
	doc := buildMinimalDoc()
	doc.Servers = nil

	irDoc, err := ir.FromDocument(doc, "pkg", "")
	if err != nil {
		t.Fatal(err)
	}
	if irDoc.BaseURL.Scheme != "https" {
		t.Errorf("BaseURL.Scheme = %q, want https (default)", irDoc.BaseURL.Scheme)
	}
}

func TestFromDocument_SpecialImports(t *testing.T) {
	// A doc with a url.URL field should set HasURLFields=true.
	urlSchema := &openapi.Schema{Type: openapi.TypeObject, Properties: openapi.SchemaRefs{}, Required: []string{"website"}}
	wsRef := &openapi.SchemaRef{}
	wsRef.Value = &openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatURI}
	urlSchema.Properties.Set("website", wsRef)

	doc := &openapi.Document{
		OpenAPI: "3.1.0",
		Info:    &openapi.Info{Title: "T", Version: "1"},
		Servers: openapi.Servers{{URL: "https://example.com"}},
	}
	doc.Components.Schemas = openapi.Schemas{}
	doc.Components.Schemas.Set("Thing", urlSchema)

	listOp := &openapi.Operation{OperationID: "listThings"}
	listOp.Responses = openapi.OperationResponses{}
	listOp.Responses.Set("200", makeResponse("OK", "", nil))
	doc.Paths = openapi.Paths{}
	doc.Paths.Set("/things", &openapi.PathItem{Get: listOp})

	irDoc, err := ir.FromDocument(doc, "pkg", "")
	if err != nil {
		t.Fatal(err)
	}
	if !irDoc.HasURLFields {
		t.Error("HasURLFields = false, want true")
	}
}

func TestFromDocument_APIKeyHeaders(t *testing.T) {
	// X-Rd-Token is an API token and is wired up exactly like X-Api-Key: a
	// client field, a ClientOption and a value read from the environment.
	for _, name := range []string{"X-Api-Key", "X-Rd-Token"} {
		t.Run(name, func(t *testing.T) {
			doc := buildMinimalDoc()

			// A parameter is only global when every path declares it.
			p := makeParam(name, openapi.ParameterLocationHeader, true,
				&openapi.Schema{Type: openapi.TypeString})
			for _, pi := range doc.Paths {
				pi.Parameters = openapi.ParameterList{p}
			}

			irDoc, err := ir.FromDocument(doc, "petapi", "")
			if err != nil {
				t.Fatal(err)
			}

			key := irDoc.APIKey()
			if key == nil {
				t.Fatal("APIKey() = nil, want the parameter")
			}

			if key.JSONName != name {
				t.Errorf("JSONName = %q, want %q", key.JSONName, name)
			}
			if key.VarName != "apiKey" {
				t.Errorf("VarName = %q, want apiKey", key.VarName)
			}
			if key.EnvName != "PET_API_KEY" {
				t.Errorf("EnvName = %q, want PET_API_KEY", key.EnvName)
			}
			if got := key.FormatExpr(); got != "c.apiKey" {
				t.Errorf("FormatExpr() = %q, want c.apiKey", got)
			}

			// Being global, it is served from the client rather than asked
			// for on every call.
			for _, op := range irDoc.Operations {
				for _, hp := range op.HeaderParams {
					if hp.JSONName == name {
						t.Errorf("%s is a header param of %s, want it global only",
							name, op.Name)
					}
				}
			}
		})
	}
}
