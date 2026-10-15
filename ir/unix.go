package ir

import (
	"fmt"
	"strings"
)

// unixTypes are the Go types an integer of unix time can be, with the jsonutil type that writes one field of them as
// such.
var unixTypes = map[string]string{
	"time.Time":  "jsonutil.TimeIntUnix",
	"civil.Date": "jsonutil.DateIntUnix",
}

// markUnixFields decides how the date-times and dates that are integers of unix time are written and read. Where every
// time.Time, or every civil.Date, a struct holds is one, a marshaler for the type writes them all: allTime and allDate
// say so. Where they are mixed with strings, each struct lists its own in [Schema.Unix], with those of the parts it
// embeds, which encoding inlines.
func markUnixFields(schemas []Schema) (allTime, allDate bool, err error) {
	unix, other := map[string]bool{}, map[string]bool{}

	for _, s := range schemas {
		for _, f := range s.Fields {
			for goType := range unixTypes {
				switch {
				case f.IsUnix && strings.TrimLeft(f.Type, "*[]") == goType:
					unix[goType] = true
				case f.IsDateTimeOrInt:
					// either form will do, so it is written as the others are, or as a string where they differ
				case containsType(f.Type, goType):
					other[goType] = true
				}
			}
		}
	}

	perField := map[string]bool{}
	for goType := range unixTypes {
		perField[goType] = unix[goType] && other[goType]
	}

	byName := make(map[string]*Schema, len(schemas))
	for i := range schemas {
		byName[schemas[i].Name] = &schemas[i]
	}

	var fields func(s *Schema, seen map[string]bool) ([]UnixField, error)

	fields = func(s *Schema, seen map[string]bool) ([]UnixField, error) {
		var out []UnixField

		for _, f := range s.Fields {
			if !f.Embedded {
				if !f.IsUnix || !perField[strings.TrimLeft(f.Type, "*[]")] {
					continue
				}

				tp, ok := unixTypes[f.Type]
				if !ok {
					return nil, fmt.Errorf("field %s: an integer %s among strings is only supported as a value of its own, not as %s",
						f.Name, strings.TrimLeft(f.Type, "*[]"), f.Type)
				}

				out = append(out, UnixField{Selector: f.Name, JSONTag: f.JSONTag, Type: tp})

				continue
			}

			part := byName[strings.TrimPrefix(f.Type, "*")]
			if part == nil || seen[part.Name] {
				continue
			}

			seen[part.Name] = true

			inner, err := fields(part, seen)
			if err != nil {
				return nil, err
			}

			if len(inner) > 0 && strings.HasPrefix(f.Type, "*") {
				return nil, fmt.Errorf("embedded %s: a part with an integer date among strings is only supported embedded as a value, not as a pointer", f.Type)
			}

			for _, u := range inner {
				u.Selector = part.Name + "." + u.Selector
				out = append(out, u)
			}
		}

		return out, nil
	}

	for i := range schemas {
		s := &schemas[i]

		s.Unix, err = fields(s, map[string]bool{s.Name: true})
		if err != nil {
			return false, false, fmt.Errorf("%s: %w", s.Name, err)
		}

		if len(s.Unix) > 0 && (s.EncodesItself() || s.MemberDecoder) {
			return false, false, fmt.Errorf("%s: an integer date among strings is not supported in a type that encodes itself", s.Name)
		}
	}

	return unix["time.Time"] && !other["time.Time"], unix["civil.Date"] && !other["civil.Date"], nil
}
