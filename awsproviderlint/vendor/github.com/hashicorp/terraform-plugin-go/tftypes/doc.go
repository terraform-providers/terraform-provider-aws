// Package tftypes provides a type system for Terraform configuration and state
// values.
//
// Terraform's configuration and state values are stored using a collection of
// types. There are primitive types, such as strings, numbers, and booleans,
// but there are also aggregate types such as lists, sets, tuples, maps, and
// objects, which consist of multiple values of primitive types aggregated into
// a single value. There is also a dynamic pseudo-type that represents an
// unknown type. It is useful for indicating that any type of data is
// acceptable.
//
// Terraform's values map neatly onto either primitives built into Go or types
// in the Go standard library, with one exception. Terraform has the concept of
// unknown values, values that may or may not be set at a future date. These
// are distinct from null values, which indicate a value that is known to not
// be set, and are mostly encountered when a user has interpolated a computed
// field into another field; the field that is interpolated into has an unknown
// value, because the field being interpolated won't have its value known until
// apply time.
//
// To address this, the tftypes package wraps all values in a special Value
// type. This Value type is capable of holding known and unknown values,
// interrogating whether the value is known or not, and accessing the concrete
// value that Terraform sent in the cases where the value is known. A common
// pattern is to use the Value.IsKnown() method to confirm that a value is
// known, then to use the Value.As() method to retrieve the underlying data for
// use.
//
// When using the Value.As() method, certain types have built-in behavior to
// support using them as destinations for converted data:
//
// * String values can be converted into strings
//
// * Number values can be converted into *big.Floats
//
// * Boolean values can be converted into bools
//
// * List, Set, and Tuple values can be converted into a slice of Values
//
// * Map and Object values can be converted into a map with string keys and
// Value values.
//
// These defaults were chosen because they're capable of losslessly
// representing all possible values for their Terraform type, with the
// exception of null values. Converting into pointer versions of any of these
// types will correctly surface null values as well.
//
// Custom, provider-defined types can define their own conversion logic that
// will be respected by Value.As(), as well, by implementing the
// FromTerraform5Value method for that type. The FromTerraform5Value method
// accepts a Value as an argument and returns an error. The Value passed in
// will be the same Value that Value.As() was called on. The recommended
// implementation of the FromTerraform5Value method is to call Value.As() on
// the passed Value, converting it into one of the built-in types above, and
// then performing whatever type casting or conversion logic is required to
// assign the data to the provider-supplied type.
package tftypes
