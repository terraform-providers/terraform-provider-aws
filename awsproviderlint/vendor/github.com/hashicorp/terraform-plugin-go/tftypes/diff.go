package tftypes

import (
	"errors"
	"fmt"
	"math/big"
)

// ValueDiff expresses a subset of a Value that is different between two
// Values. The Path property indicates where the subset is located within the
// Value, and Value1 and Value2 indicate what the subset is in each of the
// Values. If the Value does not contain a subset at that AttributePath, its
// Value will be nil. This is distinct from a Value with a nil in it (a "null"
// value), which is present in the Value.
type ValueDiff struct {
	// The Path these different subsets are located at in the original
	// Values.
	Path *AttributePath

	// The subset of the first Value passed to Diff found at the
	// AttributePath indicated by Path.
	Value1 *Value

	// The subset of the second Value passed to Diff found at the
	// AttributePath indicated by Path.
	Value2 *Value
}

func (v ValueDiff) String() string {
	val1 := "{no value set}"
	if v.Value1 != nil {
		val1 = v.Value1.String()
	}
	val2 := "{no value set}"
	if v.Value2 != nil {
		val2 = v.Value2.String()
	}
	return fmt.Sprintf("%s: value1: %s, value2: %s",
		v.Path.String(), val1, val2)
}

// Equal returns whether two ValueDiffs should be considered equal or not.
// ValueDiffs are consisdered equal when their Path, Value1, and Value2
// properties are considered equal.
func (v ValueDiff) Equal(o ValueDiff) bool {
	if !v.Path.Equal(o.Path) {
		return false
	}
	if v.Value1 == nil && o.Value1 != nil {
		return false
	}
	if v.Value1 != nil && o.Value1 == nil {
		return false
	}
	if v.Value1 != nil && o.Value1 != nil && !v.Value1.Equal(*o.Value1) {
		return false
	}
	if v.Value2 == nil && o.Value2 != nil {
		return false
	}
	if v.Value2 != nil && o.Value2 == nil {
		return false
	}
	if v.Value2 != nil && o.Value2 != nil && !v.Value2.Equal(*o.Value2) {
		return false
	}
	return true
}

// Diff computes the differences between `val1` and `val2` and surfaces them as
// a slice of ValueDiffs. The ValueDiffs in the struct will use `val1`'s values
// as Value1 and `val2`'s values as Value2. An empty or nil slice means the two
// Values can be considered equal. Values must be the same type when passed to
// Diff; passing in Values of two different types will result in an error.
// val1.Type().Is(val2.Type()) is a safe way to check that Values can be
// compared with Diff.
func (val1 Value) Diff(val2 Value) ([]ValueDiff, error) {
	if !val1.Type().Is(val2.Type()) {
		return nil, errors.New("Can't diff values of different types")
	}
	var diffs []ValueDiff

	// make sure everything in val2 is also in val1
	err := Walk(val2, func(path *AttributePath, value2 Value) (bool, error) {
		_, _, err := WalkAttributePath(val1, path)
		if err != nil && err != ErrInvalidStep {
			return false, fmt.Errorf("Error walking %q: %w", path, err)
		} else if err == ErrInvalidStep {
			diffs = append(diffs, ValueDiff{
				Path:   path,
				Value1: nil,
				Value2: &value2,
			})
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	// make sure everything in val1 is also in val2 and also that it all matches
	err = Walk(val1, func(path *AttributePath, value1 Value) (bool, error) {
		// pull out the Value at the same path in val2
		value2I, _, err := WalkAttributePath(val2, path)
		if err != nil && err != ErrInvalidStep {
			return false, fmt.Errorf("Error walking %q: %w", path, err)
		} else if err == ErrInvalidStep {
			diffs = append(diffs, ValueDiff{
				Path:   path,
				Value1: &value1,
				Value2: nil,
			})
			return true, nil
		}

		// convert from an interface{} to a Value
		value2 := value2I.(Value)

		// if they're both unknown, no need to continue
		if !value1.IsKnown() && !value2.IsKnown() {
			return false, nil
		}

		// if val1 is unknown and val2 not, we have a diff
		// no need to continue to recurse into val1, no further to go
		if !value1.IsKnown() && value2.IsKnown() {
			diffs = append(diffs, ValueDiff{
				Path:   path,
				Value1: &value1,
				Value2: &value2,
			})
			return false, nil
		}

		// if val2 is unknown and val1 not, we have a diff
		// continue to recurse though, so we can surface the elements of val1
		// that are now "missing" as diffs
		if value1.IsKnown() && !value2.IsKnown() {
			diffs = append(diffs, ValueDiff{
				Path:   path,
				Value1: &value1,
				Value2: &value2,
			})
			return true, nil
		}

		// if they're both null, no need to continue
		if value1.IsNull() && value2.IsNull() {
			return false, nil
		}

		// if val1 is null and val2 not, we have a diff
		// no need to continue to recurse into val1, no further to go
		if value1.IsNull() && !value2.IsNull() {
			diffs = append(diffs, ValueDiff{
				Path:   path,
				Value1: &value1,
				Value2: &value2,
			})
			return false, nil
		}

		// if val2 is null and val1 not, we have a diff
		// continue to recurse though, so we can surface the elements of val1
		// that are now "missing" as diffs
		if !value1.IsNull() && value2.IsNull() {
			diffs = append(diffs, ValueDiff{
				Path:   path,
				Value1: &value1,
				Value2: &value2,
			})
			return true, nil
		}

		// we know there are known, non-null values, time to compare them
		switch {
		case value1.Type().Is(String):
			var s1, s2 string
			err := value1.As(&s1)
			if err != nil {
				return false, fmt.Errorf("Error converting %s (value1) at %q: %w", value1, path, err)
			}
			err = value2.As(&s2)
			if err != nil {
				return false, fmt.Errorf("Error converting %s (value2) at %q: %w", value2, path, err)
			}
			if s1 != s2 {
				diffs = append(diffs, ValueDiff{
					Path:   path,
					Value1: &value1,
					Value2: &value2,
				})
			}
			return false, nil
		case value1.Type().Is(Number):
			n1, n2 := big.NewFloat(0), big.NewFloat(0)
			err := value1.As(&n1)
			if err != nil {
				return false, fmt.Errorf("Error converting %q: %w", path, err)
			}
			err = value2.As(&n2)
			if err != nil {
				return false, fmt.Errorf("Error converting %q: %w", path, err)
			}
			if n1.Cmp(n2) != 0 {
				diffs = append(diffs, ValueDiff{
					Path:   path,
					Value1: &value1,
					Value2: &value2,
				})
			}
			return false, nil
		case value1.Type().Is(Bool):
			var b1, b2 bool
			err := value1.As(&b1)
			if err != nil {
				return false, fmt.Errorf("Error converting %q: %w", path, err)
			}
			err = value2.As(&b2)
			if err != nil {
				return false, fmt.Errorf("Error converting %q: %w", path, err)
			}
			if b1 != b2 {
				diffs = append(diffs, ValueDiff{
					Path:   path,
					Value1: &value1,
					Value2: &value2,
				})
			}
			return false, nil
		case value1.Type().Is(List{}), value1.Type().Is(Set{}), value1.Type().Is(Tuple{}):
			var s1, s2 []Value
			err := value1.As(&s1)
			if err != nil {
				return false, fmt.Errorf("Error converting %q: %w", path, err)
			}
			err = value2.As(&s2)
			if err != nil {
				return false, fmt.Errorf("Error converting %q: %w", path, err)
			}
			// we only care about if the lengths match for lists,
			// sets, and tuples. If any of the elements differ,
			// the recursion of the walk will find them for us.
			if len(s1) != len(s2) {
				diffs = append(diffs, ValueDiff{
					Path:   path,
					Value1: &value1,
					Value2: &value2,
				})
				return true, nil
			}
			return true, nil
		case value1.Type().Is(Map{}), value1.Type().Is(Object{}):
			m1 := map[string]Value{}
			m2 := map[string]Value{}
			err := value1.As(&m1)
			if err != nil {
				return false, fmt.Errorf("Error converting %q: %w", path, err)
			}
			err = value2.As(&m2)
			if err != nil {
				return false, fmt.Errorf("Error converting %q: %w", path, err)
			}
			// we need maps and objects to have the same exact keys
			// as each other
			if len(m1) != len(m2) {
				diffs = append(diffs, ValueDiff{
					Path:   path,
					Value1: &value1,
					Value2: &value2,
				})
				return true, nil
			}
			// if we have the same keys, we can just let recursion
			// from the walk check the sub-values match
			return true, nil
		}
		return false, fmt.Errorf("unexpected type %v in Diff at %s", value1.Type(), path)
	})
	return diffs, err
}
