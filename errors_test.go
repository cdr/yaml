package yaml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestErrorStrings(t *testing.T) {
	t.Run("UnknownKey", func(t *testing.T) {
		type s struct {
		}

		y := `
sequence-2:
 a: b
`
		var tmp s
		d := NewDecoder(bytes.NewBuffer([]byte(y)))
		d.KnownFields(true)
		err := d.Decode(&tmp)
		// Default:
		// yaml: unmarshal errors:
		// 	line 3: cannot unmarshal !!map into []string
		// 	line 5: cannot unmarshal !!map into []string
		fmt.Println(err)
	})

	t.Run("MappingOnSequence", func(t *testing.T) {
		type s struct {
			Sequence  []string `yaml:"sequence"`
			Sequence2 []string `yaml:"sequence-2"`
		}

		y := `
sequence-2:
 a: b
sequence:
 a: b

`
		var tmp s
		err := Unmarshal([]byte(y), &tmp)
		// Default:
		// yaml: unmarshal errors:
		// 	line 3: cannot unmarshal !!map into []string
		// 	line 5: cannot unmarshal !!map into []string
		fmt.Println(err)
	})

	t.Run("DuplicatedKeyInGoLangStruct", func(t *testing.T) {
		type s struct {
			A    int
			Nest struct {
				A int
			} `yaml:",inline"`
		}

		y := `
A: 7
`
		var tmp s
		err := Unmarshal([]byte(y), &tmp)
		// Default:
		// panic: duplicated key 'a' in struct yaml.s [recovered]
		// 		panic: duplicated key 'a' in struct yaml.s [recovered]
		// 		panic: duplicated key 'a' in struct yaml.s
		fmt.Println(err)
	})

	t.Run("KeyAlreadyDefined", func(t *testing.T) {
		type s map[string]interface{}

		y := `
a:
 b:
  c:
   d: "bar"
   d: "foo"`
		var tmp s
		err := Unmarshal([]byte(y), &tmp)
		// Default:
		// yaml: unmarshal errors:
		// 	line 3: mapping key "A" already defined at line 2
		fmt.Println(err)
	})

	t.Run("InvalidMapKey", func(t *testing.T) {
		type s map[string]string

		y := `
A: 3
B:
 C: "Hello"`
		var tmp s
		err := Unmarshal([]byte(y), &tmp)
		// Default:
		// yaml: unmarshal errors:
		// 	line 3: mapping key "A" already defined at line 2
		fmt.Println(err)
	})

	t.Run("MappingOnScalar", func(t *testing.T) {
		type s map[string]interface{}

		y := `
A: 3
 B: "Hello""`
		var tmp s
		err := Unmarshal([]byte(y), &tmp)
		// Default:
		// 	yaml: line 3: mapping values are not allowed in this context
		fmt.Println(err)
	})

	t.Run("InvalidMapKey", func(t *testing.T) {
		type s alphabet

		y := `
a:
  s: "Hello?"
b:
 p:
  s: "Goodbye"
  i: 2
 nest: 
#   i: "asd"
   slice: 
    - i: 1
    - i: 2
    - i: "ASD"
      b: "ASD"
    - s: "Hey"
      i: "Bad"
`
		var tmp s
		err := Unmarshal([]byte(y), &tmp)
		// Default:
		// 	<nil>
		data, _ := json.Marshal(tmp)
		var buf bytes.Buffer
		json.Indent(&buf, data, "", "\t")
		fmt.Println(buf.String() + "\n")
		fmt.Println(tmp, "\n", err)
	})

}

type alphabet struct {
	A prims
	B nest
}

type nest struct {
	P    prims `yaml:",inline"`
	Nest struct {
		P     prims `yaml:",inline"`
		Slice []prims
	} `yaml:"nest"`
}

type prims struct {
	S string  `json:",omitempty"`
	I int     `json:",omitempty"`
	B bool    `json:",omitempty"`
	F float64 `json:",omitempty"`
}
