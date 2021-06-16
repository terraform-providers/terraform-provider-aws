package aws

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"regexp"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

// Base64Encode encodes data if the input isn't already encoded using base64.StdEncoding.EncodeToString.
// If the input is already base64 encoded, return the original input unchanged.
func base64Encode(data []byte) string {
	// Check whether the data is already Base64 encoded; don't double-encode
	if isBase64Encoded(data) {
		return string(data)
	}
	// data has not been encoded encode and return
	return base64.StdEncoding.EncodeToString(data)
}

func isBase64Encoded(data []byte) bool {
	_, err := base64.StdEncoding.DecodeString(string(data))
	return err == nil
}

func looksLikeJsonString(s interface{}) bool {
	return regexp.MustCompile(`^\s*{`).MatchString(s.(string))
}

func jsonBytesEqual(b1, b2 []byte) bool {
	var o1 interface{}
	if err := json.Unmarshal(b1, &o1); err != nil {
		return false
	}

	var o2 interface{}
	if err := json.Unmarshal(b2, &o2); err != nil {
		return false
	}

	return reflect.DeepEqual(o1, o2)
}

func isResourceNotFoundError(err error) bool {
	return tfresource.NotFound(err)
}

func isResourceTimeoutError(err error) bool {
	return tfresource.TimedOut(err)
}

func appendUniqueString(slice []string, elem string) []string {
	for _, e := range slice {
		if e == elem {
			return slice
		}
	}
	return append(slice, elem)
}

func StringSlicesEqualIgnoreOrder(s1, s2 []*string) bool {
	if len(s1) != len(s2) {
		return false
	}

	v1 := aws.StringValueSlice(s1)
	v2 := aws.StringValueSlice(s2)

	sort.Strings(v1)
	sort.Strings(v2)

	return reflect.DeepEqual(v1, v2)
}

func StringSlicesEqual(s1, s2 []*string) bool {
	if len(s1) != len(s2) {
		return false
	}

	v1 := aws.StringValueSlice(s1)
	v2 := aws.StringValueSlice(s2)

	return reflect.DeepEqual(v1, v2)
}
