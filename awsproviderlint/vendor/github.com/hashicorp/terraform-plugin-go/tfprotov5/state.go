package tfprotov5

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ErrUnknownRawStateType is returned when a RawState has no Flatmap or JSON
// bytes set. This should never be returned during the normal operation of a
// provider, and indicates one of the following:
//
// 1. terraform-plugin-go is out of sync with the protocol and should be
// updated.
//
// 2. terrafrom-plugin-go has a bug.
//
// 3. The `RawState` was generated or modified by something other than
// terraform-plugin-go and is no longer a valid value.
var ErrUnknownRawStateType = errors.New("RawState had no JSON or flatmap data set")

// RawState is the raw, undecoded state for providers to upgrade. It is
// undecoded as Terraform, for whatever reason, doesn't have the previous
// schema available to it, and so cannot decode the state itself and pushes
// that responsibility off onto providers.
//
// It is safe to assume that Flatmap can be ignored for any state written by
// Terraform 0.12.0 or higher, but it is not safe to assume that all states
// written by 0.12.0 or higher will be in JSON format; future versions may
// switch to an alternate encoding for states.
type RawState struct {
	JSON    []byte
	Flatmap map[string]string
}

// Unmarshal returns a `tftypes.Value` that represents the information
// contained in the RawState in an easy-to-interact-with way. It is the
// main purpose of the RawState type, and is how provider developers should
// obtain state values from the UpgradeResourceState RPC call.
//
// Pass in the type you want the `Value` to be interpreted as. Terraform's type
// system encodes in a lossy manner, meaning the type information is not
// preserved losslessly when going over the wire. Sets, lists, and tuples all
// look the same. Objects and maps all look the same, as well, as do
// user-specified values when DynamicPseudoType is used in the schema.
// Fortunately, the provider should already know the type; it should be the
// type of the schema, or DynamicPseudoType if that's what's in the schema.
// `Unmarshal` will then parse the value as though it belongs to that type, if
// possible, and return a `tftypes.Value` with the appropriate information. If
// the data can't be interpreted as that type, an error will be returned saying
// so. In these cases, double check to make sure the schema is declaring the
// same type being passed into `Unmarshal`.
//
// In the event an ErrUnknownRawStateType is returned, one of three things
// has happened:
//
// 1. terraform-plugin-go is out of date and out of sync with the protocol, and
// an issue should be opened on its repo to get it updated.
//
// 2. terraform-plugin-go has a bug somewhere, and an issue should be opened on
// its repo to get it fixed.
//
// 3. The provider or a dependency has modified the `RawState` in an
// unsupported way, or has created one from scratch, and should treat it as
// opaque and not modify it, only calling `Unmarshal` on `RawState`s received
// from RPC requests.
//
// State files written before Terraform 0.12 that haven't been upgraded yet
// cannot be unmarshaled, and must have their Flatmap property read directly.
func (s RawState) Unmarshal(typ tftypes.Type) (tftypes.Value, error) {
	if s.JSON != nil {
		return tftypes.ValueFromJSON(s.JSON, typ) //nolint:staticcheck
	}
	if s.Flatmap != nil {
		return tftypes.Value{}, fmt.Errorf("flatmap states cannot be unmarshaled, only states written by Terraform 0.12 and higher can be unmarshaled")
	}
	return tftypes.Value{}, ErrUnknownRawStateType
}
