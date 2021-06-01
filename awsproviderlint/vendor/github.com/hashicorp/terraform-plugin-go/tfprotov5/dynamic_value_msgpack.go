package tfprotov5

import (
	"math/big"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"github.com/vmihailenco/msgpack"
	msgpackCodes "github.com/vmihailenco/msgpack/codes"
)

func msgpackUnmarshal(dec *msgpack.Decoder, typ tftypes.Type, path tftypes.AttributePath) (tftypes.Value, error) {
	peek, err := dec.PeekCode()
	if err != nil {
		return tftypes.Value{}, path.NewErrorf("error peeking next byte: %w", err)
	}
	if msgpackCodes.IsExt(peek) {
		// as with go-cty, assume all extensions are unknown values
		err := dec.Skip()
		if err != nil {
			return tftypes.Value{}, path.NewErrorf("error skipping extension byte: %w", err)
		}
		return tftypes.NewValue(typ, tftypes.UnknownValue), nil
	}
	if typ.Is(tftypes.DynamicPseudoType) {
		return msgpackUnmarshalDynamic(dec, path)
	}
	if peek == msgpackCodes.Nil {
		err := dec.Skip()
		if err != nil {
			return tftypes.Value{}, path.NewErrorf("error skipping nil byte: %w", err)
		}
		return tftypes.NewValue(typ, nil), nil
	}

	switch {
	case typ.Is(tftypes.String):
		rv, err := dec.DecodeString()
		if err != nil {
			return tftypes.Value{}, path.NewErrorf("error decoding string: %w", err)
		}
		return tftypes.NewValue(tftypes.String, rv), nil
	case typ.Is(tftypes.Number):
		peek, err := dec.PeekCode()
		if err != nil {
			return tftypes.Value{}, path.NewErrorf("couldn't peek number: %w", err)
		}
		if msgpackCodes.IsFixedNum(peek) {
			rv, err := dec.DecodeInt64()
			if err != nil {
				return tftypes.Value{}, path.NewErrorf("couldn't decode number as int64: %w", err)
			}
			return tftypes.NewValue(tftypes.Number, big.NewFloat(float64(rv))), nil
		}
		switch peek {
		case msgpackCodes.Int8, msgpackCodes.Int16, msgpackCodes.Int32, msgpackCodes.Int64:
			rv, err := dec.DecodeInt64()
			if err != nil {
				return tftypes.Value{}, path.NewErrorf("couldn't decode number as int64: %w", err)
			}
			return tftypes.NewValue(tftypes.Number, big.NewFloat(float64(rv))), nil
		case msgpackCodes.Uint8, msgpackCodes.Uint16, msgpackCodes.Uint32, msgpackCodes.Uint64:
			rv, err := dec.DecodeUint64()
			if err != nil {
				return tftypes.Value{}, path.NewErrorf("couldn't decode number as uint64: %w", err)
			}
			return tftypes.NewValue(tftypes.Number, big.NewFloat(float64(rv))), nil
		case msgpackCodes.Float, msgpackCodes.Double:
			rv, err := dec.DecodeFloat64()
			if err != nil {
				return tftypes.Value{}, path.NewErrorf("couldn't decode number as float64: %w", err)
			}
			return tftypes.NewValue(tftypes.Number, big.NewFloat(float64(rv))), nil
		default:
			rv, err := dec.DecodeString()
			if err != nil {
				return tftypes.Value{}, path.NewErrorf("couldn't decode number as string: %w", err)
			}
			// according to
			// https://github.com/hashicorp/go-cty/blob/85980079f637862fa8e43ddc82dd74315e2f4c85/cty/value_init.go#L49
			// Base 10, precision 512, and rounding to nearest even
			// is the standard way to handle numbers arriving as
			// strings.
			fv, _, err := big.ParseFloat(rv, 10, 512, big.ToNearestEven)
			if err != nil {
				return tftypes.Value{}, path.NewErrorf("error parsing %q as number: %w", rv, err)
			}
			return tftypes.NewValue(tftypes.Number, fv), nil
		}
	case typ.Is(tftypes.Bool):
		rv, err := dec.DecodeBool()
		if err != nil {
			return tftypes.Value{}, path.NewErrorf("couldn't decode bool: %w", err)
		}
		return tftypes.NewValue(tftypes.Bool, rv), nil
	case typ.Is(tftypes.List{}):
		return msgpackUnmarshalList(dec, typ.(tftypes.List).ElementType, path)
	case typ.Is(tftypes.Set{}):
		return msgpackUnmarshalSet(dec, typ.(tftypes.Set).ElementType, path)
	case typ.Is(tftypes.Map{}):
		return msgpackUnmarshalMap(dec, typ.(tftypes.Map).AttributeType, path)
	case typ.Is(tftypes.Tuple{}):
		return msgpackUnmarshalTuple(dec, typ.(tftypes.Tuple).ElementTypes, path)
	case typ.Is(tftypes.Object{}):
		return msgpackUnmarshalObject(dec, typ.(tftypes.Object).AttributeTypes, path)
	}
	return tftypes.Value{}, path.NewErrorf("unsupported type %s", typ.String())
}

func msgpackUnmarshalList(dec *msgpack.Decoder, typ tftypes.Type, path tftypes.AttributePath) (tftypes.Value, error) {
	length, err := dec.DecodeArrayLen()
	if err != nil {
		return tftypes.Value{}, path.NewErrorf("error decoding list length: %w", err)
	}

	switch {
	case length < 0:
		return tftypes.NewValue(tftypes.List{
			ElementType: typ,
		}, nil), nil
	case length == 0:
		return tftypes.NewValue(tftypes.List{
			ElementType: typ,
		}, []tftypes.Value{}), nil
	}

	vals := make([]tftypes.Value, 0, length)
	for i := 0; i < length; i++ {
		path.WithElementKeyInt(int64(i))
		val, err := msgpackUnmarshal(dec, typ, path)
		if err != nil {
			return tftypes.Value{}, err
		}
		vals = append(vals, val)
		path.WithoutLastStep()
	}

	elTyp := typ
	if elTyp == tftypes.DynamicPseudoType {
		elTyp, err = tftypes.TypeFromElements(vals)
		if err != nil {
			return tftypes.Value{}, err
		}
	}

	return tftypes.NewValue(tftypes.List{
		ElementType: elTyp,
	}, vals), nil
}

func msgpackUnmarshalSet(dec *msgpack.Decoder, typ tftypes.Type, path tftypes.AttributePath) (tftypes.Value, error) {
	length, err := dec.DecodeArrayLen()
	if err != nil {
		return tftypes.Value{}, path.NewErrorf("error decoding set length: %w", err)
	}

	switch {
	case length < 0:
		return tftypes.NewValue(tftypes.Set{
			ElementType: typ,
		}, nil), nil
	case length == 0:
		return tftypes.NewValue(tftypes.Set{
			ElementType: typ,
		}, []tftypes.Value{}), nil
	}

	vals := make([]tftypes.Value, 0, length)
	for i := 0; i < length; i++ {
		path.WithElementKeyInt(int64(i))
		val, err := msgpackUnmarshal(dec, typ, path)
		if err != nil {
			return tftypes.Value{}, err
		}
		vals = append(vals, val)
		path.WithoutLastStep()
	}

	elTyp, err := tftypes.TypeFromElements(vals)
	if err != nil {
		return tftypes.Value{}, err
	}

	return tftypes.NewValue(tftypes.Set{
		ElementType: elTyp,
	}, vals), nil
}

func msgpackUnmarshalMap(dec *msgpack.Decoder, typ tftypes.Type, path tftypes.AttributePath) (tftypes.Value, error) {
	length, err := dec.DecodeMapLen()
	if err != nil {
		return tftypes.Value{}, path.NewErrorf("error decoding map length: %w", err)
	}

	switch {
	case length < 0:
		return tftypes.NewValue(tftypes.Map{
			AttributeType: typ,
		}, nil), nil
	case length == 0:
		return tftypes.NewValue(tftypes.Map{
			AttributeType: typ,
		}, map[string]tftypes.Value{}), nil
	}

	vals := make(map[string]tftypes.Value, length)
	for i := 0; i < length; i++ {
		key, err := dec.DecodeString()
		if err != nil {
			return tftypes.Value{}, path.NewErrorf("error decoding map key: %w", err)
		}
		path.WithElementKeyString(key)
		val, err := msgpackUnmarshal(dec, typ, path)
		if err != nil {
			return tftypes.Value{}, err
		}
		vals[key] = val
		path.WithoutLastStep()
	}
	return tftypes.NewValue(tftypes.Map{
		AttributeType: typ,
	}, vals), nil
}

func msgpackUnmarshalTuple(dec *msgpack.Decoder, types []tftypes.Type, path tftypes.AttributePath) (tftypes.Value, error) {
	length, err := dec.DecodeArrayLen()
	if err != nil {
		return tftypes.Value{}, path.NewErrorf("error decoding tuple length: %w", err)
	}

	switch {
	case length < 0:
		return tftypes.NewValue(tftypes.Tuple{
			ElementTypes: types,
		}, nil), nil
	case length == 0:
		return tftypes.NewValue(tftypes.Tuple{
			// no elements means no types
			ElementTypes: nil,
		}, []tftypes.Value{}), nil
	case length != len(types):
		return tftypes.Value{}, path.NewErrorf("error decoding tuple; expected %d items, got %d", len(types), length)
	}

	vals := make([]tftypes.Value, 0, length)
	for i := 0; i < length; i++ {
		path.WithElementKeyInt(int64(i))
		typ := types[i]
		val, err := msgpackUnmarshal(dec, typ, path)
		if err != nil {
			return tftypes.Value{}, err
		}
		vals = append(vals, val)
		path.WithoutLastStep()
	}

	return tftypes.NewValue(tftypes.Tuple{
		ElementTypes: types,
	}, vals), nil
}

func msgpackUnmarshalObject(dec *msgpack.Decoder, types map[string]tftypes.Type, path tftypes.AttributePath) (tftypes.Value, error) {
	length, err := dec.DecodeMapLen()
	if err != nil {
		return tftypes.Value{}, path.NewErrorf("error decoding object length: %w", err)
	}

	switch {
	case length < 0:
		return tftypes.NewValue(tftypes.Object{
			AttributeTypes: types,
		}, nil), nil
	case length == 0:
		return tftypes.NewValue(tftypes.Object{
			// no attributes means no types
			AttributeTypes: nil,
		}, map[string]tftypes.Value{}), nil
	case length != len(types):
		return tftypes.Value{}, path.NewErrorf("error decoding object; expected %d attributes, got %d", len(types), length)
	}

	vals := make(map[string]tftypes.Value, length)
	for i := 0; i < length; i++ {
		key, err := dec.DecodeString()
		if err != nil {
			return tftypes.Value{}, path.NewErrorf("error decoding object key: %w", err)
		}
		typ, exists := types[key]
		if !exists {
			return tftypes.Value{}, path.NewErrorf("unknown attribute %q", key)
		}
		path.WithAttributeName(key)
		val, err := msgpackUnmarshal(dec, typ, path)
		if err != nil {
			return tftypes.Value{}, err
		}
		vals[key] = val
		path.WithoutLastStep()
	}

	return tftypes.NewValue(tftypes.Object{
		AttributeTypes: types,
	}, vals), nil
}

func msgpackUnmarshalDynamic(dec *msgpack.Decoder, path tftypes.AttributePath) (tftypes.Value, error) {
	length, err := dec.DecodeArrayLen()
	if err != nil {
		return tftypes.Value{}, path.NewErrorf("error checking length of DynamicPseudoType value: %w", err)
	}

	switch {
	case length == -1:
		return tftypes.NewValue(tftypes.DynamicPseudoType, nil), nil
	case length != 2:
		return tftypes.Value{}, path.NewErrorf("expected %d elements in DynamicPseudoType array, got %d", 2, length)
	}

	typeJSON, err := dec.DecodeBytes()
	if err != nil {
		return tftypes.Value{}, path.NewErrorf("error decoding bytes: %w", err)
	}
	typ, err := tftypes.ParseJSONType(typeJSON) //nolint:staticcheck
	if err != nil {
		return tftypes.Value{}, path.NewErrorf("error parsing type information: %w", err)
	}
	return msgpackUnmarshal(dec, typ, path)
}
