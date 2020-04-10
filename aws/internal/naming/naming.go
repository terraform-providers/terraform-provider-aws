package naming

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var resourceUniqueIDSuffixRegexpPattern = fmt.Sprintf("\\d{%d}$", resource.UniqueIDSuffixLength)
var resourceUniqueIDSuffixRegexp = regexp.MustCompile(resourceUniqueIDSuffixRegexpPattern)

var resourceUniqueIDRegexpPattern = resourcePrefixedUniqueIDRegexpPattern(resource.UniqueIdPrefix)
var resourceUniqueIDRegexp = regexp.MustCompile(resourceUniqueIDRegexpPattern)

// Generate returns in order the name if non-empty, a prefix generated name if non-empty, or fully generated name prefixed with terraform-
func Generate(name string, namePrefix string) string {
	if name != "" {
		return name
	}

	if namePrefix != "" {
		return resource.PrefixedUniqueId(namePrefix)
	}

	return resource.UniqueId()
}

// HasResourceUniqueIdPrefix returns true if the string has the built-in unique ID prefix
func HasResourceUniqueIdPrefix(s string) bool {
	return strings.HasPrefix(s, resource.UniqueIdPrefix)
}

// HasResourceUniqueIdSuffix returns true if the string has the built-in unique ID suffix
func HasResourceUniqueIdSuffix(s string) bool {
	return resourceUniqueIDSuffixRegexp.MatchString(s)
}

// NamePrefixFromName returns a name prefix if the string matches prefix criteria
//
// The input to this function must be strictly the "name" and not any
// additional information such as a full Amazon Resource Name (ARN). The output
// is suitable for custom resource Importer State functions after nil checking
// to ensure differences are not reported with ImportStateVerify testing, e.g.
//
// if namePrefix := naming.NamePrefixFromName(d.Id()); namePrefix != nil {
//   d.Set("name_prefix", namePrefix)
// }
func NamePrefixFromName(name string) *string {
	if !HasResourceUniqueIdSuffix(name) {
		return nil
	}

	// If the name begins with terraform-, then the name may have been fully
	// generated (e.g. omitting both name and name_prefix arguments)
	if HasResourceUniqueIdPrefix(name) {
		return nil
	}

	namePrefixIndex := len(name) - resource.UniqueIDSuffixLength

	if namePrefixIndex <= 0 {
		return nil
	}

	namePrefix := name[:namePrefixIndex]

	return &namePrefix
}

// TestCheckResourceAttrNameFromPrefix verifies that the state attribute value matches name generated from given prefix
func TestCheckResourceAttrNameFromPrefix(resourceName string, attributeName string, prefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		nameRegexpPattern := resourcePrefixedUniqueIDRegexpPattern(prefix)
		attributeMatch, err := regexp.Compile(nameRegexpPattern)

		if err != nil {
			return fmt.Errorf("Unable to compile name regexp (%s): %s", nameRegexpPattern, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// TestCheckResourceAttrNameGenerated verifies that the state attribute value matches name automatically generated without prefix
func TestCheckResourceAttrNameGenerated(resourceName string, attributeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestMatchResourceAttr(resourceName, attributeName, resourceUniqueIDRegexp)(s)
	}
}

func resourcePrefixedUniqueIDRegexpPattern(prefix string) string {
	return fmt.Sprintf("^%s%s", prefix, resourceUniqueIDSuffixRegexpPattern)
}
