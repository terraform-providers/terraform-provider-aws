// These test helpers were developed by the AWS provider team at HashiCorp.

package resource

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	sentinelIndex = "*"
)

// TestCheckTypeSetElemNestedAttrs is a TestCheckFunc that accepts a resource
// name, an attribute path, which should use the sentinel value '*' for indexing
// into a TypeSet. The function verifies that an element matches the whole value
// map.
//
// You may check for unset keys, however this will also match keys set to empty
// string. Please provide a map with at least 1 non-empty value.
//
//   map[string]string{
//	     "key1": "value",
//       "key2": "",
//   }
//
// Use this function over SDK provided TestCheckFunctions when validating a
// TypeSet where its elements are a nested object with their own attrs/values.
//
// Please note, if the provided value map is not granular enough, there exists
// the possibility you match an element you were not intending to, in the TypeSet.
// Provide a full mapping of attributes to be sure the unique element exists.
func TestCheckTypeSetElemNestedAttrs(name, attr string, values map[string]string) TestCheckFunc {
	return func(s *terraform.State) error {
		is, err := primaryInstanceState(s, name)
		if err != nil {
			return err
		}

		attrParts := strings.Split(attr, ".")
		if attrParts[len(attrParts)-1] != sentinelIndex {
			return fmt.Errorf("%q does not end with the special value %q", attr, sentinelIndex)
		}
		// account for cases where the user is trying to see if the value is unset/empty
		// there may be ambiguous scenarios where a field was deliberately unset vs set
		// to the empty string, this will match both, which may be a false positive.
		var matchCount int
		for _, v := range values {
			if v != "" {
				matchCount++
			}
		}
		if matchCount == 0 {
			return fmt.Errorf("%#v has no non-empty values", values)
		}

		if testCheckTypeSetElemNestedAttrsInState(is, attrParts, matchCount, values) {
			return nil
		}
		return fmt.Errorf("%q no TypeSet element %q, with nested attrs %#v in state: %#v", name, attr, values, is.Attributes)
	}
}

// TestMatchTypeSetElemNestedAttrs is a TestCheckFunc similar to TestCheckTypeSetElemNestedAttrs
// with the exception that it verifies that an element matches a *regexp.Regexp.
//
// You may check for unset keys, however this will also match keys set to empty
// string. Please provide a map with at least 1 non-empty value e.g.
//
//   map[string]*regexp.Regexp{
//	     "key1": regexp.MustCompile("value"),
//       "key2": regexp.MustCompile(""),
//   }
//
func TestMatchTypeSetElemNestedAttrs(name, attr string, values map[string]*regexp.Regexp) TestCheckFunc {
	return func(s *terraform.State) error {
		is, err := primaryInstanceState(s, name)
		if err != nil {
			return err
		}

		attrParts := strings.Split(attr, ".")
		if attrParts[len(attrParts)-1] != sentinelIndex {
			return fmt.Errorf("%q does not end with the special value %q", attr, sentinelIndex)
		}
		// account for cases where the user is trying to see if the value is unset/empty
		// there may be ambiguous scenarios where a field was deliberately unset vs set
		// to the empty string, this will match both, which may be a false positive.
		var matchCount int
		for _, v := range values {
			if v != nil {
				matchCount++
			}
		}
		if matchCount == 0 {
			return fmt.Errorf("%#v has no non-empty values", values)
		}

		if testCheckTypeSetElemNestedAttrsInState(is, attrParts, matchCount, values) {
			return nil
		}
		return fmt.Errorf("%q no TypeSet element %q, with the regex provided, match in state: %#v", name, attr, is.Attributes)
	}
}

// TestCheckTypeSetElemAttr is a TestCheckFunc that accepts a resource
// name, an attribute path, which should use the sentinel value '*' for indexing
// into a TypeSet. The function verifies that an element matches the provided
// value.
//
// Use this function over SDK provided TestCheckFunctions when validating a
// TypeSet where its elements are a simple value
func TestCheckTypeSetElemAttr(name, attr, value string) TestCheckFunc {
	return func(s *terraform.State) error {
		is, err := primaryInstanceState(s, name)
		if err != nil {
			return err
		}

		err = testCheckTypeSetElem(is, attr, value)
		if err != nil {
			return fmt.Errorf("%q error: %s", name, err)
		}

		return nil
	}
}

// TestCheckTypeSetElemAttrPair is a TestCheckFunc that verifies a pair of name/key
// combinations are equal where the first uses the sentinel value to index into a
// TypeSet.
//
// E.g., TestCheckTypeSetElemAttrPair("aws_autoscaling_group.bar", "availability_zones.*", "data.aws_availability_zones.available", "names.0")
// E.g., TestCheckTypeSetElemAttrPair("aws_spot_fleet_request.bar", "launch_specification.*.instance_type", "data.data.aws_ec2_instance_type_offering.available", "instance_type")
func TestCheckTypeSetElemAttrPair(nameFirst, keyFirst, nameSecond, keySecond string) TestCheckFunc {
	return func(s *terraform.State) error {
		isFirst, err := primaryInstanceState(s, nameFirst)
		if err != nil {
			return err
		}

		isSecond, err := primaryInstanceState(s, nameSecond)
		if err != nil {
			return err
		}

		vSecond, okSecond := isSecond.Attributes[keySecond]
		if !okSecond {
			return fmt.Errorf("%s: Attribute %q not set, cannot be checked against TypeSet", nameSecond, keySecond)
		}

		return testCheckTypeSetElemPair(isFirst, keyFirst, vSecond)
	}
}

func testCheckTypeSetElem(is *terraform.InstanceState, attr, value string) error {
	attrParts := strings.Split(attr, ".")
	if attrParts[len(attrParts)-1] != sentinelIndex {
		return fmt.Errorf("%q does not end with the special value %q", attr, sentinelIndex)
	}
	for stateKey, stateValue := range is.Attributes {
		if stateValue == value {
			stateKeyParts := strings.Split(stateKey, ".")
			if len(stateKeyParts) == len(attrParts) {
				for i := range attrParts {
					if attrParts[i] != stateKeyParts[i] && attrParts[i] != sentinelIndex {
						break
					}
					if i == len(attrParts)-1 {
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("no TypeSet element %q, with value %q in state: %#v", attr, value, is.Attributes)
}

func testCheckTypeSetElemPair(is *terraform.InstanceState, attr, value string) error {
	attrParts := strings.Split(attr, ".")
	for stateKey, stateValue := range is.Attributes {
		if stateValue == value {
			stateKeyParts := strings.Split(stateKey, ".")
			if len(stateKeyParts) == len(attrParts) {
				for i := range attrParts {
					if attrParts[i] != stateKeyParts[i] && attrParts[i] != sentinelIndex {
						break
					}
					if i == len(attrParts)-1 {
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("no TypeSet element %q, with value %q in state: %#v", attr, value, is.Attributes)
}

// testCheckTypeSetElemNestedAttrsInState is a helper function
// to determine if nested attributes and their values are equal to those
// in the instance state. Currently, the function accepts a "values" param of type
// map[string]string or map[string]*regexp.Regexp.
// Returns true if all attributes match, else false.
func testCheckTypeSetElemNestedAttrsInState(is *terraform.InstanceState, attrParts []string, matchCount int, values interface{}) bool {
	matches := make(map[string]int)

	for stateKey, stateValue := range is.Attributes {
		stateKeyParts := strings.Split(stateKey, ".")
		// a Set/List item with nested attrs would have a flatmap address of
		// at least length 3
		// foo.0.name = "bar"
		if len(stateKeyParts) < 3 || len(attrParts) > len(stateKeyParts) {
			continue
		}
		var pathMatch bool
		for i := range attrParts {
			if attrParts[i] != stateKeyParts[i] && attrParts[i] != sentinelIndex {
				break
			}
			if i == len(attrParts)-1 {
				pathMatch = true
			}
		}
		if !pathMatch {
			continue
		}
		id := stateKeyParts[len(attrParts)-1]
		nestedAttr := strings.Join(stateKeyParts[len(attrParts):], ".")

		var match bool
		switch t := values.(type) {
		case map[string]string:
			if v, keyExists := t[nestedAttr]; keyExists && v == stateValue {
				match = true
			}
		case map[string]*regexp.Regexp:
			if v, keyExists := t[nestedAttr]; keyExists && v != nil && v.MatchString(stateValue) {
				match = true
			}
		}
		if match {
			matches[id] = matches[id] + 1
			if matches[id] == matchCount {
				return true
			}
		}
	}
	return false
}
