package tfprotov5

import (
	"bytes"
	"encoding/json"
	"math/big"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

func jsonByteDecoder(buf []byte) *json.Decoder {
	r := bytes.NewReader(buf)
	dec := json.NewDecoder(r)
	dec.UseNumber()
	return dec
}

func jsonUnmarshal(buf []byte, typ tftypes.Type, p tftypes.AttributePath) (tftypes.Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}

	if tok == nil {
		return tftypes.NewValue(typ, nil), nil
	}

	switch {
	case typ.Is(tftypes.String):
		return jsonUnmarshalString(buf, typ, p)
	case typ.Is(tftypes.Number):
		return jsonUnmarshalNumber(buf, typ, p)
	case typ.Is(tftypes.Bool):
		return jsonUnmarshalBool(buf, typ, p)
	case typ.Is(tftypes.DynamicPseudoType):
		return jsonUnmarshalDynamicPseudoType(buf, typ, p)
	case typ.Is(tftypes.List{}):
		return jsonUnmarshalList(buf, typ.(tftypes.List).ElementType, p)
	case typ.Is(tftypes.Set{}):
		return jsonUnmarshalSet(buf, typ.(tftypes.Set).ElementType, p)

	case typ.Is(tftypes.Map{}):
		return jsonUnmarshalMap(buf, typ.(tftypes.Map).AttributeType, p)
	case typ.Is(tftypes.Tuple{}):
		return jsonUnmarshalTuple(buf, typ.(tftypes.Tuple).ElementTypes, p)
	case typ.Is(tftypes.Object{}):
		return jsonUnmarshalObject(buf, typ.(tftypes.Object).AttributeTypes, p)
	}
	return tftypes.Value{}, p.NewErrorf("unknown type %s", typ)
}

func jsonUnmarshalString(buf []byte, typ tftypes.Type, p tftypes.AttributePath) (tftypes.Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}
	switch v := tok.(type) {
	case string:
		return tftypes.NewValue(tftypes.String, v), nil
	case json.Number:
		return tftypes.NewValue(tftypes.String, string(v)), nil
	case bool:
		if v {
			return tftypes.NewValue(tftypes.String, "true"), nil
		}
		return tftypes.NewValue(tftypes.String, "false"), nil
	}
	return tftypes.Value{}, p.NewErrorf("unsupported type %T sent as %s", tok, tftypes.String)
}

func jsonUnmarshalNumber(buf []byte, typ tftypes.Type, p tftypes.AttributePath) (tftypes.Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}
	switch numTok := tok.(type) {
	case json.Number:
		f, _, err := big.ParseFloat(string(numTok), 10, 512, big.ToNearestEven)
		if err != nil {
			return tftypes.Value{}, p.NewErrorf("error parsing number: %w", err)
		}
		return tftypes.NewValue(typ, f), nil
	case string:
		f, _, err := big.ParseFloat(string(numTok), 10, 512, big.ToNearestEven)
		if err != nil {
			return tftypes.Value{}, p.NewErrorf("error parsing number: %w", err)
		}
		return tftypes.NewValue(typ, f), nil
	}
	return tftypes.Value{}, p.NewErrorf("unsupported type %T sent as %s", tok, tftypes.Number)
}

func jsonUnmarshalBool(buf []byte, typ tftypes.Type, p tftypes.AttributePath) (tftypes.Value, error) {
	dec := jsonByteDecoder(buf)
	tok, err := dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}
	switch v := tok.(type) {
	case bool:
		return tftypes.NewValue(tftypes.Bool, v), nil
	case string:
		switch v {
		case "true", "1":
			return tftypes.NewValue(tftypes.Bool, true), nil
		case "false", "0":
			return tftypes.NewValue(tftypes.Bool, false), nil
		}
		switch strings.ToLower(string(v)) {
		case "true":
			return tftypes.Value{}, p.NewErrorf("to convert from string, use lowercase \"true\"")
		case "false":
			return tftypes.Value{}, p.NewErrorf("to convert from string, use lowercase \"false\"")
		}
	case json.Number:
		switch v {
		case "1":
			return tftypes.NewValue(tftypes.Bool, true), nil
		case "0":
			return tftypes.NewValue(tftypes.Bool, false), nil
		}
	}
	return tftypes.Value{}, p.NewErrorf("unsupported type %T sent as %s", tok, tftypes.Bool)
}

func jsonUnmarshalDynamicPseudoType(buf []byte, typ tftypes.Type, p tftypes.AttributePath) (tftypes.Value, error) {
	dec := jsonByteDecoder(buf)
	tok, err := dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('{') {
		return tftypes.Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('{'), tok)
	}
	var t tftypes.Type
	var valBody []byte
	for dec.More() {
		tok, err = dec.Token()
		if err != nil {
			return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
		}
		key, ok := tok.(string)
		if !ok {
			return tftypes.Value{}, p.NewErrorf("expected key to be a string, got %T", tok)
		}
		var rawVal json.RawMessage
		err = dec.Decode(&rawVal)
		if err != nil {
			return tftypes.Value{}, p.NewErrorf("error decoding value: %w", err)
		}
		switch key {
		case "type":
			t, err = tftypes.ParseJSONType(rawVal) //nolint:staticcheck
			if err != nil {
				return tftypes.Value{}, p.NewErrorf("error decoding type information: %w", err)
			}
		case "value":
			valBody = rawVal
		default:
			return tftypes.Value{}, p.NewErrorf("invalid key %q in dynamically-typed value", key)
		}
	}
	tok, err = dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('}') {
		return tftypes.Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('}'), tok)
	}
	if t == nil {
		return tftypes.Value{}, p.NewErrorf("missing type in dynamically-typed value")
	}
	if valBody == nil {
		return tftypes.Value{}, p.NewErrorf("missing value in dynamically-typed value")
	}
	return jsonUnmarshal(valBody, t, p)
}

func jsonUnmarshalList(buf []byte, elementType tftypes.Type, p tftypes.AttributePath) (tftypes.Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('[') {
		return tftypes.Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('['), tok)
	}

	// we want to have a value for this always, even if there are no
	// elements, because no elements is *technically* different than empty,
	// and we want to preserve that distinction
	//
	// var vals []tftypes.Value
	// would evaluate as nil if the list is empty
	//
	// while generally in Go it's undesirable to treat empty and nil slices
	// separately, in this case we're surfacing a non-Go-in-origin
	// distinction, so we'll allow it.
	vals := []tftypes.Value{}

	var idx int64
	for dec.More() {
		p.WithElementKeyInt(idx)
		// update the index
		idx++

		var rawVal json.RawMessage
		err = dec.Decode(&rawVal)
		if err != nil {
			return tftypes.Value{}, p.NewErrorf("error decoding value: %w", err)
		}
		val, err := jsonUnmarshal(rawVal, elementType, p)
		if err != nil {
			return tftypes.Value{}, err
		}
		vals = append(vals, val)
		p.WithoutLastStep()
	}

	tok, err = dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim(']') {
		return tftypes.Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim(']'), tok)
	}

	elTyp := elementType
	if elTyp == tftypes.DynamicPseudoType {
		elTyp, err = tftypes.TypeFromElements(vals)
		if err != nil {
			return tftypes.Value{}, p.NewErrorf("invalid elements for list: %w", err)
		}
	}
	return tftypes.NewValue(tftypes.List{
		ElementType: elTyp,
	}, vals), nil
}

func jsonUnmarshalSet(buf []byte, elementType tftypes.Type, p tftypes.AttributePath) (tftypes.Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('[') {
		return tftypes.Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('['), tok)
	}

	// we want to have a value for this always, even if there are no
	// elements, because no elements is *technically* different than empty,
	// and we want to preserve that distinction
	//
	// var vals []tftypes.Value
	// would evaluate as nil if the set is empty
	//
	// while generally in Go it's undesirable to treat empty and nil slices
	// separately, in this case we're surfacing a non-Go-in-origin
	// distinction, so we'll allow it.
	vals := []tftypes.Value{}

	for dec.More() {
		p.WithElementKeyValue(tftypes.NewValue(elementType, tftypes.UnknownValue))
		var rawVal json.RawMessage
		err = dec.Decode(&rawVal)
		if err != nil {
			return tftypes.Value{}, p.NewErrorf("error decoding value: %w", err)
		}
		val, err := jsonUnmarshal(rawVal, elementType, p)
		if err != nil {
			return tftypes.Value{}, err
		}
		vals = append(vals, val)
		p.WithoutLastStep()
	}
	tok, err = dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim(']') {
		return tftypes.Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim(']'), tok)
	}

	elTyp := elementType
	if elTyp == tftypes.DynamicPseudoType {
		elTyp, err = tftypes.TypeFromElements(vals)
		if err != nil {
			return tftypes.Value{}, p.NewErrorf("invalid elements for list: %w", err)
		}
	}
	return tftypes.NewValue(tftypes.Set{
		ElementType: elTyp,
	}, vals), nil
}

func jsonUnmarshalMap(buf []byte, attrType tftypes.Type, p tftypes.AttributePath) (tftypes.Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('{') {
		return tftypes.Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('{'), tok)
	}

	vals := map[string]tftypes.Value{}
	for dec.More() {
		p.WithElementKeyValue(tftypes.NewValue(attrType, tftypes.UnknownValue))
		tok, err := dec.Token()
		if err != nil {
			return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
		}
		key, ok := tok.(string)
		if !ok {
			return tftypes.Value{}, p.NewErrorf("expected map key to be a string, got %T", tok)
		}

		//fix the path value, we have an actual key now
		p.WithoutLastStep()
		p.WithElementKeyString(key)

		var rawVal json.RawMessage
		err = dec.Decode(&rawVal)
		if err != nil {
			return tftypes.Value{}, p.NewErrorf("error decoding value: %w", err)
		}
		val, err := jsonUnmarshal(rawVal, attrType, p)
		if err != nil {
			return tftypes.Value{}, err
		}
		vals[key] = val
		p.WithoutLastStep()
	}
	tok, err = dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('}') {
		return tftypes.Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('}'), tok)
	}

	return tftypes.NewValue(tftypes.Map{
		AttributeType: attrType,
	}, vals), nil
}

func jsonUnmarshalTuple(buf []byte, elementTypes []tftypes.Type, p tftypes.AttributePath) (tftypes.Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('[') {
		return tftypes.Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('['), tok)
	}

	// we want to have a value for this always, even if there are no
	// elements, because no elements is *technically* different than empty,
	// and we want to preserve that distinction
	//
	// var vals []tftypes.Value
	// would evaluate as nil if the tuple is empty
	//
	// while generally in Go it's undesirable to treat empty and nil slices
	// separately, in this case we're surfacing a non-Go-in-origin
	// distinction, so we'll allow it.
	vals := []tftypes.Value{}

	var idx int64
	for dec.More() {
		if idx >= int64(len(elementTypes)) {
			return tftypes.Value{}, p.NewErrorf("too many tuple elements (only have types for %d)", len(elementTypes))
		}

		p.WithElementKeyInt(idx)
		elementType := elementTypes[idx]
		idx++

		var rawVal json.RawMessage
		err = dec.Decode(&rawVal)
		if err != nil {
			return tftypes.Value{}, p.NewErrorf("error decoding value: %w", err)
		}
		val, err := jsonUnmarshal(rawVal, elementType, p)
		if err != nil {
			return tftypes.Value{}, err
		}
		vals = append(vals, val)
		p.WithoutLastStep()
	}

	tok, err = dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim(']') {
		return tftypes.Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim(']'), tok)
	}

	if len(vals) != len(elementTypes) {
		return tftypes.Value{}, p.NewErrorf("not enough tuple elements (only have %d, have types for %d)", len(vals), len(elementTypes))
	}

	return tftypes.NewValue(tftypes.Tuple{
		ElementTypes: elementTypes,
	}, vals), nil
}

func jsonUnmarshalObject(buf []byte, attrTypes map[string]tftypes.Type, p tftypes.AttributePath) (tftypes.Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('{') {
		return tftypes.Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('{'), tok)
	}

	vals := map[string]tftypes.Value{}
	for dec.More() {
		p.WithElementKeyValue(tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
		tok, err := dec.Token()
		if err != nil {
			return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
		}
		key, ok := tok.(string)
		if !ok {
			return tftypes.Value{}, p.NewErrorf("object attribute key was %T, not string", tok)
		}
		attrType, ok := attrTypes[key]
		if !ok {
			return tftypes.Value{}, p.NewErrorf("unsupported attribute %q", key)
		}
		p.WithoutLastStep()
		p.WithAttributeName(key)

		var rawVal json.RawMessage
		err = dec.Decode(&rawVal)
		if err != nil {
			return tftypes.Value{}, p.NewErrorf("error decoding value: %w", err)
		}
		val, err := jsonUnmarshal(rawVal, attrType, p)
		if err != nil {
			return tftypes.Value{}, err
		}
		vals[key] = val
		p.WithoutLastStep()
	}

	tok, err = dec.Token()
	if err != nil {
		return tftypes.Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('}') {
		return tftypes.Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('}'), tok)
	}

	// make sure we have a value for every attribute
	for k, typ := range attrTypes {
		if _, ok := vals[k]; !ok {
			vals[k] = tftypes.NewValue(typ, nil)
		}
	}

	return tftypes.NewValue(tftypes.Object{
		AttributeTypes: attrTypes,
	}, vals), nil
}
