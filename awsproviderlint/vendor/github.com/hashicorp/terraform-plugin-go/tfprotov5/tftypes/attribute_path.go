package tftypes

import (
	"errors"
	"fmt"
)

var (
	// ErrNotAttributePathStepper is returned when a type that doesn't full
	// the AttributePathStepper interface is passed to WalkAttributePath.
	ErrNotAttributePathStepper = errors.New("doesn't fill tftypes.AttributePathStepper interface")

	// ErrInvalidStep is returned when an AttributePath has the wrong kind
	// of AttributePathStep for the type that WalkAttributePath is
	// operating on.
	ErrInvalidStep = errors.New("step cannot be applied to this value")
)

// AttributePath is a type that can point to a specific value within an
// aggregate Terraform value. It consists of steps, each identifying one
// element or attribute of the current value, and making that the current
// value. This allows referring to arbitrarily precise values.
type AttributePath struct {
	// Steps are the steps that must be followed from the root of the value
	// to obtain the value being indicated.
	Steps []AttributePathStep
}

// NewErrorf returns an error associated with the value indicated by `a`. This
// is equivalent to calling a.NewError(fmt.Errorf(f, args...)).
func (a AttributePath) NewErrorf(f string, args ...interface{}) error {
	return attributePathError{
		error: fmt.Errorf(f, args...),
		path:  a,
	}
}

// NewError returns an error that associates `err` with the value indicated by
// `a`.
func (a AttributePath) NewError(err error) error {
	return attributePathError{
		error: err,
		path:  a,
	}
}

// WithAttributeName adds an AttributeName step to `a`, using `name` as the
// attribute's name.
func (a *AttributePath) WithAttributeName(name string) {
	a.Steps = append(a.Steps, AttributeName(name))
}

// WithElementKeyString adds an ElementKeyString step to `a`, using `key` as
// the element's key.
func (a *AttributePath) WithElementKeyString(key string) {
	a.Steps = append(a.Steps, ElementKeyString(key))
}

// WithElementKeyInt adds an ElementKeyInt step to `a`, using `key` as the
// element's key.
func (a *AttributePath) WithElementKeyInt(key int64) {
	a.Steps = append(a.Steps, ElementKeyInt(key))
}

// WithElementKeyValue adds an ElementKeyValue to `a`, using `key` as the
// element's key.
func (a *AttributePath) WithElementKeyValue(key Value) {
	a.Steps = append(a.Steps, ElementKeyValue(key))
}

// WithoutLastStep removes the last step, whatever kind of step it was, from
// `a`.
func (a *AttributePath) WithoutLastStep() {
	a.Steps = a.Steps[:len(a.Steps)-1]
}

// AttributePathStep is an intentionally unimplementable interface that
// functions as an enum, allowing us to use different strongly-typed step types
// as a generic "step" type.
//
// An AttributePathStep is meant to indicate a single step in an AttributePath,
// indicating a specific attribute or element that is the next value in the
// path.
type AttributePathStep interface {
	unfulfillable() // make this interface fillable only by this package
}

var (
	_ AttributePathStep = AttributeName("")
	_ AttributePathStep = ElementKeyString("")
	_ AttributePathStep = ElementKeyInt(0)
)

// AttributeName is an AttributePathStep implementation that indicates the next
// step in the AttributePath is to select an attribute. The value of the
// AttributeName is the name of the attribute to be selected.
type AttributeName string

func (a AttributeName) unfulfillable() {}

// ElementKeyString is an AttributePathStep implementation that indicates the
// next step in the AttributePath is to select an element using a string key.
// The value of the ElementKeyString is the key of the element to select.
type ElementKeyString string

func (e ElementKeyString) unfulfillable() {}

// ElementKeyInt is an AttributePathStep implementation that indicates the next
// step in the AttributePath is to select an element using an int64 key. The
// value of the ElementKeyInt is the key of the element to select.
type ElementKeyInt int64

func (e ElementKeyInt) unfulfillable() {}

// ElementKeyValue is an AttributePathStep implementation that indicates the
// next step in the AttributePath is to select an element using the element
// itself as a key. The value of the ElementKeyValue is the key of the element
// to select.
type ElementKeyValue Value

func (e ElementKeyValue) unfulfillable() {}

// AttributePathStepper is an interface that types can implement to make them
// traversable by WalkAttributePath, allowing providers to retrieve the
// specific value an AttributePath is pointing to.
type AttributePathStepper interface {
	// Return the attribute or element the AttributePathStep is referring
	// to, or an error if the AttributePathStep is referring to an
	// attribute or element that doesn't exist.
	ApplyTerraform5AttributePathStep(AttributePathStep) (interface{}, error)
}

// WalkAttributePath will return the value that `path` is pointing to, using
// `in` as the root. If an error is returned, the AttributePath returned will
// indicate the steps that remained to be applied when the error was
// encountered.
//
// map[string]interface{} and []interface{} types have built-in support. Other
// types need to use the AttributePathStepper interface to tell
// WalkAttributePath how to traverse themselves.
func WalkAttributePath(in interface{}, path AttributePath) (interface{}, AttributePath, error) {
	if len(path.Steps) < 1 {
		return in, path, nil
	}
	stepper, ok := in.(AttributePathStepper)
	if !ok {
		stepper, ok = builtinAttributePathStepper(in)
		if !ok {
			return in, path, ErrNotAttributePathStepper
		}
	}
	next, err := stepper.ApplyTerraform5AttributePathStep(path.Steps[0])
	if err != nil {
		return in, path, err
	}
	path.Steps = path.Steps[1:]
	return WalkAttributePath(next, path)
}

func builtinAttributePathStepper(in interface{}) (AttributePathStepper, bool) {
	switch v := in.(type) {
	case map[string]interface{}:
		return mapStringInterfaceAttributePathStepper(v), true
	case []interface{}:
		return interfaceSliceAttributePathStepper(v), true
	default:
		return nil, false
	}
}

type mapStringInterfaceAttributePathStepper map[string]interface{}

func (m mapStringInterfaceAttributePathStepper) ApplyTerraform5AttributePathStep(step AttributePathStep) (interface{}, error) {
	_, isAttributeName := step.(AttributeName)
	_, isElementKeyString := step.(ElementKeyString)
	if !isAttributeName && !isElementKeyString {
		return nil, ErrInvalidStep
	}
	var stepValue string
	if isAttributeName {
		stepValue = string(step.(AttributeName))
	}
	if isElementKeyString {
		stepValue = string(step.(ElementKeyString))
	}
	v, ok := m[stepValue]
	if !ok {
		return nil, ErrInvalidStep
	}
	return v, nil
}

type interfaceSliceAttributePathStepper []interface{}

func (i interfaceSliceAttributePathStepper) ApplyTerraform5AttributePathStep(step AttributePathStep) (interface{}, error) {
	eki, isElementKeyInt := step.(ElementKeyInt)
	if !isElementKeyInt {
		return nil, ErrInvalidStep
	}
	// slices can only have items up to the max value of int
	// but we get ElementKeyInt as an int64
	// we keep ElementKeyInt as an int64 and cast the length of the slice
	// to int64 here because if ElementKeyInt is greater than the max value
	// of int, we will always (correctly) error out here. This lets us
	// confidently cast ElementKeyInt to an int below, knowing we're not
	// truncating data
	if int64(eki) >= int64(len(i)) {
		return nil, ErrInvalidStep
	}
	return i[int(eki)], nil
}
