package yaml

import (
	"fmt"
	"reflect"
	"strings"
)

type YamlTextErrorCause string

const (
	// CauseUnknownField is used when an unexpected fields is in the yaml.
	CauseUnknownField YamlTextErrorCause = "unknown field"
	// CauseKeyAlreadyDefined can be caused by mapping or fields when the key appears
	// twice.
	CauseKeyAlreadyDefined YamlTextErrorCause = "key already defined"
	// CauseWrongType happens when the expected yaml node is incorrect.
	// E.g: a map when expecting a sequence
	CauseWrongType YamlTextErrorCause = "incorrect yaml node"
)

// YamlError is the top level detailed error.
type YamlError struct {
	// Cause will be a structured error with additional details
	// about the error itself.
	Cause error
	// Original is a 1 sentence error for brevity.
	Original error
}

func (w YamlError) Unwrap() error {
	return w.Cause
}

func (w YamlError) Error() string {
	return w.Cause.Error()
}

// GoLangStructError are errors that happen when decoding the golang struct,
// before the yaml is touched.
// These errors are usually internal server errors as it's a mistake
// on the Go dev, not the yaml document.
// TODO: @emyrk Flush these out more maybe.
type GoLangStructError struct {
	Err error
}

func (w GoLangStructError) Unwrap() error {
	return w.Err
}

func (w GoLangStructError) Error() string {
	return w.Err.Error()
}

// YamlTextError happens when mapping the yaml text to a golang struct.
// These errors concern the user who wrote the yaml document.
type YamlTextError struct {
	// Node is the yaml node when the error took place.
	Node Node
	// Name is the name of the field if the name cannot be inferred
	// from Node. Sometimes if a field fails, the `Node` does not have
	// the field's name, but the function context does.
	// If 'Name' is blank, then `Path()` will return the full path.
	Name string

	// Cause is the error that caused the TextError.
	Cause YamlTextErrorCause
	// To is the GoLang value the yaml text was attempted to be decoded into.
	To reflect.Value

	// Meta is extra fields that can be added if additional context is needed
	Meta map[string]string
}

// metaField is provided to handle arbitrary extra values
type metaField struct {
	Key   string
	Value string
}

// Error reconstructs the original error from the fields.
// This should probably be improved, for now it serves as an example to deconstruct the parts
// to get the info needed to maintain the current errors.
func (w YamlTextError) Error() string {
	path := strings.Join(w.Node.Path(), "->")
	var _ = path
	switch w.Cause {
	case CauseUnknownField:
		return fmt.Sprintf("line %d: field %s not found in type %s", w.Node.Line, w.Name, w.To.Type())
	case CauseKeyAlreadyDefined:
		if w.Name != "" {
			// Field already defined
			return fmt.Sprintf("line %d: field %s already set in type %s", w.Node.Line, w.Name, w.To.Type())
		}
		// Mapping already defined
		l := w.Meta["line_num"]
		return fmt.Sprintf("line %d: mapping key %#v already defined at line %s", w.Node.Line, w.Node.Value, l)
	case CauseWrongType:
		value := w.Node.Value
		tag := w.Node.Tag
		if tag != seqTag && tag != mapTag {
			if len(value) > 10 {
				value = " `" + value[:7] + "...`"
			} else {
				value = " `" + value + "`"
			}
		}

		return fmt.Sprintf("line %d: cannot unmarshal %s%s into %s", w.Node.Line, shortTag(w.Node.Tag), value, w.To.Type())
	}
	return fmt.Sprintf("this should never happen")
}

// TODO: @emyrk handle document/alias kinds
func (w YamlTextError) ToKind() Kind {
	t := w.To.Type()
	switch t.Kind() {
	case reflect.Array, reflect.Slice:
		return SequenceNode
	case reflect.Map, reflect.Struct:
		return MappingNode
	default:
		return ScalarNode
	}
}

func NewUnknownFieldError(err error, n Node, out reflect.Value, name string) error {
	return YamlError{
		Cause: YamlTextError{
			Node:  n,
			Name:  name,
			To:    out,
			Cause: CauseUnknownField,
		},
		Original: err,
	}
}

func NewAlreadyDefinedError(err error, n Node, out reflect.Value, name string, lineAt int) error {
	return YamlError{
		Cause: YamlTextError{
			Node:  n,
			Cause: CauseKeyAlreadyDefined,
			To:    out,
			Name:  name,
			Meta: map[string]string{
				"line_num": fmt.Sprintf("%d", lineAt),
			},
		},
		Original: err,
	}
}

func NewWrongTypeError(err error, n Node, out reflect.Value) error {
	return YamlError{
		Cause: YamlTextError{
			Node:  n,
			Cause: CauseWrongType,
			To:    out,
		},
		Original: err,
	}
}

func NewGoLangStructError(err error) error {
	return YamlError{
		Cause: GoLangStructError{
			Err: err,
		},
		Original: err,
	}
}

func underlyingPrimitive(i interface{}) string {
	t := reflect.TypeOf(i)
	switch t.Kind() {
	case reflect.Struct:
		return "key:value" // TODO: @emyrk idk about this one
	case reflect.Map:
		return "key:value"
	case reflect.Slice:
		return "[]value"
	case reflect.Ptr:
		// TODO: @emyrk Catch panic?
		return underlyingPrimitive(reflect.ValueOf(i).Elem().Interface())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return "uint"
	case reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return "float"
	case reflect.String:
		return "string"
	default:
		return "unknown"
	}
}
