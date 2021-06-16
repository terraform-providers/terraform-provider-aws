package tfprotov5

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ErrUnknownDynamicValueType is returned when a DynamicValue has no MsgPack or
// JSON bytes set. This should never be returned during the normal operation of
// a provider, and indicates one of the following:
//
// 1. terraform-plugin-go is out of sync with the protocol and should be
// updated.
//
// 2. terrafrom-plugin-go has a bug.
//
// 3. The `DynamicValue` was generated or modified by something other than
// terraform-plugin-go and is no longer a valid value.
var ErrUnknownDynamicValueType = errors.New("DynamicValue had no JSON or msgpack data set")

// NewDynamicValue creates a DynamicValue from a tftypes.Value. You must
// specify the tftype.Type you want to send the value as, and it must be a type
// that is compatible with the Type of the Value. Usually it should just be the
// Type of the Value, but it can also be the DynamicPseudoType.
func NewDynamicValue(t tftypes.Type, v tftypes.Value) (DynamicValue, error) {
	b, err := v.MarshalMsgPack(t) //nolint:staticcheck
	if err != nil {
		return DynamicValue{}, err
	}
	return DynamicValue{
		MsgPack: b,
	}, nil
}

// DynamicValue represents a nested encoding value that came from the protocol.
// The only way providers should ever interact with it is by calling its
// `Unmarshal` method to retrive a `tftypes.Value`. Although the type system
// allows for other interactions, they are explicitly not supported, and will
// not be considered when evaluating for breaking changes. Treat this type as
// an opaque value, and *only* call its `Unmarshal` method.
type DynamicValue struct {
	MsgPack []byte
	JSON    []byte
}

// Unmarshal returns a `tftypes.Value` that represents the information
// contained in the DynamicValue in an easy-to-interact-with way. It is the
// main purpose of the DynamicValue type, and is how provider developers should
// obtain config, state, and other values from the protocol.
//
// Pass in the type you want the `Value` to be interpreted as. Terraform's type
// system encodes in a lossy manner, meaning the type information is not
// preserved losslessly when going over the wire. Sets, lists, and tuples all
// look the same, as do user-specified values when the provider has a
// DynamicPseudoType in its schema. Objects and maps all look the same, as
// well, as do DynamicPseudoType values sometimes. Fortunately, the provider
// should already know the type; it should be the type of the schema, or
// PseudoDynamicType if that's what's in the schema. `Unmarshal` will then
// parse the value as though it belongs to that type, if possible, and return a
// `tftypes.Value` with the appropriate information. If the data can't be
// interpreted as that type, an error will be returned saying so. In these
// cases, double check to make sure the schema is declaring the same type being
// passed into `Unmarshal`.
//
// In the event an ErrUnknownDynamicValueType is returned, one of three things
// has happened:
//
// 1. terraform-plugin-go is out of date and out of sync with the protocol, and
// an issue should be opened on its repo to get it updated.
//
// 2. terraform-plugin-go has a bug somewhere, and an issue should be opened on
// its repo to get it fixed.
//
// 3. The provider or a dependency has modified the `DynamicValue` in an
// unsupported way, or has created one from scratch, and should treat it as
// opaque and not modify it, only calling `Unmarshal` on `DynamicValue`s
// received from RPC requests.
func (d DynamicValue) Unmarshal(typ tftypes.Type) (tftypes.Value, error) {
	if d.JSON != nil {
		return tftypes.ValueFromJSON(d.JSON, typ) //nolint:staticcheck
	}
	if d.MsgPack != nil {
		return tftypes.ValueFromMsgPack(d.MsgPack, typ) //nolint:staticcheck
	}
	return tftypes.Value{}, ErrUnknownDynamicValueType
}
