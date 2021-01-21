package tftypes

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/google/go-cmp/cmp"
	"github.com/vmihailenco/msgpack"
)

// ValueConverter is an interface that provider-defined types can implement to
// control how Value.As will convert a Value into that type. The passed Value
// is the Value that Value.As is being called on. The intended usage is to call
// Value.As on the passed Value, converting it into a builtin type, and then
// converting or casting that builtin type to the provider-defined type.
type ValueConverter interface {
	FromTerraform5Value(Value) error
}

// ValueCreator is an interface that provider-defined types can implement to
// control how NewValue will convert that type into a Value. The returned
// interface should return one of the builtin Value representations that should
// be used for that Value.
type ValueCreator interface {
	ToTerraform5Value() (interface{}, error)
}

type msgPackUnknownType struct{}

var msgPackUnknownVal = msgPackUnknownType{}

func (u msgPackUnknownType) MarshalMsgpack() ([]byte, error) {
	return []byte{0xd4, 0, 0}, nil
}

// Value is a piece of data from Terraform or being returned to Terraform. It
// has a Type associated with it, defining its shape and characteristics, and a
// Go representation of that Type containing the data itself. Values are a
// special type and are not represented as pure Go values beause they can
// contain UnknownValues, which cannot be losslessly represented in Go's type
// system.
//
// The recommended usage of a Value is to check that it is known, using
// Value.IsKnown, then to convert it to a Go type, using Value.As. The Go type
// can then be manipulated.
type Value struct {
	typ   Type
	value interface{}
}

// NewValue returns a Value constructed using the specified Type and stores the
// passed value in it.
//
// The passed value should be in one of the builtin Value representations or
// implement the ValueCreator interface.
//
// The builtin Value representations are:
//
// * string for String
//
// * *big.Float for Number
//
// * bool for Bool
//
// * map[string]Value for Map and Object
//
// * []Value for Tuple, List, and Set
func NewValue(t Type, val interface{}) Value {
	if val == nil || val == UnknownValue {
		return Value{
			typ:   t,
			value: val,
		}
	}
	if creator, ok := val.(ValueCreator); ok {
		var err error
		val, err = creator.ToTerraform5Value()
		if err != nil {
			panic("error creating tftypes.Value: " + err.Error())
		}
	}
	switch val.(type) {
	case string, *big.Float, bool, map[string]Value, []Value:
		return Value{
			typ:   t,
			value: val,
		}
	}
	panic(fmt.Sprintf("unknown type %T passed to NewValue", val))
}

// As converts a Value into a Go value. `dst` must be set to a pointer to a
// value of a supported type for the Value's type or an implementation of the
// ValueConverter interface.
//
// For Strings, `dst` must be a pointer to a string or a pointer to a pointer
// to a string. If it's a pointer to a pointer to a string, if the Value is
// null, the pointer to the string will be set to nil. If it's a pointer to a
// string, if the Value is null, the string will be set to the empty value.
//
// For Numbers, `dst` must be a poitner to a big.Float or a pointer to a
// pointer to a big.Float. If it's a pointer to a pointer to a big.Float, if
// the Value is null, the pointer to the big.Float will be set to nil. If it's
// a pointer to a big.Float, if the Value is null, the big.Float will be set to
// 0.
//
// For Bools, `dst` must be a pointer to a bool or a pointer to a pointer to a
// bool. If it's a pointer to a pointer to a bool, if the Value is null, the
// pointer to the bool will be set to nil. If it's a pointer to a bool, if the
// Value is null, the bool will be set to false.
//
// For Maps and Objects, `dst` must be a pointer to a map[string]Value or a
// pointer to a pointer to a map[string]Value. If it's a pointer to a pointer
// to a map[string]Value, if the Value is null, the pointer to the
// map[string]Value will be set to nil. If it's a pointer to a
// map[string]Value, if the Value is null, the map[string]Value will be set to
// an empty map.
//
// For Lists, Sets, and Tuples, `dst` must be a pointer to a []Value or a
// pointer to a pointer to a []Value. If it's a pointer to a pointer to a
// []Value, if the Value is null, the poitner to []Value will be set to nil. If
// it's a pointer to a []Value, if the Value is null, the []Value will be set
// to an empty slice.
//
// Future builtin conversions may be added over time.
//
// If `val` is unknown, an error will be returned, as unknown values can't be
// represented in Go's type system. Providers should check Value.IsKnown before
// calling Value.As.
func (val Value) As(dst interface{}) error {
	unmarshaler, ok := dst.(ValueConverter)
	if ok {
		return unmarshaler.FromTerraform5Value(val)
	}
	if !val.IsKnown() {
		return fmt.Errorf("unmarshaling unknown values is not supported")
	}
	switch target := dst.(type) {
	case *string:
		if val.IsNull() {
			*target = ""
			return nil
		}
		v, ok := val.value.(string)
		if !ok {
			return fmt.Errorf("can't unmarshal %s into %T, expected string", val.typ, dst)
		}
		*target = v
		return nil
	case **string:
		if val.IsNull() {
			*target = nil
			return nil
		}
		return val.As(*target)
	case *big.Float:
		if val.IsNull() {
			target.Set(big.NewFloat(0))
			return nil
		}
		v, ok := val.value.(*big.Float)
		if !ok {
			return fmt.Errorf("can't unmarshal %s into %T, expected *big.Float", val.typ, dst)
		}
		target.Set(v)
		return nil
	case **big.Float:
		if val.IsNull() {
			*target = nil
			return nil
		}
		return val.As(*target)
	case *bool:
		if val.IsNull() {
			*target = false
			return nil
		}
		v, ok := val.value.(bool)
		if !ok {
			return fmt.Errorf("can't unmarshal %s into %T, expected boolean", val.typ, dst)
		}
		*target = v
		return nil
	case **bool:
		if val.IsNull() {
			*target = nil
			return nil
		}
		return val.As(*target)
	case *map[string]Value:
		if val.IsNull() {
			*target = map[string]Value{}
			return nil
		}
		v, ok := val.value.(map[string]Value)
		if !ok {
			return fmt.Errorf("can't unmarshal %s into %T, expected map[string]tftypes.Value", val.typ, dst)
		}
		*target = v
		return nil
	case **map[string]Value:
		if val.IsNull() {
			*target = nil
			return nil
		}
		return val.As(*target)
	case *[]Value:
		if val.IsNull() {
			*target = []Value{}
			return nil
		}
		v, ok := val.value.([]Value)
		if !ok {
			return fmt.Errorf("can't unmarshal %s into %T expected []tftypes.Value", val.typ, dst)
		}
		*target = v
		return nil
	case **[]Value:
		if val.IsNull() {
			*target = nil
			return nil
		}
		return val.As(*target)
	}
	return fmt.Errorf("can't unmarshal into %T, needs FromTerraform5Value method", dst)
}

// Is calls Type.Is using the Type of the Value.
func (val Value) Is(t Type) bool {
	if val.typ == nil || t == nil {
		return val.typ == nil && t == nil
	}
	return val.typ.Is(t)
}

// IsKnown returns true if `val` is known. If `val` is an aggregate type, only
// the top level of the aggregate type is checked; elements and attributes are
// not checked.
func (val Value) IsKnown() bool {
	return val.value != UnknownValue
}

// IsFullyKnown returns true if `val` is known. If `val` is an aggregate type,
// IsFullyKnown only returns true if all elements and attributes are known, as
// well.
func (val Value) IsFullyKnown() bool {
	if !val.IsKnown() {
		return false
	}
	switch val.typ.(type) {
	case primitive:
		return true
	case List, Set, Tuple:
		for _, v := range val.value.([]Value) {
			if !v.IsFullyKnown() {
				return false
			}
		}
		return true
	case Map, Object:
		for _, v := range val.value.(map[string]Value) {
			if !v.IsFullyKnown() {
				return false
			}
		}
		return true
	}
	panic(fmt.Sprintf("unknown type %T", val.typ))
}

// IsNull returns true if the Value is null.
func (val Value) IsNull() bool {
	return val.value == nil
}

// MarshalMsgPack returns a msgpack representation of the Value. This is used
// for constructing tfprotov5.DynamicValues.
//
// Deprecated: this is not meant to be called by third parties. Don't use it.
func (val Value) MarshalMsgPack(t Type) ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)

	err := marshalMsgPack(val, t, AttributePath{}, enc)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func unexpectedValueTypeError(p AttributePath, expected, got interface{}, typ Type) error {
	return p.NewErrorf("unexpected value type %T, %s values must be of type %T", got, typ, expected)
}

// ValueComparer returns a github.com/google/go-cmp/cmp#Option that can be used
// to tell go-cmp how to compare Values.
func ValueComparer() cmp.Option {
	return cmp.Comparer(valueComparer)
}

func numberComparer(i, j *big.Float) bool {
	return (i == nil && j == nil) || (i != nil && j != nil && i.Cmp(j) == 0)
}

func valueComparer(i, j Value) bool {
	if !i.typ.Is(j.typ) {
		return false
	}
	return cmp.Equal(i.value, j.value, cmp.Comparer(numberComparer), ValueComparer())
}

// TypeFromElements returns the common type that the passed elements all have
// in common. An error will be returned if the passed elements are not of the
// same type.
func TypeFromElements(elements []Value) (Type, error) {
	var typ Type
	for _, el := range elements {
		if typ == nil {
			typ = el.typ
			continue
		}
		if !typ.Is(el.typ) {
			return nil, errors.New("elements do not all have the same types")
		}
	}
	if typ == nil {
		return DynamicPseudoType, nil
	}
	return typ, nil
}
