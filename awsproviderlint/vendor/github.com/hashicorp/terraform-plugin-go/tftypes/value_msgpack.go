package tftypes

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/vmihailenco/msgpack"
	msgpackCodes "github.com/vmihailenco/msgpack/codes"
)

type msgPackUnknownType struct{}

var msgPackUnknownVal = msgPackUnknownType{}

func (u msgPackUnknownType) MarshalMsgpack() ([]byte, error) {
	return []byte{0xd4, 0, 0}, nil
}

// ValueFromMsgPack returns a Value from the MsgPack-encoded bytes, using the
// provided Type to determine what shape the Value should be.
// DynamicPseudoTypes will be transparently parsed into the types they
// represent.
//
// Deprecated: this function is exported for internal use in
// terraform-plugin-go.  Third parties should not use it, and its behavior is
// not covered under the API compatibility guarantees. Don't use this.
func ValueFromMsgPack(data []byte, typ Type) (Value, error) {
	r := bytes.NewReader(data)
	dec := msgpack.NewDecoder(r)
	return msgpackUnmarshal(dec, typ, NewAttributePath())
}

func msgpackUnmarshal(dec *msgpack.Decoder, typ Type, path *AttributePath) (Value, error) {
	peek, err := dec.PeekCode()
	if err != nil {
		return Value{}, path.NewErrorf("error peeking next byte: %w", err)
	}
	if msgpackCodes.IsExt(peek) {
		// as with go-cty, assume all extensions are unknown values
		err := dec.Skip()
		if err != nil {
			return Value{}, path.NewErrorf("error skipping extension byte: %w", err)
		}
		return NewValue(typ, UnknownValue), nil
	}
	if typ.Is(DynamicPseudoType) {
		return msgpackUnmarshalDynamic(dec, path)
	}
	if peek == msgpackCodes.Nil {
		err := dec.Skip()
		if err != nil {
			return Value{}, path.NewErrorf("error skipping nil byte: %w", err)
		}
		return NewValue(typ, nil), nil
	}

	switch {
	case typ.Is(String):
		rv, err := dec.DecodeString()
		if err != nil {
			return Value{}, path.NewErrorf("error decoding string: %w", err)
		}
		return NewValue(String, rv), nil
	case typ.Is(Number):
		peek, err := dec.PeekCode()
		if err != nil {
			return Value{}, path.NewErrorf("couldn't peek number: %w", err)
		}
		if msgpackCodes.IsFixedNum(peek) {
			rv, err := dec.DecodeInt64()
			if err != nil {
				return Value{}, path.NewErrorf("couldn't decode number as int64: %w", err)
			}
			return NewValue(Number, big.NewFloat(float64(rv))), nil
		}
		switch peek {
		case msgpackCodes.Int8, msgpackCodes.Int16, msgpackCodes.Int32, msgpackCodes.Int64:
			rv, err := dec.DecodeInt64()
			if err != nil {
				return Value{}, path.NewErrorf("couldn't decode number as int64: %w", err)
			}
			return NewValue(Number, big.NewFloat(float64(rv))), nil
		case msgpackCodes.Uint8, msgpackCodes.Uint16, msgpackCodes.Uint32, msgpackCodes.Uint64:
			rv, err := dec.DecodeUint64()
			if err != nil {
				return Value{}, path.NewErrorf("couldn't decode number as uint64: %w", err)
			}
			return NewValue(Number, big.NewFloat(float64(rv))), nil
		case msgpackCodes.Float, msgpackCodes.Double:
			rv, err := dec.DecodeFloat64()
			if err != nil {
				return Value{}, path.NewErrorf("couldn't decode number as float64: %w", err)
			}
			return NewValue(Number, big.NewFloat(float64(rv))), nil
		default:
			rv, err := dec.DecodeString()
			if err != nil {
				return Value{}, path.NewErrorf("couldn't decode number as string: %w", err)
			}
			// according to
			// https://github.com/hashicorp/go-cty/blob/85980079f637862fa8e43ddc82dd74315e2f4c85/cty/value_init.go#L49
			// Base 10, precision 512, and rounding to nearest even
			// is the standard way to handle numbers arriving as
			// strings.
			fv, _, err := big.ParseFloat(rv, 10, 512, big.ToNearestEven)
			if err != nil {
				return Value{}, path.NewErrorf("error parsing %q as number: %w", rv, err)
			}
			return NewValue(Number, fv), nil
		}
	case typ.Is(Bool):
		rv, err := dec.DecodeBool()
		if err != nil {
			return Value{}, path.NewErrorf("couldn't decode bool: %w", err)
		}
		return NewValue(Bool, rv), nil
	case typ.Is(List{}):
		return msgpackUnmarshalList(dec, typ.(List).ElementType, path)
	case typ.Is(Set{}):
		return msgpackUnmarshalSet(dec, typ.(Set).ElementType, path)
	case typ.Is(Map{}):
		return msgpackUnmarshalMap(dec, typ.(Map).AttributeType, path)
	case typ.Is(Tuple{}):
		return msgpackUnmarshalTuple(dec, typ.(Tuple).ElementTypes, path)
	case typ.Is(Object{}):
		return msgpackUnmarshalObject(dec, typ.(Object).AttributeTypes, path)
	}
	return Value{}, path.NewErrorf("unsupported type %s", typ.String())
}

func msgpackUnmarshalList(dec *msgpack.Decoder, typ Type, path *AttributePath) (Value, error) {
	length, err := dec.DecodeArrayLen()
	if err != nil {
		return Value{}, path.NewErrorf("error decoding list length: %w", err)
	}

	switch {
	case length < 0:
		return NewValue(List{
			ElementType: typ,
		}, nil), nil
	case length == 0:
		return NewValue(List{
			ElementType: typ,
		}, []Value{}), nil
	}

	vals := make([]Value, 0, length)
	for i := 0; i < length; i++ {
		innerPath := path.WithElementKeyInt(int64(i))
		val, err := msgpackUnmarshal(dec, typ, innerPath)
		if err != nil {
			return Value{}, err
		}
		vals = append(vals, val)
	}

	elTyp := typ
	if elTyp.Is(DynamicPseudoType) {
		elTyp, err = TypeFromElements(vals)
		if err != nil {
			return Value{}, err
		}
	}

	return NewValue(List{
		ElementType: elTyp,
	}, vals), nil
}

func msgpackUnmarshalSet(dec *msgpack.Decoder, typ Type, path *AttributePath) (Value, error) {
	length, err := dec.DecodeArrayLen()
	if err != nil {
		return Value{}, path.NewErrorf("error decoding set length: %w", err)
	}

	switch {
	case length < 0:
		return NewValue(Set{
			ElementType: typ,
		}, nil), nil
	case length == 0:
		return NewValue(Set{
			ElementType: typ,
		}, []Value{}), nil
	}

	vals := make([]Value, 0, length)
	for i := 0; i < length; i++ {
		innerPath := path.WithElementKeyInt(int64(i))
		val, err := msgpackUnmarshal(dec, typ, innerPath)
		if err != nil {
			return Value{}, err
		}
		vals = append(vals, val)
	}

	elTyp, err := TypeFromElements(vals)
	if err != nil {
		return Value{}, err
	}

	return NewValue(Set{
		ElementType: elTyp,
	}, vals), nil
}

func msgpackUnmarshalMap(dec *msgpack.Decoder, typ Type, path *AttributePath) (Value, error) {
	length, err := dec.DecodeMapLen()
	if err != nil {
		return Value{}, path.NewErrorf("error decoding map length: %w", err)
	}

	switch {
	case length < 0:
		return NewValue(Map{
			AttributeType: typ,
		}, nil), nil
	case length == 0:
		return NewValue(Map{
			AttributeType: typ,
		}, map[string]Value{}), nil
	}

	vals := make(map[string]Value, length)
	for i := 0; i < length; i++ {
		key, err := dec.DecodeString()
		if err != nil {
			return Value{}, path.NewErrorf("error decoding map key: %w", err)
		}
		innerPath := path.WithElementKeyString(key)
		val, err := msgpackUnmarshal(dec, typ, innerPath)
		if err != nil {
			return Value{}, err
		}
		vals[key] = val
	}
	return NewValue(Map{
		AttributeType: typ,
	}, vals), nil
}

func msgpackUnmarshalTuple(dec *msgpack.Decoder, types []Type, path *AttributePath) (Value, error) {
	length, err := dec.DecodeArrayLen()
	if err != nil {
		return Value{}, path.NewErrorf("error decoding tuple length: %w", err)
	}

	switch {
	case length < 0:
		return NewValue(Tuple{
			ElementTypes: types,
		}, nil), nil
	case length == 0:
		return NewValue(Tuple{
			// no elements means no types
			ElementTypes: nil,
		}, []Value{}), nil
	case length != len(types):
		return Value{}, path.NewErrorf("error decoding tuple; expected %d items, got %d", len(types), length)
	}

	vals := make([]Value, 0, length)
	for i := 0; i < length; i++ {
		innerPath := path.WithElementKeyInt(int64(i))
		typ := types[i]
		val, err := msgpackUnmarshal(dec, typ, innerPath)
		if err != nil {
			return Value{}, err
		}
		vals = append(vals, val)
	}

	return NewValue(Tuple{
		ElementTypes: types,
	}, vals), nil
}

func msgpackUnmarshalObject(dec *msgpack.Decoder, types map[string]Type, path *AttributePath) (Value, error) {
	length, err := dec.DecodeMapLen()
	if err != nil {
		return Value{}, path.NewErrorf("error decoding object length: %w", err)
	}

	switch {
	case length < 0:
		return NewValue(Object{
			AttributeTypes: types,
		}, nil), nil
	case length == 0:
		return NewValue(Object{
			// no attributes means no types
			AttributeTypes: map[string]Type{},
		}, map[string]Value{}), nil
	case length != len(types):
		return Value{}, path.NewErrorf("error decoding object; expected %d attributes, got %d", len(types), length)
	}

	vals := make(map[string]Value, length)
	for i := 0; i < length; i++ {
		key, err := dec.DecodeString()
		if err != nil {
			return Value{}, path.NewErrorf("error decoding object key: %w", err)
		}
		typ, exists := types[key]
		if !exists {
			return Value{}, path.NewErrorf("unknown attribute %q", key)
		}
		innerPath := path.WithAttributeName(key)
		val, err := msgpackUnmarshal(dec, typ, innerPath)
		if err != nil {
			return Value{}, err
		}
		vals[key] = val
	}

	return NewValue(Object{
		AttributeTypes: types,
	}, vals), nil
}

func msgpackUnmarshalDynamic(dec *msgpack.Decoder, path *AttributePath) (Value, error) {
	length, err := dec.DecodeArrayLen()
	if err != nil {
		return Value{}, path.NewErrorf("error checking length of DynamicPseudoType value: %w", err)
	}

	switch {
	case length == -1:
		return NewValue(DynamicPseudoType, nil), nil
	case length != 2:
		return Value{}, path.NewErrorf("expected %d elements in DynamicPseudoType array, got %d", 2, length)
	}

	typeJSON, err := dec.DecodeBytes()
	if err != nil {
		return Value{}, path.NewErrorf("error decoding bytes: %w", err)
	}
	typ, err := ParseJSONType(typeJSON) //nolint:staticcheck
	if err != nil {
		return Value{}, path.NewErrorf("error parsing type information: %w", err)
	}
	return msgpackUnmarshal(dec, typ, path)
}

func marshalMsgPack(val Value, typ Type, p *AttributePath, enc *msgpack.Encoder) error {
	if typ.Is(DynamicPseudoType) && !val.Type().Is(DynamicPseudoType) {
		return marshalMsgPackDynamicPseudoType(val, typ, p, enc)

	}
	if !val.IsKnown() {
		err := enc.Encode(msgPackUnknownVal)
		if err != nil {
			return p.NewErrorf("error encoding UnknownValue: %w", err)
		}
		return nil
	}
	if val.IsNull() {
		err := enc.EncodeNil()
		if err != nil {
			return p.NewErrorf("error encoding null value: %w", err)
		}
		return nil
	}
	switch {
	case typ.Is(String):
		return marshalMsgPackString(val, typ, p, enc)
	case typ.Is(Number):
		return marshalMsgPackNumber(val, typ, p, enc)
	case typ.Is(Bool):
		return marshalMsgPackBool(val, typ, p, enc)
	case typ.Is(List{}):
		return marshalMsgPackList(val, typ, p, enc)
	case typ.Is(Set{}):
		return marshalMsgPackSet(val, typ, p, enc)
	case typ.Is(Map{}):
		return marshalMsgPackMap(val, typ, p, enc)
	case typ.Is(Tuple{}):
		return marshalMsgPackTuple(val, typ, p, enc)
	case typ.Is(Object{}):
		return marshalMsgPackObject(val, typ, p, enc)
	}
	return fmt.Errorf("unknown type %s", typ)
}

func marshalMsgPackDynamicPseudoType(val Value, typ Type, p *AttributePath, enc *msgpack.Encoder) error {
	typeJSON, err := val.Type().MarshalJSON()
	if err != nil {
		return p.NewErrorf("error generating JSON for type %s: %w", val.Type(), err)
	}
	err = enc.EncodeArrayLen(2)
	if err != nil {
		return p.NewErrorf("error encoding array length:  %w", err)
	}
	err = enc.EncodeBytes(typeJSON)
	if err != nil {
		return p.NewErrorf("error encoding JSON type info: %w", err)
	}
	err = marshalMsgPack(val, val.Type(), p, enc)
	if err != nil {
		return p.NewErrorf("error marshaling DynamicPseudoType value: %w", err)
	}
	return nil
}

func marshalMsgPackString(val Value, typ Type, p *AttributePath, enc *msgpack.Encoder) error {
	s, ok := val.value.(string)
	if !ok {
		return unexpectedValueTypeError(p, s, val.value, typ)
	}
	err := enc.EncodeString(s)
	if err != nil {
		return p.NewErrorf("error encoding string value: %w", err)
	}
	return nil
}

func marshalMsgPackNumber(val Value, typ Type, p *AttributePath, enc *msgpack.Encoder) error {
	n, ok := val.value.(*big.Float)
	if !ok {
		return unexpectedValueTypeError(p, n, val.value, typ)
	}
	if n.IsInf() {
		if n.Sign() == -1 {
			err := enc.EncodeFloat64(math.Inf(-1))
			if err != nil {
				return p.NewErrorf("error encoding negative infinity: %w", err)
			}
		} else if n.Sign() == 1 {
			err := enc.EncodeFloat64(math.Inf(1))
			if err != nil {
				return p.NewErrorf("error encoding positive infinity: %w", err)
			}
		} else {
			return p.NewErrorf("error encoding unknown infiniy: sign %d is unknown", n.Sign())
		}
	} else if iv, acc := n.Int64(); acc == big.Exact {
		err := enc.EncodeInt(iv)
		if err != nil {
			return p.NewErrorf("error encoding int value: %w", err)
		}
	} else if fv, acc := n.Float64(); acc == big.Exact {
		err := enc.EncodeFloat64(fv)
		if err != nil {
			return p.NewErrorf("error encoding float value: %w", err)
		}
	} else {
		err := enc.EncodeString(n.Text('f', -1))
		if err != nil {
			return p.NewErrorf("error encoding number string value: %w", err)
		}
	}
	return nil
}

func marshalMsgPackBool(val Value, typ Type, p *AttributePath, enc *msgpack.Encoder) error {
	b, ok := val.value.(bool)
	if !ok {
		return unexpectedValueTypeError(p, b, val.value, typ)
	}
	err := enc.EncodeBool(b)
	if err != nil {
		return p.NewErrorf("error encoding bool value: %w", err)
	}
	return nil
}

func marshalMsgPackList(val Value, typ Type, p *AttributePath, enc *msgpack.Encoder) error {
	l, ok := val.value.([]Value)
	if !ok {
		return unexpectedValueTypeError(p, l, val.value, typ)
	}
	err := enc.EncodeArrayLen(len(l))
	if err != nil {
		return p.NewErrorf("error encoding list length: %w", err)
	}
	for pos, i := range l {
		err := marshalMsgPack(i, typ.(List).ElementType, p.WithElementKeyInt(int64(pos)), enc)
		if err != nil {
			return err
		}
	}
	return nil
}

func marshalMsgPackSet(val Value, typ Type, p *AttributePath, enc *msgpack.Encoder) error {
	s, ok := val.value.([]Value)
	if !ok {
		return unexpectedValueTypeError(p, s, val.value, typ)
	}
	err := enc.EncodeArrayLen(len(s))
	if err != nil {
		return p.NewErrorf("error encoding set length: %w", err)
	}
	for _, i := range s {
		err := marshalMsgPack(i, typ.(Set).ElementType, p.WithElementKeyValue(i), enc)
		if err != nil {
			return err
		}
	}
	return nil
}

func marshalMsgPackMap(val Value, typ Type, p *AttributePath, enc *msgpack.Encoder) error {
	m, ok := val.value.(map[string]Value)
	if !ok {
		return unexpectedValueTypeError(p, m, val.value, typ)
	}
	err := enc.EncodeMapLen(len(m))
	if err != nil {
		return p.NewErrorf("error encoding map length: %w", err)
	}
	for k, v := range m {
		err := marshalMsgPack(NewValue(String, k), String, p.WithElementKeyString(k), enc)
		if err != nil {
			return p.NewErrorf("error encoding map key: %w", err)
		}
		err = marshalMsgPack(v, typ.(Map).AttributeType, p, enc)
		if err != nil {
			return err
		}
	}
	return nil
}

func marshalMsgPackTuple(val Value, typ Type, p *AttributePath, enc *msgpack.Encoder) error {
	t, ok := val.value.([]Value)
	if !ok {
		return unexpectedValueTypeError(p, t, val.value, typ)
	}
	types := typ.(Tuple).ElementTypes
	err := enc.EncodeArrayLen(len(types))
	if err != nil {
		return p.NewErrorf("error encoding tuple length: %w", err)
	}
	for pos, v := range t {
		ty := types[pos]
		err := marshalMsgPack(v, ty, p.WithElementKeyInt(int64(pos)), enc)
		if err != nil {
			return err
		}
	}
	return nil
}

func marshalMsgPackObject(val Value, typ Type, p *AttributePath, enc *msgpack.Encoder) error {
	o, ok := val.value.(map[string]Value)
	if !ok {
		return unexpectedValueTypeError(p, o, val.value, typ)
	}
	types := typ.(Object).AttributeTypes
	keys := make([]string, 0, len(types))
	for k := range types {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	err := enc.EncodeMapLen(len(keys))
	if err != nil {
		return p.NewErrorf("error encoding object length: %w", err)
	}
	for _, k := range keys {
		ty := types[k]
		v, ok := o[k]
		if !ok {
			return p.WithAttributeName(k).NewErrorf("no value set")
		}
		err := marshalMsgPack(NewValue(String, k), String, p.WithAttributeName(k), enc)
		if err != nil {
			return err
		}
		err = marshalMsgPack(v, ty, p.WithAttributeName(k), enc)
		if err != nil {
			return err
		}
	}
	return nil
}
