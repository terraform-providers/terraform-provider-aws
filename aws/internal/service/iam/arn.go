package iam

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
)

const (
	ARNSeparator = "/"
	ARNService   = "iam"

	InstanceProfileResourcePrefix = "instance-profile"
)

// InstanceProfileARNToName converts Amazon Resource Name (ARN) to Name.
func InstanceProfileARNToName(inputARN string) (string, error) {
	parsedARN, err := arn.Parse(inputARN)

	if err != nil {
		return "", fmt.Errorf("error parsing ARN (%s): %w", inputARN, err)
	}

	if actual, expected := parsedARN.Service, ARNService; actual != expected {
		return "", fmt.Errorf("expected service %s in ARN (%s), got: %s", expected, inputARN, actual)
	}

	resourceParts := strings.Split(parsedARN.Resource, ARNSeparator)

	if actual, expected := len(resourceParts), 2; actual < expected {
		return "", fmt.Errorf("expected at least %d resource parts in ARN (%s), got: %d", expected, inputARN, actual)
	}

	if actual, expected := resourceParts[0], InstanceProfileResourcePrefix; actual != expected {
		return "", fmt.Errorf("expected resource prefix %s in ARN (%s), got: %s", expected, inputARN, actual)
	}

	return resourceParts[len(resourceParts)-1], nil
}
