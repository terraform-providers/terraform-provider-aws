package validation

import "fmt"

// ValidateListUniqueStrings is a ValidateFunc that ensures a list has no
// duplicate items in it. It's useful for when a list is needed over a set
// because order matters, yet the items still need to be unique.
//
// Deprecated: use ListOfUniqueStrings
func ValidateListUniqueStrings(i interface{}, k string) (warnings []string, errors []error) {
	return ListOfUniqueStrings(i, k)
}

// ListOfUniqueStrings is a ValidateFunc that ensures a list has no
// duplicate items in it. It's useful for when a list is needed over a set
// because order matters, yet the items still need to be unique.
func ListOfUniqueStrings(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.([]interface{})
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be List", k))
		return warnings, errors
	}

	for _, e := range v {
		if _, eok := e.(string); !eok {
			errors = append(errors, fmt.Errorf("expected %q to only contain string elements, found :%v", k, e))
			return warnings, errors
		}
	}

	for n1, i1 := range v {
		for n2, i2 := range v {
			if i1.(string) == i2.(string) && n1 != n2 {
				errors = append(errors, fmt.Errorf("expected %q to not have duplicates: found 2 or more of %v", k, i1))
				return warnings, errors
			}
		}
	}

	return warnings, errors
}
