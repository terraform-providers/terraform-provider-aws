//go:generate go run -tags generate generators/servicetags/main.go
//go:generate go run -tags generate generators/listtags/main.go
//go:generate go run -tags generate generators/gettag/main.go
//go:generate go run -tags generate generators/createtags/main.go
//go:generate go run -tags generate generators/updatetags/main.go

package keyvaluetags

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
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
type KeyValueTags map[string]*string

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

// KeyExists returns true if a tag key exists.
// If the key is not found, returns nil.
// Use KeyExists to determine if key is present.
func (tags KeyValueTags) KeyExists(key string) bool {
	if _, ok := tags[key]; ok {
		return true
	}

	return false
}

// KeyValue returns a tag key value.
// If the key is not found, returns nil.
// Use KeyExists to determine if key is present.
func (tags KeyValueTags) KeyValue(key string) *string {
	if v, ok := tags[key]; ok {
		return v
	}

	return nil
}

// Keys returns tag keys.
func (tags KeyValueTags) Keys() []string {
	result := make([]string, 0, len(tags))

	for k := range tags {
		result = append(result, k)
	}

	return result
}

// Map returns tag keys mapped to their values.
func (tags KeyValueTags) Map() map[string]string {
	result := make(map[string]string, len(tags))

	for k, v := range tags {
		result[k] = *v
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
		if oldV, ok := tags[k]; !ok || *oldV != *newV {
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
		if v, ok := tags[key]; !ok || *v != *value {
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
		hash = hash ^ hashcode.String(fmt.Sprintf("%s-%s", k, *v))
	}

	return hash
}

// UrlEncode returns the KeyValueTags encoded as URL Query parameters.
func (tags KeyValueTags) UrlEncode() string {
	values := url.Values{}

	for k, v := range tags {
		values.Add(k, *v)
	}

	return values.Encode()
}

// New creates KeyValueTags from common Terraform Provider SDK types.
// Supports map[string]string, map[string]*string, map[string]interface{}, and []interface{}.
// When passed []interface{}, all elements are treated as keys and assigned nil values.
func New(i interface{}) KeyValueTags {
	switch value := i.(type) {
	case map[string]string:
		kvtm := make(KeyValueTags, len(value))

		for k, v := range value {
			str := v // Prevent referencing issues
			kvtm[k] = &str
		}

		return kvtm
	case map[string]*string:
		return KeyValueTags(value)
	case map[string]interface{}:
		kvtm := make(KeyValueTags, len(value))

		for k, v := range value {
			str := v.(string)
			kvtm[k] = &str
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
