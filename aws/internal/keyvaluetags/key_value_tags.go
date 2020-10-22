//go:generate go run -tags generate generators/servicetags/main.go
//go:generate go run -tags generate generators/listtags/main.go
//go:generate go run -tags generate generators/gettag/main.go
//go:generate go run -tags generate generators/createtags/main.go
//go:generate go run -tags generate generators/updatetags/main.go

package keyvaluetags

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
)

const (
	AwsTagKeyPrefix              = `aws:`
	ElasticbeanstalkTagKeyPrefix = `elasticbeanstalk:`
	NameTagKey                   = `Name`
	RdsTagKeyPrefix              = `rds:`
)

// IgnoreConfig contains various options for removing resource tags.
type IgnoreConfig struct {
	Keys        KeyValueTags
	KeyPrefixes KeyValueTags
}

// KeyValueTags is a standard implementation for AWS key-value resource tags.
// The AWS Go SDK is split into multiple service packages, each service with
// its own Go struct type representing a resource tag. To standardize logic
// across all these Go types, we convert them into this Go type.
type KeyValueTags map[string]*TagData

// IgnoreAws returns non-AWS tag keys.
func (tags KeyValueTags) IgnoreAws() KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		if !strings.HasPrefix(k, AwsTagKeyPrefix) {
			result[k] = v
		}
	}

	return result
}

// IgnoreConfig returns any tags not removed by a given configuration.
func (tags KeyValueTags) IgnoreConfig(config *IgnoreConfig) KeyValueTags {
	if config == nil {
		return tags
	}

	result := tags.IgnorePrefixes(config.KeyPrefixes)
	result = result.Ignore(config.Keys)

	return result
}

// IgnoreElasticbeanstalk returns non-AWS and non-Elasticbeanstalk tag keys.
func (tags KeyValueTags) IgnoreElasticbeanstalk() KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		if strings.HasPrefix(k, AwsTagKeyPrefix) {
			continue
		}

		if strings.HasPrefix(k, ElasticbeanstalkTagKeyPrefix) {
			continue
		}

		if k == NameTagKey {
			continue
		}

		result[k] = v
	}

	return result
}

// IgnorePrefixes returns non-matching tag key prefixes.
func (tags KeyValueTags) IgnorePrefixes(ignoreTagPrefixes KeyValueTags) KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		var ignore bool

		for ignoreTagPrefix := range ignoreTagPrefixes {
			if strings.HasPrefix(k, ignoreTagPrefix) {
				ignore = true
				break
			}
		}

		if ignore {
			continue
		}

		result[k] = v
	}

	return result
}

// IgnoreRDS returns non-AWS and non-RDS tag keys.
func (tags KeyValueTags) IgnoreRds() KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		if strings.HasPrefix(k, AwsTagKeyPrefix) {
			continue
		}

		if strings.HasPrefix(k, RdsTagKeyPrefix) {
			continue
		}

		result[k] = v
	}

	return result
}

// Ignore returns non-matching tag keys.
func (tags KeyValueTags) Ignore(ignoreTags KeyValueTags) KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		if _, ok := ignoreTags[k]; ok {
			continue
		}

		result[k] = v
	}

	return result
}

// KeyAdditionalBoolValue returns the boolean value of an additional tag field.
// If the key or additional field is not found, returns nil.
func (tags KeyValueTags) KeyAdditionalBoolValue(key string, fieldName string) *bool {
	tag, ok := tags[key]

	if !ok || tag == nil || tag.AdditionalBoolFields == nil {
		return nil
	}

	if v, ok := tag.AdditionalBoolFields[fieldName]; ok {
		return v
	}

	return nil
}

// KeyAdditionalStringValue returns the string value of an additional tag field.
// If the key or additional field is not found, returns nil.
func (tags KeyValueTags) KeyAdditionalStringValue(key string, fieldName string) *string {
	tag, ok := tags[key]

	if !ok || tag == nil || tag.AdditionalStringFields == nil {
		return nil
	}

	if v, ok := tag.AdditionalStringFields[fieldName]; ok {
		return v
	}

	return nil
}

// KeyExists returns true if a tag key exists.
// If the key is not found, returns nil.
func (tags KeyValueTags) KeyExists(key string) bool {
	if _, ok := tags[key]; ok {
		return true
	}

	return false
}

// KeyTagData returns all tag key data.
// If the key is not found, returns nil.
// Use KeyExists to determine if key is present.
func (tags KeyValueTags) KeyTagData(key string) *TagData {
	if v, ok := tags[key]; ok {
		return v
	}

	return nil
}

// KeyValue returns a tag key value.
// If the key is not found, returns nil.
// Use KeyExists to determine if key is present.
func (tags KeyValueTags) KeyValue(key string) *string {
	v, ok := tags[key]

	if !ok || v == nil {
		return nil
	}

	return v.Value
}

// Keys returns tag keys.
func (tags KeyValueTags) Keys() []string {
	result := make([]string, 0, len(tags))

	for k := range tags {
		result = append(result, k)
	}

	return result
}

// ListofMap returns a list of flattened tags.
// Compatible with setting Terraform state for strongly typed configuration blocks.
func (tags KeyValueTags) ListofMap() []map[string]interface{} {
	result := make([]map[string]interface{}, len(tags))

	for k, v := range tags {
		m := map[string]interface{}{
			"key":   k,
			"value": "",
		}

		if v == nil {
			result = append(result, m)
			continue
		}

		if v.Value != nil {
			m["value"] = *v.Value
		}

		for k, v := range v.AdditionalBoolFields {
			m[ToSnakeCase(k)] = false

			if v != nil {
				m[ToSnakeCase(k)] = *v
			}
		}

		for k, v := range v.AdditionalStringFields {
			m[ToSnakeCase(k)] = ""

			if v != nil {
				m[ToSnakeCase(k)] = *v
			}
		}

		result = append(result, m)
	}

	return result
}

// Map returns tag keys mapped to their values.
func (tags KeyValueTags) Map() map[string]string {
	result := make(map[string]string, len(tags))

	for k, v := range tags {
		if v == nil || v.Value == nil {
			result[k] = ""
			continue
		}

		result[k] = *v.Value
	}

	return result
}

// Merge adds missing and updates existing tags.
func (tags KeyValueTags) Merge(mergeTags KeyValueTags) KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		result[k] = v
	}

	for k, v := range mergeTags {
		result[k] = v
	}

	return result
}

// Only returns matching tag keys.
func (tags KeyValueTags) Only(onlyTags KeyValueTags) KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		if _, ok := onlyTags[k]; !ok {
			continue
		}

		result[k] = v
	}

	return result
}

// Removed returns tags removed.
func (tags KeyValueTags) Removed(newTags KeyValueTags) KeyValueTags {
	result := make(KeyValueTags)

	for k, v := range tags {
		if _, ok := newTags[k]; !ok {
			result[k] = v
		}
	}

	return result
}

// Updated returns tags added and updated.
func (tags KeyValueTags) Updated(newTags KeyValueTags) KeyValueTags {
	result := make(KeyValueTags)

	for k, newV := range newTags {
		if oldV, ok := tags[k]; !ok || !oldV.Equal(newV) {
			result[k] = newV
		}
	}

	return result
}

// Chunks returns a slice of KeyValueTags, each of the specified size.
func (tags KeyValueTags) Chunks(size int) []KeyValueTags {
	result := []KeyValueTags{}

	i := 0
	var chunk KeyValueTags
	for k, v := range tags {
		if i%size == 0 {
			chunk = make(KeyValueTags)
			result = append(result, chunk)
		}

		chunk[k] = v

		i++
	}

	return result
}

// ContainsAll returns whether or not all the target tags are contained.
func (tags KeyValueTags) ContainsAll(target KeyValueTags) bool {
	for key, value := range target {
		if v, ok := tags[key]; !ok || !v.Equal(value) {
			return false
		}
	}

	return true
}

// Hash returns a stable hash value.
// The returned value may be negative (i.e. not suitable for a 'Set' function).
func (tags KeyValueTags) Hash() int {
	hash := 0

	for k, v := range tags {
		if v == nil || v.Value == nil {
			hash = hash ^ hashcode.String(k)
			continue
		}

		hash = hash ^ hashcode.String(fmt.Sprintf("%s-%s", k, *v.Value))
	}

	return hash
}

// String returns the default string representation of the KeyValueTags.
func (tags KeyValueTags) String() string {
	var builder strings.Builder

	keys := tags.Keys()
	sort.Strings(keys)

	builder.WriteString("map[")
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(" ")
		}
		fmt.Fprintf(&builder, "%s:%s", k, tags[k].String())
	}
	builder.WriteString("]")

	return builder.String()
}

// UrlEncode returns the KeyValueTags encoded as URL Query parameters.
func (tags KeyValueTags) UrlEncode() string {
	values := url.Values{}

	for k, v := range tags {
		if v == nil || v.Value == nil {
			continue
		}

		values.Add(k, *v.Value)
	}

	return values.Encode()
}

// New creates KeyValueTags from common Terraform Provider SDK types.
// Supports map[string]string, map[string]*string, map[string]interface{}, and []interface{}.
// When passed []interface{}, all elements are treated as keys and assigned nil values.
func New(i interface{}) KeyValueTags {
	switch value := i.(type) {
	case map[string]*TagData:
		kvtm := make(KeyValueTags, len(value))

		for k, v := range value {
			tagData := v
			kvtm[k] = tagData
		}

		return kvtm
	case map[string]string:
		kvtm := make(KeyValueTags, len(value))

		for k, v := range value {
			str := v // Prevent referencing issues
			kvtm[k] = &TagData{Value: &str}
		}

		return kvtm
	case map[string]*string:
		kvtm := make(KeyValueTags, len(value))

		for k, v := range value {
			strPtr := v

			if strPtr == nil {
				kvtm[k] = nil
				continue
			}

			kvtm[k] = &TagData{Value: strPtr}
		}

		return kvtm
	case map[string]interface{}:
		kvtm := make(KeyValueTags, len(value))

		for k, v := range value {
			str := v.(string)
			kvtm[k] = &TagData{Value: &str}
		}

		return kvtm
	case []string:
		kvtm := make(KeyValueTags, len(value))

		for _, v := range value {
			kvtm[v] = nil
		}

		return kvtm
	case []interface{}:
		kvtm := make(KeyValueTags, len(value))

		for _, v := range value {
			kvtm[v.(string)] = nil
		}

		return kvtm
	default:
		return make(KeyValueTags)
	}
}

// TagData represents the data associated with a resource tag key.
// Almost exclusively for AWS services, this is just a tag value,
// however there are services that attach additional data to tags.
// An example is autoscaling with the PropagateAtLaunch field.
type TagData struct {
	// Additional boolean field names and values associated with this tag.
	// Each service is responsible for properly handling this data.
	AdditionalBoolFields map[string]*bool

	// Additional string field names and values associated with this tag.
	// Each service is responsible for properly handling this data.
	AdditionalStringFields map[string]*string

	// Tag value.
	Value *string
}

func (td *TagData) Equal(other *TagData) bool {
	if td == nil && other == nil {
		return true
	}

	if td == nil || other == nil {
		return false
	}

	if !reflect.DeepEqual(td.AdditionalBoolFields, other.AdditionalBoolFields) {
		return false
	}

	if !reflect.DeepEqual(td.AdditionalStringFields, other.AdditionalStringFields) {
		return false
	}

	if !reflect.DeepEqual(td.Value, other.Value) {
		return false
	}

	return true
}

func (td *TagData) String() string {
	if td == nil {
		return ""
	}

	var fields []string

	if len(td.AdditionalBoolFields) > 0 {
		var additionalBoolFields []string

		for k, v := range td.AdditionalBoolFields {
			additionalBoolField := fmt.Sprintf("%s:", k)

			if v != nil {
				additionalBoolField += fmt.Sprintf("%t", *v)
			}

			additionalBoolFields = append(additionalBoolFields, additionalBoolField)
		}

		fields = append(fields, fmt.Sprintf("AdditionalBoolFields: map[%s]", strings.Join(additionalBoolFields, " ")))
	}

	if len(td.AdditionalStringFields) > 0 {
		var additionalStringFields []string

		for k, v := range td.AdditionalStringFields {
			additionalStringField := fmt.Sprintf("%s:", k)

			if v != nil {
				additionalStringField += *v
			}

			additionalStringFields = append(additionalStringFields, additionalStringField)
		}

		fields = append(fields, fmt.Sprintf("AdditionalStringFields: map[%s]", strings.Join(additionalStringFields, " ")))
	}

	if td.Value != nil {
		fields = append(fields, fmt.Sprintf("Value: %s", *td.Value))
	}

	return fmt.Sprintf("TagData{%s}", strings.Join(fields, ", "))
}

// ToSnakeCase converts a string to snake case.
//
// For example, AWS Go SDK field names are in PascalCase,
// while Terraform schema attribute names are in snake_case.
func ToSnakeCase(str string) string {
	result := regexp.MustCompile("(.)([A-Z][a-z]+)").ReplaceAllString(str, "${1}_${2}")
	result = regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(result, "${1}_${2}")
	return strings.ToLower(result)
}
