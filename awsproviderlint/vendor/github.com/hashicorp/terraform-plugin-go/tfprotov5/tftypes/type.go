package tftypes

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Type is an interface representing a Terraform type. It is only meant to be
// implemented by the tftypes package. Types define the shape and
// characteristics of data coming from or being sent to Terraform.
type Type interface {
	// Is is used to determine what type a Type implementation is. It is
	// the recommended method for determining whether two types are
	// equivalent or not.
	Is(Type) bool

	// String returns a string representation of the Type's name.
	String() string

	// MarshalJSON returns a JSON representation of the Type's signature.
	// It is modeled based on Terraform's requirements for type signature
	// JSON representations, and may change over time to match Terraform's
	// formatting.
	//
	// Deprecated: this is not meant to be called by third-party code.
	MarshalJSON() ([]byte, error)

	// private is meant to keep this interface from being implemented by
	// types from other packages.
	private()
}

type jsonType struct {
	t Type
}

// ParseJSONType returns a Type from its JSON representation. The JSON
// representation should come from Terraform or from MarshalJSON as the format
// is not part of this package's API guarantees.
//
// Deprecated: this is not meant to be called by third-party code.
func ParseJSONType(buf []byte) (Type, error) {
	var t jsonType
	err := json.Unmarshal(buf, &t)
	return t.t, err
}

func (t *jsonType) UnmarshalJSON(buf []byte) error {
	r := bytes.NewReader(buf)
	dec := json.NewDecoder(r)

	tok, err := dec.Token()
	if err != nil {
		return err
	}

	switch v := tok.(type) {
	case string:
		switch v {
		case "bool":
			t.t = Bool
		case "number":
			t.t = Number
		case "string":
			t.t = String
		case "dynamic":
			t.t = DynamicPseudoType
		default:
			return fmt.Errorf("invalid primitive type name %q", v)
		}

		if dec.More() {
			return fmt.Errorf("extraneous data after type description")
		}
		return nil
	case json.Delim:
		if rune(v) != '[' {
			return fmt.Errorf("invalid complex type description")
		}

		tok, err = dec.Token()
		if err != nil {
			return err
		}

		kind, ok := tok.(string)
		if !ok {
			return fmt.Errorf("invalid complex type kind name")
		}

		switch kind {
		case "list":
			var ety jsonType
			err = dec.Decode(&ety)
			if err != nil {
				return err
			}
			t.t = List{
				ElementType: ety.t,
			}
		case "map":
			var ety jsonType
			err = dec.Decode(&ety)
			if err != nil {
				return err
			}
			t.t = Map{
				AttributeType: ety.t,
			}
		case "set":
			var ety jsonType
			err = dec.Decode(&ety)
			if err != nil {
				return err
			}
			t.t = Set{
				ElementType: ety.t,
			}
		case "object":
			var atys map[string]jsonType
			err = dec.Decode(&atys)
			if err != nil {
				return err
			}
			types := make(map[string]Type, len(atys))
			for k, v := range atys {
				types[k] = v.t
			}
			t.t = Object{
				AttributeTypes: types,
			}
		case "tuple":
			var etys []jsonType
			err = dec.Decode(&etys)
			if err != nil {
				return err
			}
			types := make([]Type, 0, len(etys))
			for _, ty := range etys {
				types = append(types, ty.t)
			}
			t.t = Tuple{
				ElementTypes: types,
			}
		default:
			return fmt.Errorf("invalid complex type kind name")
		}

		tok, err = dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := tok.(json.Delim); !ok || rune(delim) != ']' || dec.More() {
			return fmt.Errorf("unexpected extra data in type description")
		}

		return nil

	default:
		return fmt.Errorf("invalid type description")
	}
}
