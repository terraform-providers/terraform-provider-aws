package addrs

import (
	"fmt"
)

// instanceKey represents the key of an instance within an object that
// contains multiple instances due to using "count" or "for_each" arguments
// in configuration.
//
// intKey and stringKey are the two implementations of this type. No other
// implementations are allowed. The single instance of an object that _isn't_
// using "count" or "for_each" is represented by NoKey, which is a nil
// InstanceKey.
type instanceKey interface {
	instanceKeySigil()
	String() string
}

// NoKey represents the absense of an instanceKey, for the single instance
// of a configuration object that does not use "count" or "for_each" at all.
var NoKey instanceKey

// intKey is the InstanceKey representation representing integer indices, as
// used when the "count" argument is specified or if for_each is used with
// a sequence type.
type intKey int

func (k intKey) instanceKeySigil() {
}

func (k intKey) String() string {
	return fmt.Sprintf("[%d]", int(k))
}

// stringKey is the InstanceKey representation representing string indices, as
// used when the "for_each" argument is specified with a map or object type.
type stringKey string

func (k stringKey) instanceKeySigil() {
}

func (k stringKey) String() string {
	// FIXME: This isn't _quite_ right because Go's quoted string syntax is
	// slightly different than HCL's, but we'll accept it for now.
	return fmt.Sprintf("[%q]", string(k))
}
