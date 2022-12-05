package graphql

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"sort"

	lls "github.com/emirpasic/gods/stacks/linkedliststack"
	"github.com/r0busta/graphql/ident"
	"github.com/thoas/go-funk"
)

func constructQuery(v interface{}, variables map[string]interface{}) string {
	query := query(v)
	if len(variables) > 0 {
		return "query(" + queryArguments(variables) + ")" + query
	}
	return query
}

func constructMutation(v interface{}, variables map[string]interface{}) string {
	query := query(v)
	if len(variables) > 0 {
		return "mutation(" + queryArguments(variables) + ")" + query
	}
	return "mutation" + query
}

// queryArguments constructs a minified arguments string for variables.
//
// E.g., map[string]interface{}{"a": Int(123), "b": NewBoolean(true)} -> "$a:Int!$b:Boolean".
func queryArguments(variables map[string]interface{}) string {
	// Sort keys in order to produce deterministic output for testing purposes.
	// TODO: If tests can be made to work with non-deterministic output, then no need to sort.
	keys := make([]string, 0, len(variables))
	for k := range variables {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	for _, k := range keys {
		io.WriteString(&buf, "$")
		io.WriteString(&buf, k)
		io.WriteString(&buf, ":")
		writeArgumentType(&buf, reflect.TypeOf(variables[k]), true)
		// Don't insert a comma here.
		// Commas in GraphQL are insignificant, and we want minified output.
		// See https://facebook.github.io/graphql/October2016/#sec-Insignificant-Commas.
	}
	return buf.String()
}

// writeArgumentType writes a minified GraphQL type for t to w.
// value indicates whether t is a value (required) type or pointer (optional) type.
// If value is true, then "!" is written at the end of t.
func writeArgumentType(w io.Writer, t reflect.Type, value bool) {
	if t.Kind() == reflect.Ptr {
		// Pointer is an optional type, so no "!" at the end of the pointer's underlying type.
		writeArgumentType(w, t.Elem(), false)
		return
	}

	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		// List. E.g., "[Int]".
		io.WriteString(w, "[")
		writeArgumentType(w, t.Elem(), true)
		io.WriteString(w, "]")
	default:
		// Named type. E.g., "Int".
		name := t.Name()
		if name == "string" { // HACK: Workaround for https://github.com/shurcooL/githubv4/issues/12.
			name = "ID"
		}
		io.WriteString(w, name)
	}

	if value {
		// Value is a required type, so add "!" to the end.
		io.WriteString(w, "!")
	}
}

// query uses writeQuery to recursively construct
// a minified query string from the provided struct v.
//
// E.g., struct{Foo Int, BarBaz *Boolean} -> "{foo,barBaz}".
func query(v interface{}) string {
	var buf bytes.Buffer
	visited := lls.New()
	writeQuery(&buf, reflect.TypeOf(v), false, visited)
	return buf.String()
}

// writeQuery writes a minified query for t to w.
// If inline is true, the struct fields of t are inlined into parent struct.
func writeQuery(w io.Writer, t reflect.Type, inline bool, visited *lls.Stack) {
	switch t.Kind() {
	case reflect.Ptr, reflect.Slice:
		writeQuery(w, t.Elem(), false, visited)
	case reflect.Struct:
		visited.Push(t)

		// If the type implements json.Unmarshaler, it's a scalar. Don't expand it.
		if reflect.PtrTo(t).Implements(jsonUnmarshaler) {
			visited.Pop()
			return
		}

		if !inline {
			io.WriteString(w, "{")
		}
		first := true
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			baseFieldType := field.Type
			if baseFieldType.Kind() == reflect.Ptr || baseFieldType.Kind() == reflect.Slice {
				baseFieldType = baseFieldType.Elem()
			}

			if funk.Contains(visited.Values(), baseFieldType) {
				continue
			}

			if baseFieldType.Kind() == reflect.Struct && baseFieldType.NumField() == 1 && structContainsVisitedType(visited, baseFieldType) {
				continue
			}

			if !first {
				io.WriteString(w, ",")
			}

			value, ok := field.Tag.Lookup("graphql")

			inlineField := field.Anonymous && !ok
			if !inlineField {
				if ok {
					io.WriteString(w, value)
				} else {
					io.WriteString(w, ident.ParseMixedCaps(field.Name).ToLowerCamelCase())
				}
			}

			if !reflect.PtrTo(field.Type).Implements(jsonUnmarshaler) {
				visited.Push(baseFieldType)
			}
			writeQuery(w, field.Type, inlineField, visited)
			if !reflect.PtrTo(field.Type).Implements(jsonUnmarshaler) {
				visited.Pop()
			}

			first = false
		}
		if !inline {
			io.WriteString(w, "}")
		}

		visited.Pop()
	}
}

func structContainsVisitedType(visited *lls.Stack, t reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i).Type
		if ft.Kind() == reflect.Ptr || ft.Kind() == reflect.Slice {
			ft = ft.Elem()
		}

		if funk.Contains(visited.Values(), ft) {
			return true
		}
	}

	return false
}

var jsonUnmarshaler = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
