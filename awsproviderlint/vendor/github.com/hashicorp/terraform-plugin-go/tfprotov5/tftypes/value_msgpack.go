package tftypes

import (
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/vmihailenco/msgpack"
)

func marshalMsgPack(val Value, typ Type, p AttributePath, enc *msgpack.Encoder) error {
	if typ.Is(DynamicPseudoType) && !val.Is(DynamicPseudoType) {
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

func marshalMsgPackDynamicPseudoType(val Value, typ Type, p AttributePath, enc *msgpack.Encoder) error {
	typeJSON, err := val.typ.MarshalJSON()
	if err != nil {
		return p.NewErrorf("error generating JSON for type %s: %w", val.typ, err)
	}
	err = enc.EncodeArrayLen(2)
	if err != nil {
		return p.NewErrorf("error encoding array length:  %w", err)
	}
	err = enc.EncodeBytes(typeJSON)
	if err != nil {
		return p.NewErrorf("error encoding JSON type info: %w", err)
	}
	err = marshalMsgPack(val, val.typ, p, enc)
	if err != nil {
		return p.NewErrorf("error marshaling DynamicPseudoType value: %w", err)
	}
	return nil
}

func marshalMsgPackString(val Value, typ Type, p AttributePath, enc *msgpack.Encoder) error {
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

func marshalMsgPackNumber(val Value, typ Type, p AttributePath, enc *msgpack.Encoder) error {
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

func marshalMsgPackBool(val Value, typ Type, p AttributePath, enc *msgpack.Encoder) error {
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

func marshalMsgPackList(val Value, typ Type, p AttributePath, enc *msgpack.Encoder) error {
	l, ok := val.value.([]Value)
	if !ok {
		return unexpectedValueTypeError(p, l, val.value, typ)
	}
	err := enc.EncodeArrayLen(len(l))
	if err != nil {
		return p.NewErrorf("error encoding list length: %w", err)
	}
	for pos, i := range l {
		p.WithElementKeyInt(int64(pos))
		err := marshalMsgPack(i, typ.(List).ElementType, p, enc)
		if err != nil {
			return err
		}
		p.WithoutLastStep()
	}
	return nil
}

func marshalMsgPackSet(val Value, typ Type, p AttributePath, enc *msgpack.Encoder) error {
	s, ok := val.value.([]Value)
	if !ok {
		return unexpectedValueTypeError(p, s, val.value, typ)
	}
	err := enc.EncodeArrayLen(len(s))
	if err != nil {
		return p.NewErrorf("error encoding set length: %w", err)
	}
	for _, i := range s {
		p.WithElementKeyValue(i)
		err := marshalMsgPack(i, typ.(Set).ElementType, p, enc)
		if err != nil {
			return err
		}
		p.WithoutLastStep()
	}
	return nil
}

func marshalMsgPackMap(val Value, typ Type, p AttributePath, enc *msgpack.Encoder) error {
	m, ok := val.value.(map[string]Value)
	if !ok {
		return unexpectedValueTypeError(p, m, val.value, typ)
	}
	err := enc.EncodeMapLen(len(m))
	if err != nil {
		return p.NewErrorf("error encoding map length: %w", err)
	}
	for k, v := range m {
		p.WithElementKeyString(k)
		err := marshalMsgPack(NewValue(String, k), String, p, enc)
		if err != nil {
			return p.NewErrorf("error encoding map key: %w", err)
		}
		err = marshalMsgPack(v, typ.(Map).AttributeType, p, enc)
		if err != nil {
			return err
		}
		p.WithoutLastStep()
	}
	return nil
}

func marshalMsgPackTuple(val Value, typ Type, p AttributePath, enc *msgpack.Encoder) error {
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
		p.WithElementKeyInt(int64(pos))
		ty := types[pos]
		err := marshalMsgPack(v, ty, p, enc)
		if err != nil {
			return err
		}
		p.WithoutLastStep()
	}
	return nil
}

func marshalMsgPackObject(val Value, typ Type, p AttributePath, enc *msgpack.Encoder) error {
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
		p.WithAttributeName(k)
		ty := types[k]
		v, ok := o[k]
		if !ok {
			return p.NewErrorf("no value set")
		}
		err := marshalMsgPack(NewValue(String, k), String, p, enc)
		if err != nil {
			return err
		}
		err = marshalMsgPack(v, ty, p, enc)
		if err != nil {
			return err
		}
		p.WithoutLastStep()
	}
	return nil
}
