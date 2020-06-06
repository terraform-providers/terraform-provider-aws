package equivalency

import (
	"bytes"
	"encoding/json"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/aws/aws-sdk-go/service/batch"
)

type containerProperties batch.ContainerProperties

func (cp *containerProperties) Reduce() error {
	// Deal with Environment objects which may be re-ordered in the API
	sort.Slice(cp.Environment, func(i, j int) bool {
		return aws.StringValue(cp.Environment[i].Name) < aws.StringValue(cp.Environment[j].Name)
	})

	// Prevent difference of API response that adds an empty array when not configured during the request
	if len(cp.Command) == 0 {
		cp.Command = nil
	}

	// Prevent difference of API response that adds an empty array when not configured during the request
	if len(cp.Environment) == 0 {
		cp.Environment = nil
	}

	// Prevent difference of API response that adds an empty array when not configured during the request
	if len(cp.MountPoints) == 0 {
		cp.MountPoints = nil
	}

	// Prevent difference of API response that adds an empty array when not configured during the request
	if len(cp.ResourceRequirements) == 0 {
		cp.ResourceRequirements = nil
	}

	// Prevent difference of API response that adds an empty array when not configured during the request
	if len(cp.Ulimits) == 0 {
		cp.Ulimits = nil
	}

	// Prevent difference of API response that adds an empty array when not configured during the request
	if len(cp.Volumes) == 0 {
		cp.Volumes = nil
	}

	return nil
}

// EquivalentBatchContainerPropertiesJSON determines equality between two Batch ContainerProperties JSON strings
func EquivalentBatchContainerPropertiesJSON(str1, str2 string) (bool, error) {
	if str1 == "" {
		str1 = "{}"
	}

	if str2 == "" {
		str2 = "{}"
	}

	var cp1, cp2 containerProperties

	if err := json.Unmarshal([]byte(str1), &cp1); err != nil {
		return false, err
	}

	if err := cp1.Reduce(); err != nil {
		return false, err
	}

	canonicalJson1, err := jsonutil.BuildJSON(cp1)

	if err != nil {
		return false, err
	}

	if err := json.Unmarshal([]byte(str2), &cp2); err != nil {
		return false, err
	}

	if err := cp2.Reduce(); err != nil {
		return false, err
	}

	canonicalJson2, err := jsonutil.BuildJSON(cp2)

	if err != nil {
		return false, err
	}

	equal := bytes.Equal(canonicalJson1, canonicalJson2)

	if !equal {
		log.Printf("[DEBUG] Canonical Batch Container Properties JSON are not equal.\nFirst: %s\nSecond: %s\n", canonicalJson1, canonicalJson2)
	}

	return equal, nil
}
