package naming

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Generate returns in order the name if non-empty, a prefix generated name if non-empty, or fully generated name prefixed with terraform-
func Generate(name string, namePrefix string) string {
	return GenerateWithSuffix(name, namePrefix, "")
}

// GenerateWithSuffix returns in order the name if non-empty, a prefix generated name if non-empty, or fully generated name prefixed with "terraform-".
// In the latter two cases, any suffix is appended to the generated name
func GenerateWithSuffix(name string, namePrefix string, nameSuffix string) string {
	if name != "" {
		return name
	}

	if namePrefix != "" {
		return resource.PrefixedUniqueId(namePrefix) + nameSuffix
	}

	return resource.UniqueId() + nameSuffix
}

// HasResourceUniqueIdSuffix returns true if the string has the built-in unique ID suffix
func HasResourceUniqueIdSuffix(s string) bool {
	return HasResourceUniqueIdPlusAdditionalSuffix(s, "")
}

// HasResourceUniqueIdPlusAdditionalSuffix returns true if the string has the built-in unique ID suffix plus an additional suffix
func HasResourceUniqueIdPlusAdditionalSuffix(s string, additionalSuffix string) bool {
	return resourceUniqueIDPlusAdditionalSuffixRegexp(additionalSuffix).MatchString(s)
}

// NamePrefixFromName returns a name prefix if the string matches prefix criteria
//
// The input to this function must be strictly the "name" and not any
// additional information such as a full Amazon Resource Name (ARN).
//
// An expected usage might be:
//
//   d.Set("name_prefix", naming.NamePrefixFromName(d.Id()))
//
func NamePrefixFromName(name string) *string {
	return NamePrefixFromNameWithSuffix(name, "")
}

func NamePrefixFromNameWithSuffix(name, nameSuffix string) *string {
	if !HasResourceUniqueIdPlusAdditionalSuffix(name, nameSuffix) {
		return nil
	}

	namePrefixIndex := len(name) - resource.UniqueIDSuffixLength - len(nameSuffix)

	if namePrefixIndex <= 0 {
		return nil
	}

	namePrefix := name[:namePrefixIndex]

	return &namePrefix
}

// TestCheckResourceAttrNameFromPrefix verifies that the state attribute value matches name generated from given prefix
func TestCheckResourceAttrNameFromPrefix(resourceName string, attributeName string, prefix string) resource.TestCheckFunc {
	return TestCheckResourceAttrNameWithSuffixFromPrefix(resourceName, attributeName, prefix, "")
}

// TestCheckResourceAttrNameWithSuffixFromPrefix verifies that the state attribute value matches name with suffix generated from given prefix
func TestCheckResourceAttrNameWithSuffixFromPrefix(resourceName string, attributeName string, prefix string, suffix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		nameRegexpPattern := resourcePrefixedUniqueIDPlusAdditionalSuffixRegexpPattern(prefix, suffix)
		attributeMatch, err := regexp.Compile(nameRegexpPattern)

		if err != nil {
			return fmt.Errorf("Unable to compile name regexp (%s): %w", nameRegexpPattern, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// TestCheckResourceAttrNameGenerated verifies that the state attribute value matches name automatically generated without prefix
func TestCheckResourceAttrNameGenerated(resourceName string, attributeName string) resource.TestCheckFunc {
	return TestCheckResourceAttrNameWithSuffixGenerated(resourceName, attributeName, "")
}

// TestCheckResourceAttrNameWithSuffixGenerated verifies that the state attribute value matches name with suffix automatically generated without prefix
func TestCheckResourceAttrNameWithSuffixGenerated(resourceName string, attributeName string, suffix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestMatchResourceAttr(resourceName, attributeName, resourceUniqueIDPrefixPlusAdditionalSuffixRegexp(suffix))(s)
	}
}

// Regexp pattern for "<26 lowercase hex digits><additional suffix><end-of-string>".
func resourceUniqueIDPlusAdditionalSuffixRegexpPattern(additionalSuffix string) string {
	return fmt.Sprintf("[[:xdigit:]]{%d}%s$", resource.UniqueIDSuffixLength, additionalSuffix)
}

// Regexp for "<26 lowercase hex digits><additional suffix><end-of-string>".
func resourceUniqueIDPlusAdditionalSuffixRegexp(additionalSuffix string) *regexp.Regexp {
	return regexp.MustCompile(resourceUniqueIDPlusAdditionalSuffixRegexpPattern(additionalSuffix))
}

// Regexp pattern for "<start-of-string><prefix><26 lowercase hex digits><additional suffix><end-of-string>".
func resourcePrefixedUniqueIDPlusAdditionalSuffixRegexpPattern(prefix string, additionalSuffix string) string {
	return fmt.Sprintf("^%s%s", prefix, resourceUniqueIDPlusAdditionalSuffixRegexpPattern(additionalSuffix))
}

// Regexp pattern for "<start-of-string>terraform-<26 lowercase hex digits><additional suffix><end-of-string>".
func resourceUniqueIDPrefixPlusAdditionalSuffixRegexpPattern(additionalSuffix string) string {
	return resourcePrefixedUniqueIDPlusAdditionalSuffixRegexpPattern(resource.UniqueIdPrefix, additionalSuffix)
}

// Regexp for "<start-of-string>terraform-<26 lowercase hex digits><additional suffix><end-of-string>".
func resourceUniqueIDPrefixPlusAdditionalSuffixRegexp(additionalSuffix string) *regexp.Regexp {
	return regexp.MustCompile(resourceUniqueIDPrefixPlusAdditionalSuffixRegexpPattern(additionalSuffix))
}
