package tftypes

import "fmt"

const (
	// DynamicPseudoType is a pseudo-type in Terraform's type system that
	// is used as a wildcard type. It indicates that any Terraform type can
	// be used.
	DynamicPseudoType = primitive("DynamicPseudoType")

	// String is a primitive type in Terraform that represents a UTF-8
	// string of bytes.
	String = primitive("String")

	// Number is a primitive type in Terraform that represents a real
	// number.
	Number = primitive("Number")

	// Bool is a primitive type in Terraform that represents a true or
	// false boolean value.
	Bool = primitive("Bool")
)

var (
	_ Type = primitive("test")
)

type primitive string

func (p primitive) Is(t Type) bool {
	v, ok := t.(primitive)
	if !ok {
		return false
	}
	return p == v
}

func (p primitive) String() string {
	return "tftypes." + string(p)
}

func (p primitive) private() {}

func (p primitive) MarshalJSON() ([]byte, error) {
	switch p {
	case String:
		return []byte(`"string"`), nil
	case Number:
		return []byte(`"number"`), nil
	case Bool:
		return []byte(`"bool"`), nil
	case DynamicPseudoType:
		return []byte(`"dynamic"`), nil
	}
	return nil, fmt.Errorf("unknown primitive type %q", p)
}
