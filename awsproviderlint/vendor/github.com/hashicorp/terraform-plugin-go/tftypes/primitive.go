package tftypes

import (
	"fmt"
	"math/big"
)

var (
	// DynamicPseudoType is a pseudo-type in Terraform's type system that
	// is used as a wildcard type. It indicates that any Terraform type can
	// be used.
	DynamicPseudoType = primitive{name: "DynamicPseudoType"}

	// String is a primitive type in Terraform that represents a UTF-8
	// string of bytes.
	String = primitive{name: "String"}

	// Number is a primitive type in Terraform that represents a real
	// number.
	Number = primitive{name: "Number"}

	// Bool is a primitive type in Terraform that represents a true or
	// false boolean value.
	Bool = primitive{name: "Bool"}
)

var (
	_ Type = primitive{name: "test"}
)

type primitive struct {
	name string

	// used to make this type uncomparable
	// see https://golang.org/ref/spec#Comparison_operators
	// this enforces the use of Is, instead
	_ []struct{}
}

func (p primitive) Equal(o primitive) bool {
	return p.equals(o, true)
}

func (p primitive) Is(t Type) bool {
	return p.equals(t, false)
}

func (p primitive) equals(t Type, exact bool) bool {
	v, ok := t.(primitive)
	if !ok {
		return false
	}
	return p.name == v.name
}

func (p primitive) String() string {
	return "tftypes." + string(p.name)
}

func (p primitive) private() {}

func (p primitive) MarshalJSON() ([]byte, error) {
	switch p.name {
	case String.name:
		return []byte(`"string"`), nil
	case Number.name:
		return []byte(`"number"`), nil
	case Bool.name:
		return []byte(`"bool"`), nil
	case DynamicPseudoType.name:
		return []byte(`"dynamic"`), nil
	}
	return nil, fmt.Errorf("unknown primitive type %q", p)
}

func (p primitive) supportedGoTypes() []string {
	switch p.name {
	case String.name:
		return []string{"string", "*string"}
	case Number.name:
		return []string{
			"*big.Float",
			"uint", "*uint",
			"uint8", "*uint8",
			"uint16", "*uint16",
			"uint32", "*uint32",
			"uint64", "*uint64",
			"int", "*int",
			"int8", "*int8",
			"int16", "*int16",
			"int32", "*int32",
			"int64", "*int64",
			"float64", "*float64",
		}
	case Bool.name:
		return []string{"bool", "*bool"}
	case DynamicPseudoType.name:
		possibleTypes := []Type{
			String, Bool, Number,
			Tuple{}, Object{},
		}
		results := []string{}
		for _, t := range possibleTypes {
			results = append(results, t.supportedGoTypes()...)
		}
		return results
	}
	panic(fmt.Sprintf("unknown primitive type %q", p.name))
}

func valueCanBeString(val interface{}) bool {
	switch val.(type) {
	case string:
		return true
	case *string:
		return true
	default:
		return false
	}
}

func valueFromString(in interface{}) (Value, error) {
	switch value := in.(type) {
	case *string:
		if value == nil {
			return Value{
				typ:   String,
				value: nil,
			}, nil
		}
		return Value{
			typ:   String,
			value: *value,
		}, nil
	case string:
		return Value{
			typ:   String,
			value: value,
		}, nil
	default:
		return Value{}, fmt.Errorf("tftypes.NewValue can't use %T as a tftypes.String; expected types are: %s", in, formattedSupportedGoTypes(String))
	}
}

func valueCanBeBool(val interface{}) bool {
	switch val.(type) {
	case bool:
		return true
	case *bool:
		return true
	default:
		return false
	}
}

func valueFromBool(in interface{}) (Value, error) {
	switch value := in.(type) {
	case *bool:
		if value == nil {
			return Value{
				typ:   Bool,
				value: nil,
			}, nil
		}
		return Value{
			typ:   Bool,
			value: *value,
		}, nil
	case bool:
		return Value{
			typ:   Bool,
			value: value,
		}, nil
	default:
		return Value{}, fmt.Errorf("tftypes.NewValue can't use %T as a tftypes.Bool; expected types are: %s", in, formattedSupportedGoTypes(Bool))
	}
}

func valueCanBeNumber(val interface{}) bool {
	switch val.(type) {
	case uint, uint8, uint16, uint32, uint64:
		return true
	case *uint, *uint8, *uint16, *uint32, *uint64:
		return true
	case int, int8, int16, int32, int64:
		return true
	case *int, *int8, *int16, *int32, *int64:
		return true
	case float64, *float64:
		return true
	case *big.Float:
		return true
	default:
		return false
	}
}

func valueFromNumber(in interface{}) (Value, error) {
	switch value := in.(type) {
	case *big.Float:
		return Value{
			typ:   Number,
			value: value,
		}, nil
	case uint:
		return Value{
			typ:   Number,
			value: new(big.Float).SetUint64(uint64(value)),
		}, nil
	case *uint:
		if value == nil {
			return Value{
				typ:   Number,
				value: nil,
			}, nil
		}
		return Value{
			typ:   Number,
			value: new(big.Float).SetUint64(uint64(*value)),
		}, nil
	case uint8:
		return Value{
			typ:   Number,
			value: new(big.Float).SetUint64(uint64(value)),
		}, nil
	case *uint8:
		if value == nil {
			return Value{
				typ:   Number,
				value: nil,
			}, nil
		}
		return Value{
			typ:   Number,
			value: new(big.Float).SetUint64(uint64(*value)),
		}, nil
	case uint16:
		return Value{
			typ:   Number,
			value: new(big.Float).SetUint64(uint64(value)),
		}, nil
	case *uint16:
		if value == nil {
			return Value{
				typ:   Number,
				value: nil,
			}, nil
		}
		return Value{
			typ:   Number,
			value: new(big.Float).SetUint64(uint64(*value)),
		}, nil
	case uint32:
		return Value{
			typ:   Number,
			value: new(big.Float).SetUint64(uint64(value)),
		}, nil
	case *uint32:
		if value == nil {
			return Value{
				typ:   Number,
				value: nil,
			}, nil
		}
		return Value{
			typ:   Number,
			value: new(big.Float).SetUint64(uint64(*value)),
		}, nil
	case uint64:
		return Value{
			typ:   Number,
			value: new(big.Float).SetUint64(value),
		}, nil
	case *uint64:
		if value == nil {
			return Value{
				typ:   Number,
				value: nil,
			}, nil
		}
		return Value{
			typ:   Number,
			value: new(big.Float).SetUint64(*value),
		}, nil
	case int:
		return Value{
			typ:   Number,
			value: new(big.Float).SetInt64(int64(value)),
		}, nil
	case *int:
		if value == nil {
			return Value{
				typ:   Number,
				value: nil,
			}, nil
		}
		return Value{
			typ:   Number,
			value: new(big.Float).SetInt64(int64(*value)),
		}, nil
	case int8:
		return Value{
			typ:   Number,
			value: new(big.Float).SetInt64(int64(value)),
		}, nil
	case *int8:
		if value == nil {
			return Value{
				typ:   Number,
				value: nil,
			}, nil
		}
		return Value{
			typ:   Number,
			value: new(big.Float).SetInt64(int64(*value)),
		}, nil
	case int16:
		return Value{
			typ:   Number,
			value: new(big.Float).SetInt64(int64(value)),
		}, nil
	case *int16:
		if value == nil {
			return Value{
				typ:   Number,
				value: nil,
			}, nil
		}
		return Value{
			typ:   Number,
			value: new(big.Float).SetInt64(int64(*value)),
		}, nil
	case int32:
		return Value{
			typ:   Number,
			value: new(big.Float).SetInt64(int64(value)),
		}, nil
	case *int32:
		if value == nil {
			return Value{
				typ:   Number,
				value: nil,
			}, nil
		}
		return Value{
			typ:   Number,
			value: new(big.Float).SetInt64(int64(*value)),
		}, nil
	case int64:
		return Value{
			typ:   Number,
			value: new(big.Float).SetInt64(value),
		}, nil
	case *int64:
		if value == nil {
			return Value{
				typ:   Number,
				value: nil,
			}, nil
		}
		return Value{
			typ:   Number,
			value: new(big.Float).SetInt64(*value),
		}, nil
	case float64:
		return Value{
			typ:   Number,
			value: big.NewFloat(value),
		}, nil
	case *float64:
		if value == nil {
			return Value{
				typ:   Number,
				value: nil,
			}, nil
		}
		return Value{
			typ:   Number,
			value: big.NewFloat(*value),
		}, nil
	default:
		return Value{}, fmt.Errorf("tftypes.NewValue can't use %T as a tftypes.Number; expected types are: %s", in, formattedSupportedGoTypes(Number))
	}
}

func valueFromDynamicPseudoType(val interface{}) (Value, error) {
	switch {
	case valueCanBeString(val):
		v, err := valueFromString(val)
		if err != nil {
			return Value{}, err
		}
		v.typ = DynamicPseudoType
		return v, nil
	case valueCanBeNumber(val):
		v, err := valueFromNumber(val)
		if err != nil {
			return Value{}, err
		}
		v.typ = DynamicPseudoType
		return v, nil
	case valueCanBeBool(val):
		v, err := valueFromBool(val)
		if err != nil {
			return Value{}, err
		}
		v.typ = DynamicPseudoType
		return v, nil
	case valueCanBeObject(val):
		v, err := valueFromObject(nil, nil, val)
		if err != nil {
			return Value{}, err
		}
		v.typ = DynamicPseudoType
		return v, nil
	case valueCanBeTuple(val):
		v, err := valueFromTuple(nil, val)
		if err != nil {
			return Value{}, err
		}
		v.typ = DynamicPseudoType
		return v, nil
	default:
		return Value{}, fmt.Errorf("tftypes.NewValue can't use %T as a tftypes.DynamicPseudoType; expected types are: %s", val, formattedSupportedGoTypes(DynamicPseudoType))
	}
}
