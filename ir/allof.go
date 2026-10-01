package ir

import (
	"encoding/json/v2"
	"fmt"
	"slices"
	"strings"

	"github.com/MarkRosemaker/openapi"
)

// objectShape returns the members an object schema declares and those it requires, following references and allOf.
// ok is false for a schema that is not a plain object, such as a union.
func objectShape(s *openapi.Schema) (members, required []string, ok bool) {
	s = deref(s)
	if s == nil || len(s.OneOf) > 0 || len(s.AnyOf) > 0 || s.Type != "" && s.Type != openapi.TypeObject {
		return nil, nil, false
	}

	for name := range s.Properties.ByIndex() {
		members = append(members, name)
	}

	required = slices.Clone(s.Required)

	for _, e := range s.AllOf {
		m, r, ok := objectShape(e)
		if !ok {
			return nil, nil, false
		}

		members, required = append(members, m...), append(required, r...)
	}

	slices.Sort(members)
	slices.Sort(required)

	return slices.Compact(members), slices.Compact(required), true
}

// constString returns the string a schema allows as its only value, from const or a single enum value.
func constString(s *openapi.Schema) (string, bool) {
	s = deref(s)

	v := s.Const
	if len(v) == 0 && len(s.Enum) == 1 {
		v = s.Enum[0]
	}

	var str string
	if len(v) == 0 || json.Unmarshal(v, &str) != nil {
		return "", false
	}

	return str, true
}

// propertyOf returns the schema of the member name that s declares, following references and allOf.
func propertyOf(s *openapi.Schema, name string) *openapi.Schema {
	s = deref(s)
	if p, ok := s.Properties[name]; ok {
		return p
	}

	for _, e := range s.AllOf {
		if p := propertyOf(e, name); p != nil {
			return p
		}
	}

	return nil
}

// discriminate returns the member that tells a union's variants apart and each variant's value of it: the
// discriminator's propertyName, or a member every variant declares with a string const of its own. ok is false if
// there is no such member.
func discriminate(u *openapi.Schema, variants openapi.SchemaList) (member string, values []string, ok bool) {
	if d := u.Discriminator; d != nil {
		values = make([]string, len(variants))
		for i, v := range variants {
			if v.Ref == nil {
				return "", nil, false
			}

			// an explicit mapping wins over the component's name, its first entry if several name the variant
			values[i] = v.Ref.Identifier[strings.LastIndex(v.Ref.Identifier, "/")+1:]
			for key, target := range d.Mapping.ByIndex() {
				if openapi.MappingRef(target.Value) == v.Ref.Identifier {
					values[i] = key
					break
				}
			}
		}

		return d.PropertyName, values, distinct(values)
	}

	if len(variants) == 0 {
		return "", nil, false
	}

	first, _, ok := objectShape(variants[0])
	if !ok {
		return "", nil, false
	}

	for _, name := range first {
		values = values[:0]

		for _, v := range variants {
			p := propertyOf(v, name)
			if p == nil {
				break
			}

			value, ok := constString(p)
			if !ok {
				break
			}

			values = append(values, value)
		}

		if len(values) == len(variants) && distinct(values) {
			return name, values, true
		}
	}

	return "", nil, false
}

func distinct(values []string) bool {
	seen := make(map[string]bool, len(values))
	for _, v := range values {
		if seen[v] {
			return false
		}

		seen[v] = true
	}

	return true
}

// hasJSONMethods reports whether the Go type generated for s encodes itself, so that embedding it would hand its
// methods to the struct embedding it.
func hasJSONMethods(s *openapi.Schema) bool {
	s = deref(s)

	switch {
	case s == nil, isDateTimeOrIntegerOneOf(s), nullableVariant(s) != nil:
		return false
	case len(s.PrefixItems) > 0, s.Type == "" && (len(s.OneOf) > 0 || len(s.AnyOf) > 0):
		return true
	default:
		return slices.ContainsFunc(s.AllOf, hasJSONMethods)
	}
}

// isFoldable reports whether s is a plain object whose properties can be written as fields of the struct of an
// allOf it is part of.
func isFoldable(s *openapi.Schema) bool {
	s = deref(s)

	return s != nil && (s.Type == openapi.TypeObject || s.Type == "") && len(s.Properties) > 0 &&
		len(s.AllOf) == 0 && len(s.OneOf) == 0 && len(s.AnyOf) == 0 && mapValues(s) == nil
}

// countSchemaUses counts the references to each component schema anywhere in doc, by reference identifier.
func countSchemaUses(doc *openapi.Document) (map[string]int, error) {
	data, err := doc.ToJSON()
	if err != nil {
		return nil, err
	}

	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	uses := map[string]int{}

	var walk func(any)

	walk = func(v any) {
		switch v := v.(type) {
		case map[string]any:
			if ref, ok := v["$ref"].(string); ok && strings.HasPrefix(ref, "#/components/schemas/") {
				uses[ref]++
			}

			for _, e := range v {
				walk(e)
			}
		case []any:
			for _, e := range v {
				walk(e)
			}
		}
	}

	walk(v)

	return uses, nil
}

// unionVariants builds the variants of the union u: one pointer field per alternative but null.
func unionVariants(u *openapi.Schema, variants openapi.SchemaList) ([]UnionVariant, string, error) {
	member, values, discriminated := discriminate(u, variants)

	used := make(map[string]bool, len(variants))
	out := make([]UnionVariant, 0, len(variants))

	for i, v := range variants {
		// every field nil already means null, so a null variant needs no field
		if isNull(v) {
			continue
		}

		tp, err := SchemaGoType(v)
		if err != nil {
			return nil, "", fmt.Errorf("variant %d: %w", i, err)
		}

		base := unionVariantFieldName(tp, i)

		fieldName := base
		for n := 2; used[fieldName]; n++ {
			fieldName = fmt.Sprintf("%s%d", base, n)
		}

		used[fieldName] = true

		uv := UnionVariant{FieldName: fieldName, Type: tp.String()}
		if discriminated {
			uv.Value = values[i]
		}

		uv.Members, uv.Required, _ = objectShape(v)
		out = append(out, uv)
	}

	if !discriminated {
		member = ""
	}

	return out, member, nil
}

// fromAllOfSchema builds the struct of an allOf. A part that only this schema uses is folded into its fields, any
// other is embedded, and a union among the parts is held as a field of its own, decoded by the struct's own methods.
func fromAllOfSchema(name string, s *openapi.Schema, uses map[string]int, folded map[string]bool) (*Schema, error) {
	requiredSet := make(map[string]bool)
	for _, r := range s.Required {
		requiredSet[r] = true
	}

	out := &Schema{
		Name:        name,
		Description: getDescription(s, name),
		Kind:        SchemaKindAllOf,
	}

	addProperties := func(part *openapi.Schema) error {
		for _, r := range part.Required {
			requiredSet[r] = true
		}

		for jsonName, propRef := range part.Properties.ByIndex() {
			field, err := getField(jsonName, propRef, requiredSet)
			if err != nil {
				return fmt.Errorf("allOf property %q: %w", jsonName, err)
			}

			out.Fields = append(out.Fields, field)
			out.Members = append(out.Members, jsonName)
		}

		return nil
	}

	var unions []*openapi.Schema

	for _, entry := range s.AllOf {
		part := deref(entry)

		switch {
		case part == nil: // not resolved, so nothing is known of it but its name
			typeName, err := SchemaGoType(entry)
			if err != nil {
				return nil, err
			}

			out.Fields = append(out.Fields, Field{Type: typeName.String(), Embedded: true})
		case len(part.OneOf) > 0 || len(part.AnyOf) > 0:
			if entry.Ref == nil {
				out.Unimplemented = "an allOf with an inline union"
				continue
			}

			unions = append(unions, entry)
		case entry.Ref == nil:
			if err := addProperties(entry); err != nil {
				return nil, err
			}
		case uses[entry.Ref.Identifier] == 1 && isFoldable(part):
			folded[entry.Ref.Identifier] = true

			if err := addProperties(part); err != nil {
				return nil, err
			}
		case hasJSONMethods(part):
			out.Unimplemented = "an allOf embedding a part that encodes itself"
		default:
			typeName, err := SchemaGoType(entry)
			if err != nil {
				return nil, err
			}

			out.Fields = append(out.Fields, Field{Type: typeName.String(), Embedded: true})

			members, _, _ := objectShape(part)
			out.Members = append(out.Members, members...)
		}
	}

	switch len(unions) {
	case 0:
		return out, nil
	case 1:
	default:
		out.Unimplemented = "an allOf of more than one union"
		return out, nil
	}

	entry := unions[0]
	u := deref(entry)

	isOneOf := len(u.OneOf) > 0
	alts := u.AnyOf
	if isOneOf {
		alts = u.OneOf
	}

	tp, err := SchemaGoType(entry)
	if err != nil {
		return nil, err
	}

	variants, member, err := unionVariants(u, alts)
	if err != nil {
		return nil, err
	}

	if slices.ContainsFunc(alts, isNull) {
		out.Unimplemented = "an allOf whose union can be null"
	}

	for _, v := range alts {
		if _, _, ok := objectShape(v); !ok && !isNull(v) {
			out.Unimplemented = "an allOf whose union has an alternative that is not a plain object"
		}
	}

	slices.Sort(out.Members)
	out.Members = slices.Compact(out.Members)

	out.Fields = append(out.Fields, Field{Name: tp.Name, Type: tp.Name, JSONTag: `json:"-"`, Required: true})
	out.AllOfUnion = &AllOfUnion{
		FieldName:     tp.Name,
		IsOneOf:       isOneOf,
		Discriminator: member,
		Variants:      variants,
	}

	return out, nil
}
