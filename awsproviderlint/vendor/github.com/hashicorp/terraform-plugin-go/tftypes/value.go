package tftypes

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"

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

func (val Value) String() string {
	typ := val.Type()

	// null and unknown values we use static strings for
	if val.IsNull() {
		return typ.String() + "<null>"
	}
	if !val.IsKnown() {
		return typ.String() + "<unknown>"
	}

	// everything else is built up
	var res strings.Builder
	switch {
	case typ.Is(String):
		var s string
		err := val.As(&s)
		if err != nil {
			panic(err)
		}
		res.WriteString(typ.String() + `<"` + s + `">`)
	case typ.Is(Number):
		n := big.NewFloat(0)
		err := val.As(&n)
		if err != nil {
			panic(err)
		}
		res.WriteString(typ.String() + `<"` + n.String() + `">`)
	case typ.Is(Bool):
		var b bool
		err := val.As(&b)
		if err != nil {
			panic(err)
		}
		res.WriteString(typ.String() + `<"` + strconv.FormatBool(b) + `">`)
	case typ.Is(List{}), typ.Is(Set{}), typ.Is(Tuple{}):
		var l []Value
		err := val.As(&l)
		if err != nil {
			panic(err)
		}
		res.WriteString(typ.String() + `<`)
		for pos, el := range l {
			if pos != 0 {
				res.WriteString(", ")
			}
			res.WriteString(el.String())
		}
		res.WriteString(">")
	case typ.Is(Map{}), typ.Is(Object{}):
		m := map[string]Value{}
		err := val.As(&m)
		if err != nil {
			panic(err)
		}
		res.WriteString(typ.String() + `<`)
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for pos, key := range keys {
			if pos != 0 {
				res.WriteString(", ")
			}
			res.WriteString(`"` + key + `":`)
			res.WriteString(m[key].String())
		}
		res.WriteString(">")
	}
	return res.String()
}

// ApplyTerraform5AttributePathStep applies an AttributePathStep to a Value,
// returning the Value found at that AttributePath within the Value. It
// fulfills that AttributePathStepper interface, allowing Values to be passed
// to WalkAttributePath. This allows retrieving a subset of a Value using an
// AttributePath. If the AttributePathStep can't be applied to the Value,
// either because it is the wrong type or because no Value exists at that
// AttributePathStep, an ErrInvalidStep error will be returned.
func (val Value) ApplyTerraform5AttributePathStep(step AttributePathStep) (interface{}, error) {
	if !val.IsKnown() || val.IsNull() {
		return nil, ErrInvalidStep
	}
	switch s := step.(type) {
	case AttributeName:
		if !val.Type().Is(Object{}) {
			return nil, ErrInvalidStep
		}
		o := map[string]Value{}
		err := val.As(&o)
		if err != nil {
			return nil, err
		}
		res, ok := o[string(s)]
		if !ok {
			return nil, ErrInvalidStep
		}
		return res, nil
	case ElementKeyString:
		if !val.Type().Is(Map{}) {
			return nil, ErrInvalidStep
		}
		m := map[string]Value{}
		err := val.As(&m)
		if err != nil {
			return nil, err
		}
		res, ok := m[string(s)]
		if !ok {
			return nil, ErrInvalidStep
		}
		return res, nil
	case ElementKeyInt:
		if !val.Type().Is(List{}) && !val.Type().Is(Tuple{}) {
			return nil, ErrInvalidStep
		}
		if int64(s) < 0 {
			return nil, ErrInvalidStep
		}
		sl := []Value{}
		err := val.As(&sl)
		if err != nil {
			return nil, err
		}
		if int64(len(sl)) <= int64(s) {
			return nil, ErrInvalidStep
		}
		return sl[int64(s)], nil
	case ElementKeyValue:
		if !val.Type().Is(Set{}) {
			return nil, ErrInvalidStep
		}
		sl := []Value{}
		err := val.As(&sl)
		if err != nil {
			return nil, err
		}
		for _, el := range sl {
			diffs, err := el.Diff(Value(s))
			if err != nil {
				return nil, err
			}
			if len(diffs) == 0 {
				return el, nil
			}
		}
		return nil, ErrInvalidStep
	default:
		return nil, fmt.Errorf("unexpected AttributePathStep type %T", step)
	}
}

// Equal returns true if two Values should be considered equal. Values are
// considered equal if their types are considered equal and if they represent
// data that is considered equal.
func (val Value) Equal(o Value) bool {
	if val.typ == nil && o.typ == nil && val.value == nil && o.value == nil {
		return true
	}
	if !val.Type().Is(o.Type()) {
		return false
	}
	diff, err := val.Diff(o)
	if err != nil {
		panic(err)
	}
	return len(diff) < 1
}

// Copy returns a defensively-copied clone of Value that shares no underlying
// data structures with the original Value and can be mutated without
// accidentally mutating the original.
func (val Value) Copy() Value {
	newVal := val.value
	switch v := val.value.(type) {
	case []Value:
		newVals := make([]Value, 0, len(v))
		for _, value := range v {
			newVals = append(newVals, value.Copy())
		}
		newVal = newVals
	case map[string]Value:
		newVals := make(map[string]Value, len(v))
		for k, value := range v {
			newVals[k] = value.Copy()
		}
		newVal = newVals
	}
	return NewValue(val.Type(), newVal)
}

// NewValue returns a Value constructed using the specified Type and stores the
// passed value in it.
//
// The passed value should be in one of the builtin Value representations or
// implement the ValueCreator interface.
//
// If the passed value is not a valid value for the passed type, NewValue will
// panic. Any value and type combination that does not return an error from
// ValidateValue is guaranteed to not panic. When calling NewValue with user
// input with a type not known at compile time, it is recommended to call
// ValidateValue before calling NewValue, to allow graceful handling of the
// error.
//
// The builtin Value representations are:
//
// * String: string, *string
//
// * Number: *big.Float, int64, *int64, int32, *int32, int16, *int16, int8,
//           *int8, int, *int, uint64, *uint64, uint32, *uint32, uint16,
//           *uint16, uint8, *uint8, uint, *uint, float64, *float64
//
// * Bool: bool, *bool
//
// * Map and Object: map[string]Value
//
// * Tuple, List, and Set: []Value
func NewValue(t Type, val interface{}) Value {
	v, err := newValue(t, val)
	if err != nil {
		panic(err)
	}
	return v
}

// ValidateValue checks that the Go type passed as `val` can be used as a value
// for the Type passed as `t`. A nil error response indicates that the value is
// valid for the type.
func ValidateValue(t Type, val interface{}) error {
	_, err := newValue(t, val)
	return err
}

func newValue(t Type, val interface{}) (Value, error) {
	if val == nil || val == UnknownValue {
		return Value{
			typ:   t,
			value: val,
		}, nil
	}
	if creator, ok := val.(ValueCreator); ok {
		var err error
		val, err = creator.ToTerraform5Value()
		if err != nil {
			return Value{}, fmt.Errorf("error creating tftypes.Value: %w", err)
		}
	}

	switch {
	case t.Is(DynamicPseudoType):
		v, err := valueFromDynamicPseudoType(val)
		if err != nil {
			return Value{}, err
		}
		return v, nil
	case t.Is(String):
		v, err := valueFromString(val)
		if err != nil {
			return Value{}, err
		}
		return v, nil
	case t.Is(Number):
		v, err := valueFromNumber(val)
		if err != nil {
			return Value{}, err
		}
		return v, nil
	case t.Is(Bool):
		v, err := valueFromBool(val)
		if err != nil {
			return Value{}, err
		}
		return v, nil
	case t.Is(Map{}):
		v, err := valueFromMap(t.(Map).AttributeType, val)
		if err != nil {
			return Value{}, err
		}
		return v, nil
	case t.Is(Object{}):
		v, err := valueFromObject(t.(Object).AttributeTypes, t.(Object).OptionalAttributes, val)
		if err != nil {
			return Value{}, err
		}
		return v, nil
	case t.Is(List{}):
		v, err := valueFromList(t.(List).ElementType, val)
		if err != nil {
			return Value{}, err
		}
		return v, nil
	case t.Is(Set{}):
		v, err := valueFromSet(t.(Set).ElementType, val)
		if err != nil {
			return Value{}, err
		}
		return v, nil
	case t.Is(Tuple{}):
		v, err := valueFromTuple(t.(Tuple).ElementTypes, val)
		if err != nil {
			return Value{}, err
		}
		return v, nil
	default:
		return Value{}, fmt.Errorf("unknown type %s passed to tftypes.NewValue", t)
	}
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
			return fmt.Errorf("can't unmarshal %s into %T, expected string", val.Type(), dst)
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
			return fmt.Errorf("can't unmarshal %s into %T, expected *big.Float", val.Type(), dst)
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
			return fmt.Errorf("can't unmarshal %s into %T, expected boolean", val.Type(), dst)
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
			return fmt.Errorf("can't unmarshal %s into %T, expected map[string]tftypes.Value", val.Type(), dst)
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
			return fmt.Errorf("can't unmarshal %s into %T expected []tftypes.Value", val.Type(), dst)
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

// Type returns the Type of the Value.
func (val Value) Type() Type {
	return val.typ
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
	if val.value == nil {
		return true
	}
	switch val.Type().(type) {
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
	panic(fmt.Sprintf("unknown type %T", val.Type()))
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

	err := marshalMsgPack(val, t, NewAttributePath(), enc)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func unexpectedValueTypeError(p *AttributePath, expected, got interface{}, typ Type) error {
	return p.NewErrorf("unexpected value type %T, %s values must be of type %T", got, typ, expected)
}
